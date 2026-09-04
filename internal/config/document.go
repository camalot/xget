package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

// GlobalSection is the section name used for global configuration values.
const GlobalSection = "global"

// ErrKeyNotSet is returned when a requested key is absent from the document.
var ErrKeyNotSet = errors.New("key is not set")

// ValueKind describes how a configuration value is stored and parsed.
type ValueKind int

const (
	KindBool ValueKind = iota
	KindString
	KindStringSlice
)

// Format identifies the on-disk serialization of a config document.
type Format string

const (
	FormatTOML Format = "toml"
	FormatYAML Format = "yaml"
)

var (
	keysOnce    sync.Once
	globalKeys  map[string]ValueKind
	repoKeys    map[string]ValueKind
	errNotAFile = errors.New("config path is a directory")
)

func buildKeys(v any) map[string]ValueKind {
	keys := map[string]ValueKind{}
	t := reflect.TypeOf(v)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" || tag == "name" {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Bool:
			keys[tag] = KindBool
		case reflect.String:
			keys[tag] = KindString
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				keys[tag] = KindStringSlice
			}
		}
	}
	return keys
}

func initKeys() {
	keysOnce.Do(func() {
		globalKeys = buildKeys(Global{})
		repoKeys = buildKeys(Repository{})
	})
}

// SectionKeys returns the settable keys and their kinds for the given section.
func SectionKeys(section string) map[string]ValueKind {
	initKeys()
	if section == GlobalSection {
		return globalKeys
	}
	return repoKeys
}

// ValidateSection checks that a section name is either "global" or "owner/repo".
func ValidateSection(section string) error {
	if section == GlobalSection {
		return nil
	}
	parts := strings.Split(section, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid section %q: expected %q or \"owner/repo\"", section, GlobalSection)
	}
	return nil
}

func lookupKey(section, key string) (ValueKind, error) {
	kind, ok := SectionKeys(section)[key]
	if !ok {
		return 0, fmt.Errorf("unknown key %q for section %q (valid keys: %s)", key, section, strings.Join(SortedKeys(section), ", "))
	}
	return kind, nil
}

// SortedKeys returns the settable key names for a section in alphabetical order.
func SortedKeys(section string) []string {
	keys := SectionKeys(section)
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ParseAssignment splits a "key=value" argument.
func ParseAssignment(arg string) (string, string, error) {
	key, value, found := strings.Cut(arg, "=")
	if !found {
		return "", "", fmt.Errorf("invalid assignment %q: expected key=value", arg)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", fmt.Errorf("invalid assignment %q: missing key", arg)
	}
	return key, value, nil
}

// Document is an editable view of a configuration file that preserves
// unknown sections and keys.
type Document struct {
	Path   string
	Format Format
	Data   map[string]any
	// Existed reports whether the file was present when the document was loaded.
	Existed bool
}

// FormatForPath determines the serialization format from a file extension.
func FormatForPath(path string) (Format, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".toml":
		return FormatTOML, nil
	case ".yml", ".yaml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unsupported config file extension %q: expected .toml, .yml, or .yaml", filepath.Ext(path))
	}
}

// DefaultWritePath returns the fallback location used when no config file exists.
func DefaultWritePath(homePath string) string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = filepath.Join(homePath, ".config")
	}
	return filepath.Join(dir, "xget", ".xget.yml")
}

// ResolvePath determines which config file the config command should operate on.
// The explicit path wins, then XGET_CONFIG/EGET_CONFIG, then the first existing
// candidate in the normal search order, then DefaultWritePath.
func ResolvePath(explicitPath string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}
	if custom := configuredPath(); custom != "" {
		return custom, nil
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	for _, candidate := range candidatePaths(homePath) {
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return DefaultWritePath(homePath), nil
}

// LoadDocument reads a config file into an editable document. A missing file
// yields an empty document so that values can be set into it.
func LoadDocument(path string) (*Document, error) {
	format, err := FormatForPath(path)
	if err != nil {
		return nil, err
	}
	doc := &Document{Path: path, Format: format, Data: map[string]any{}}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return doc, nil
		}
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s: %w", path, errNotAFile)
	}

	raw, err := os.ReadFile(path) // #nosec G304 -- path is user supplied by design
	if err != nil {
		return nil, err
	}
	doc.Existed = true
	if len(strings.TrimSpace(string(raw))) == 0 {
		return doc, nil
	}

	switch format {
	case FormatTOML:
		if err := toml.Unmarshal(raw, &doc.Data); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	case FormatYAML:
		if err := yaml.Unmarshal(raw, &doc.Data); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	if doc.Data == nil {
		doc.Data = map[string]any{}
	}
	return doc, nil
}

// Save writes the document back to disk, creating parent directories as needed.
func (d *Document) Save() error {
	var (
		out []byte
		err error
	)
	switch d.Format {
	case FormatTOML:
		out, err = toml.Marshal(d.Data)
	case FormatYAML:
		out, err = yaml.Marshal(d.Data)
	default:
		return fmt.Errorf("unsupported config format %q", d.Format)
	}
	if err != nil {
		return err
	}

	if dir := filepath.Dir(d.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	if err := os.WriteFile(d.Path, out, 0o600); err != nil {
		return err
	}
	d.Existed = true
	return nil
}

func (d *Document) section(name string) (map[string]any, bool) {
	raw, ok := d.Data[name]
	if !ok {
		return nil, false
	}
	switch typed := raw.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		converted := make(map[string]any, len(typed))
		for k, v := range typed {
			converted[fmt.Sprint(k)] = v
		}
		d.Data[name] = converted
		return converted, true
	default:
		return nil, false
	}
}

func (d *Document) ensureSection(name string) map[string]any {
	if existing, ok := d.section(name); ok {
		return existing
	}
	created := map[string]any{}
	d.Data[name] = created
	return created
}

func toStringSlice(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []string:
		return append([]string{}, typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			result = append(result, fmt.Sprint(item))
		}
		return result
	default:
		return []string{fmt.Sprint(typed)}
	}
}

func parseScalar(kind ValueKind, key, value string) (any, error) {
	switch kind {
	case KindBool:
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid boolean value %q for key %q", value, key)
		}
		return parsed, nil
	default:
		return value, nil
	}
}

// FormatValues renders a stored value as one or more display strings.
func FormatValues(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case []any, []string:
		return toStringSlice(typed)
	case bool:
		return []string{strconv.FormatBool(typed)}
	default:
		return []string{fmt.Sprint(typed)}
	}
}

// Set assigns a value. For list-valued keys the value is appended; duplicates
// are ignored. Scalar keys are replaced.
func (d *Document) Set(section, key, value string) error {
	if err := ValidateSection(section); err != nil {
		return err
	}
	kind, err := lookupKey(section, key)
	if err != nil {
		return err
	}
	sec := d.ensureSection(section)

	if kind == KindStringSlice {
		items := toStringSlice(sec[key])
		for _, item := range items {
			if item == value {
				return nil
			}
		}
		sec[key] = append(items, value)
		return nil
	}

	parsed, err := parseScalar(kind, key, value)
	if err != nil {
		return err
	}
	sec[key] = parsed
	return nil
}

// Get returns the display values for a key.
func (d *Document) Get(section, key string) ([]string, error) {
	if err := ValidateSection(section); err != nil {
		return nil, err
	}
	if _, err := lookupKey(section, key); err != nil {
		return nil, err
	}
	sec, ok := d.section(section)
	if !ok {
		return nil, ErrKeyNotSet
	}
	raw, ok := sec[key]
	if !ok || raw == nil {
		return nil, ErrKeyNotSet
	}
	values := FormatValues(raw)
	if len(values) == 0 {
		return nil, ErrKeyNotSet
	}
	return values, nil
}

// Clear removes a key from a section entirely.
func (d *Document) Clear(section, key string) error {
	if err := ValidateSection(section); err != nil {
		return err
	}
	if _, err := lookupKey(section, key); err != nil {
		return err
	}
	sec, ok := d.section(section)
	if !ok {
		return ErrKeyNotSet
	}
	if _, ok := sec[key]; !ok {
		return ErrKeyNotSet
	}
	delete(sec, key)
	return nil
}

// Pop removes a single value from a list-valued key. For scalar keys it clears
// the key when the current value matches.
func (d *Document) Pop(section, key, value string) error {
	if err := ValidateSection(section); err != nil {
		return err
	}
	kind, err := lookupKey(section, key)
	if err != nil {
		return err
	}
	sec, ok := d.section(section)
	if !ok {
		return ErrKeyNotSet
	}
	raw, ok := sec[key]
	if !ok {
		return ErrKeyNotSet
	}

	if kind == KindStringSlice {
		items := toStringSlice(raw)
		remaining := make([]string, 0, len(items))
		removed := false
		for _, item := range items {
			if !removed && item == value {
				removed = true
				continue
			}
			remaining = append(remaining, item)
		}
		if !removed {
			return ErrKeyNotSet
		}
		if len(remaining) == 0 {
			delete(sec, key)
			return nil
		}
		sec[key] = remaining
		return nil
	}

	current := FormatValues(raw)
	if len(current) != 1 || current[0] != value {
		return ErrKeyNotSet
	}
	delete(sec, key)
	return nil
}

// Entry is a flattened section/key/value triple.
type Entry struct {
	Section string
	Key     string
	Value   string
}

// Entries returns every value in the document, sorted by section then key,
// with global first.
func (d *Document) Entries() []Entry {
	sections := make([]string, 0, len(d.Data))
	for name := range d.Data {
		sections = append(sections, name)
	}
	sort.Slice(sections, func(i, j int) bool {
		if sections[i] == GlobalSection {
			return true
		}
		if sections[j] == GlobalSection {
			return false
		}
		return sections[i] < sections[j]
	})

	entries := []Entry{}
	for _, name := range sections {
		sec, ok := d.section(name)
		if !ok {
			entries = append(entries, Entry{Section: name, Key: "", Value: fmt.Sprint(d.Data[name])})
			continue
		}
		keys := make([]string, 0, len(sec))
		for key := range sec {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, value := range FormatValues(sec[key]) {
				entries = append(entries, Entry{Section: name, Key: key, Value: value})
			}
		}
	}
	return entries
}
