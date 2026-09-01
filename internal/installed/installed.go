package installed

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

type Store struct {
	Packages map[string]Package `yaml:"packages"`
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "xget", ".xget.installed.yml"), nil
}

func Load(path string) (*Store, error) {
	store := &Store{Packages: map[string]Package{}}
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
		store.Packages = map[string]Package{}
	}
	store.migrate()
	return store, nil
}

func (s *Store) migrate() {
	if s == nil || len(s.Packages) == 0 {
		return
	}
	migrated := map[string]Package{}
	for key, pkg := range s.Packages {
		if pkg.Repo != "" && !strings.Contains(pkg.Name, "/") {
			pkg.Name = pkg.Repo
		}
		pkg.Repo = ""
		newKey := pkg.Key()
		if newKey == "unknown:" {
			newKey = key
		}
		migrated[newKey] = pkg
	}
	s.Packages = migrated
}

func Save(path string, store *Store) error {
	if store == nil {
		return fmt.Errorf("installed store cannot be nil")
	}
	if store.Packages == nil {
		store.Packages = map[string]Package{}
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
	if store.Packages == nil {
		store.Packages = map[string]Package{}
	}
	if pkg.Repo != "" && !strings.Contains(pkg.Name, "/") {
		pkg.Name = pkg.Repo
	}
	pkg.Repo = ""
	key := pkg.Key()
	if existing, ok := store.Packages[key]; ok && !existing.InstalledAt.IsZero() {
		pkg.InstalledAt = existing.InstalledAt
	}
	store.Packages[key] = pkg
	return Save(path, store)
}

func (p Package) Key() string {
	source := strings.ToLower(p.Source)
	if source == "" {
		source = "unknown"
	}
	return source + ":" + p.Name
}

func SortedPackages(store *Store) []Package {
	if store == nil || len(store.Packages) == 0 {
		return nil
	}
	names := make([]string, 0, len(store.Packages))
	for name := range store.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	packages := make([]Package, 0, len(names))
	for _, name := range names {
		packages = append(packages, store.Packages[name])
	}
	return packages
}
