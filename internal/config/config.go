package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

type Global struct {
	All          bool   `mapstructure:"all" toml:"all" yaml:"all"`
	DownloadOnly bool   `mapstructure:"download_only" toml:"download_only" yaml:"download_only"`
	File         string `mapstructure:"file" toml:"file" yaml:"file"`
	GithubToken  string `mapstructure:"github_token" toml:"github_token" yaml:"github_token"`
	Quiet        bool   `mapstructure:"quiet" toml:"quiet" yaml:"quiet"`
	ShowHash     bool   `mapstructure:"show_hash" toml:"show_hash" yaml:"show_hash"`
	Source       bool   `mapstructure:"download_source" toml:"download_source" yaml:"download_source"`
	System       string `mapstructure:"system" toml:"system" yaml:"system"`
	Target       string `mapstructure:"target" toml:"target" yaml:"target"`
	UpgradeOnly  bool   `mapstructure:"upgrade_only" toml:"upgrade_only" yaml:"upgrade_only"`
	DisableSSL   bool   `mapstructure:"disable_ssl" toml:"disable_ssl" yaml:"disable_ssl"`
}

type Repository struct {
	All          bool     `mapstructure:"all" toml:"all" yaml:"all"`
	AssetFilters []string `mapstructure:"asset_filters" toml:"asset_filters" yaml:"asset_filters"`
	DownloadOnly bool     `mapstructure:"download_only" toml:"download_only" yaml:"download_only"`
	File         string   `mapstructure:"file" toml:"file" yaml:"file"`
	Name         string   `mapstructure:"name" toml:"name" yaml:"name"`
	Quiet        bool     `mapstructure:"quiet" toml:"quiet" yaml:"quiet"`
	ShowHash     bool     `mapstructure:"show_hash" toml:"show_hash" yaml:"show_hash"`
	Source       bool     `mapstructure:"download_source" toml:"download_source" yaml:"download_source"`
	System       string   `mapstructure:"system" toml:"system" yaml:"system"`
	Tag          string   `mapstructure:"tag" toml:"tag" yaml:"tag"`
	Target       string   `mapstructure:"target" toml:"target" yaml:"target"`
	UpgradeOnly  bool     `mapstructure:"upgrade_only" toml:"upgrade_only" yaml:"upgrade_only"`
	Verify       string   `mapstructure:"verify_sha256" toml:"verify_sha256" yaml:"verify_sha256"`
	DisableSSL   bool     `mapstructure:"disable_ssl" toml:"disable_ssl" yaml:"disable_ssl"`
}

type Config struct {
	Path         string
	Global       Global
	Repositories map[string]Repository
}

func Default() *Config {
	return &Config{
		Global: Global{
			All:          false,
			DownloadOnly: false,
			GithubToken:  "",
			Quiet:        false,
			ShowHash:     false,
			Source:       false,
			UpgradeOnly:  false,
			DisableSSL:   false,
		},
		Repositories: map[string]Repository{},
	}
}

func GetOSConfigPath(homePath string, ext string) string {
	var configDir string

	defaultConfig := map[string]string{
		"windows": "LocalAppData",
		"default": ".config",
	}

	var goos string
	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("LOCALAPPDATA")
		goos = "windows"
	default:
		configDir = os.Getenv("XDG_CONFIG_HOME")
		goos = "default"
	}

	if configDir == "" {
		configDir = filepath.Join(homePath, defaultConfig[goos])
	}

	return filepath.Join(configDir, "xget", "xget."+ext)
}

func candidatePaths(homePath string) []string {
	candidates := []string{}

	if custom, ok := os.LookupEnv("XGET_CONFIG"); ok && custom != "" {
		candidates = append(candidates, custom)
	}

	for _, ext := range []string{"toml", "yaml", "yml"} {
		candidates = append(candidates,
			filepath.Join(homePath, ".xget."+ext),
			"xget."+ext,
			GetOSConfigPath(homePath, ext),
		)
	}

	return candidates
}

func loadFromFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := Default()
	cfg.Path = path

	if v.IsSet("global") {
		if err := v.Sub("global").Unmarshal(&cfg.Global); err != nil {
			return nil, fmt.Errorf("parse global: %w", err)
		}
	}

	for key := range v.AllSettings() {
		if key == "global" {
			continue
		}
		section := v.Sub(key)
		if section == nil {
			continue
		}
		repo := Repository{}
		if err := section.Unmarshal(&repo); err != nil {
			return nil, fmt.Errorf("parse repository %q: %w", key, err)
		}
		repo.Name = key
		if !section.IsSet("all") {
			repo.All = cfg.Global.All
		}
		if !section.IsSet("asset_filters") {
			repo.AssetFilters = []string{}
		}
		if !section.IsSet("download_only") {
			repo.DownloadOnly = cfg.Global.DownloadOnly
		}
		if !section.IsSet("quiet") {
			repo.Quiet = cfg.Global.Quiet
		}
		if !section.IsSet("show_hash") {
			repo.ShowHash = cfg.Global.ShowHash
		}
		if !section.IsSet("download_source") {
			repo.Source = cfg.Global.Source
		}
		if !section.IsSet("upgrade_only") {
			repo.UpgradeOnly = cfg.Global.UpgradeOnly
		}
		if !section.IsSet("disable_ssl") {
			repo.DisableSSL = cfg.Global.DisableSSL
		}
		if !section.IsSet("target") {
			repo.Target = cfg.Global.Target
		}
		cfg.Repositories[key] = repo
	}
	return cfg, nil
}

func Load() (*Config, error) {
	homePath, _ := os.UserHomeDir()
	var lastNotExist error

	for _, p := range candidatePaths(homePath) {
		cfg, err := loadFromFile(p)
		if err == nil {
			return cfg, nil
		}

		var notFound viper.ConfigFileNotFoundError
		if errors.Is(err, os.ErrNotExist) || errors.As(err, &notFound) {
			lastNotExist = err
			continue
		}
		return nil, fmt.Errorf("%s: %w", p, err)
	}

	_ = lastNotExist
	return Default(), nil
}
