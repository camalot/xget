package installed

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Name             string    `yaml:"name"`
	Repo             string    `yaml:"repo"`
	InstallLocation  string    `yaml:"install_location,omitempty"`
	InstalledAt      time.Time `yaml:"installed_at"`
	DownloadURL      string    `yaml:"download_url"`
	Asset            string    `yaml:"asset"`
	ExtractedFiles   []string  `yaml:"extracted_files,omitempty"`
	Options          Options   `yaml:"options,omitempty"`
	RefreshedAt      time.Time `yaml:"refreshed_at"`
	CurrentVersion   string    `yaml:"current_version,omitempty"`
	CurrentTag       string    `yaml:"current_tag,omitempty"`
	InstalledVersion string    `yaml:"installed_version,omitempty"`
	InstalledTag     string    `yaml:"installed_tag,omitempty"`
	Source           string    `yaml:"source"`
	SHA256           string    `yaml:"sha256"`
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
	return store, nil
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

func Upsert(path string, pkg Package) error {
	store, err := Load(path)
	if err != nil {
		return err
	}
	if store.Packages == nil {
		store.Packages = map[string]Package{}
	}
	if existing, ok := store.Packages[pkg.Name]; ok && !existing.InstalledAt.IsZero() {
		pkg.InstalledAt = existing.InstalledAt
	}
	store.Packages[pkg.Name] = pkg
	return Save(path, store)
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
