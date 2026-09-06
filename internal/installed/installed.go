package installed

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/camalot/xget/internal/home"
	"go.yaml.in/yaml/v3"
)

type Options struct {
	Tag            string   `yaml:"tag,omitempty"`
	Prerelease     bool     `yaml:"pre_release,omitempty"`
	DownloadSource bool     `yaml:"download_source,omitempty"`
	Output         string   `yaml:"output,omitempty"`
	System         string   `yaml:"system,omitempty"`
	ExtractFile    string   `yaml:"file,omitempty"`
	All            bool     `yaml:"all,omitempty"`
	DownloadOnly   bool     `yaml:"download_only,omitempty"`
	UpgradeOnly    bool     `yaml:"upgrade_only,omitempty"`
	Asset          []string `yaml:"asset,omitempty"`
	Ignore         []string `yaml:"ignore,omitempty"`
	Verify         string   `yaml:"verify,omitempty"`
}

type Package struct {
	Name            string    `yaml:"name"`
	Repo            string    `yaml:"repo,omitempty"`
	InstallLocation string    `yaml:"install_location,omitempty"`
	InstalledAt     time.Time `yaml:"installed_at"`
	DownloadURL     string    `yaml:"download_url"`
	Asset           string    `yaml:"asset"`
	ExtractedFiles  []string  `yaml:"extracted_files,omitempty"`
	Options         Options   `yaml:"options,omitempty"`
	RefreshedAt     time.Time `yaml:"refreshed_at"`
	CurrentTag      string    `yaml:"current_tag,omitempty"`
	InstalledTag    string    `yaml:"installed_tag,omitempty"`
	Source          string    `yaml:"source"`
	SHA256          string    `yaml:"sha256"`
}

// Store maps a package key to every install location tracked for that package.
type Store struct {
	Packages map[string][]Package `yaml:"packages"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "xget", ".xget.installed.yml"), nil
}

func Load(path string) (*Store, error) {
	store := &Store{Packages: map[string][]Package{}}
	// #nosec G304 -- path is xget's installed metadata store path or a caller-provided test path.
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return store, nil
	}
	if err := yaml.Unmarshal(data, store); err != nil {
		return nil, err
	}
	if store.Packages == nil {
		store.Packages = map[string][]Package{}
	}
	store.migrate()
	return store, nil
}

// UnmarshalYAML accepts both the current list-per-package layout and the older
// layout that stored a single record per package.
func (s *Store) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Packages map[string]yaml.Node `yaml:"packages"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	s.Packages = map[string][]Package{}
	for key := range raw.Packages {
		entry := raw.Packages[key]
		switch entry.Kind {
		case yaml.SequenceNode:
			var records []Package
			if err := entry.Decode(&records); err != nil {
				return err
			}
			s.Packages[key] = records
		case yaml.MappingNode:
			var record Package
			if err := entry.Decode(&record); err != nil {
				return err
			}
			s.Packages[key] = []Package{record}
		}
	}
	return nil
}

func (s *Store) migrate() {
	if s == nil || len(s.Packages) == 0 {
		return
	}
	migrated := map[string][]Package{}
	for key, records := range s.Packages {
		for _, pkg := range records {
			pkg = normalize(pkg)
			newKey := pkg.Key()
			if newKey == "unknown:" {
				newKey = key
			}
			migrated[newKey] = append(migrated[newKey], pkg)
		}
	}
	for key := range migrated {
		sortRecords(migrated[key])
	}
	s.Packages = migrated
}

func normalize(pkg Package) Package {
	if pkg.Repo != "" && !strings.Contains(pkg.Name, "/") {
		pkg.Name = pkg.Repo
	}
	pkg.Repo = ""
	return pkg
}

func sortRecords(records []Package) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].InstallLocation < records[j].InstallLocation
	})
}

func Save(path string, store *Store) error {
	if store == nil {
		return fmt.Errorf("installed store cannot be nil")
	}
	if store.Packages == nil {
		store.Packages = map[string][]Package{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	data, err := yaml.Marshal(store)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s *Store) MarshalYAML() (interface{}, error) {
	packages := &yaml.Node{Kind: yaml.MappingNode}
	keys := make([]string, 0, len(s.Packages))
	for key := range s.Packages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := yaml.Node{}
		if err := value.Encode(s.Packages[key]); err != nil {
			return nil, err
		}
		packages.Content = append(packages.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key, Style: yaml.DoubleQuotedStyle},
			&value,
		)
	}
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "packages"},
			packages,
		},
	}, nil
}

func Upsert(path string, pkg Package) error {
	store, err := Load(path)
	if err != nil {
		return err
	}
	store.Set(pkg)
	return Save(path, store)
}

// Set replaces the record tracked for the package's install location, adding a
// new record when that location is not tracked yet.
func (s *Store) Set(pkg Package) {
	if s.Packages == nil {
		s.Packages = map[string][]Package{}
	}
	pkg = normalize(pkg)
	key := pkg.Key()
	for i, existing := range s.Packages[key] {
		if SamePath(existing.InstallLocation, pkg.InstallLocation) {
			s.Packages[key][i] = pkg
			return
		}
	}
	s.Packages[key] = append(s.Packages[key], pkg)
	sortRecords(s.Packages[key])
}

// Remove drops the record for the package's install location, and drops the key
// entirely once its last location is gone.
func (s *Store) Remove(pkg Package) {
	key := pkg.Key()
	records, ok := s.Packages[key]
	if !ok {
		return
	}
	kept := make([]Package, 0, len(records))
	for _, existing := range records {
		if SamePath(existing.InstallLocation, pkg.InstallLocation) {
			continue
		}
		kept = append(kept, existing)
	}
	if len(kept) == 0 {
		delete(s.Packages, key)
		return
	}
	s.Packages[key] = kept
}

// Find returns the record tracked for key at location.
func (s *Store) Find(key, location string) (Package, bool) {
	for _, pkg := range s.Packages[key] {
		if SamePath(pkg.InstallLocation, location) {
			return pkg, true
		}
	}
	return Package{}, false
}

func (p Package) Key() string {
	source := strings.ToLower(p.Source)
	if source == "" {
		source = "unknown"
	}
	return source + ":" + p.Name
}

// SamePath reports whether two install locations refer to the same directory,
// comparing them after tilde expansion and absolute-path resolution.
func SamePath(first, second string) bool {
	first, second = expandPath(first), expandPath(second)
	if filepath.Clean(first) == filepath.Clean(second) {
		return true
	}
	firstAbsolute, firstErr := filepath.Abs(first)
	secondAbsolute, secondErr := filepath.Abs(second)
	return firstErr == nil && secondErr == nil && strings.EqualFold(firstAbsolute, secondAbsolute)
}

func expandPath(path string) string {
	expanded, err := home.Expand(path)
	if err != nil {
		return path
	}
	return expanded
}

// SortedPackages flattens the store into one entry per tracked install
// location, ordered by package key and then by location.
func SortedPackages(store *Store) []Package {
	if store == nil || len(store.Packages) == 0 {
		return nil
	}
	keys := make([]string, 0, len(store.Packages))
	for key := range store.Packages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	packages := make([]Package, 0, len(keys))
	for _, key := range keys {
		records := append([]Package(nil), store.Packages[key]...)
		sortRecords(records)
		packages = append(packages, records...)
	}
	return packages
}
