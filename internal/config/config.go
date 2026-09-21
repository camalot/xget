package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Global struct {
	All                 bool     `mapstructure:"all" toml:"all" yaml:"all"`
	Ignore              []string `mapstructure:"ignore" toml:"ignore" yaml:"ignore"`
	DownloadOnly        bool     `mapstructure:"download_only" toml:"download_only" yaml:"download_only"`
	File                string   `mapstructure:"file" toml:"file" yaml:"file"`
	GithubToken         string   `mapstructure:"github_token" toml:"github_token" yaml:"github_token"`
	Quiet               bool     `mapstructure:"quiet" toml:"quiet" yaml:"quiet"`
	ShowHash            bool     `mapstructure:"show_hash" toml:"show_hash" yaml:"show_hash"`
	Source              bool     `mapstructure:"download_source" toml:"download_source" yaml:"download_source"`
	System              string   `mapstructure:"system" toml:"system" yaml:"system"`
	Target              string   `mapstructure:"target" toml:"target" yaml:"target"`
	UpgradeOnly         bool     `mapstructure:"upgrade_only" toml:"upgrade_only" yaml:"upgrade_only"`
	DisableSSL          bool     `mapstructure:"disable_ssl" toml:"disable_ssl" yaml:"disable_ssl"`
	SourceType          string   `mapstructure:"source" toml:"source" yaml:"source"`
	DisableTokenWarning bool     `mapstructure:"disable_token_warning" toml:"disable_token_warning" yaml:"disable_token_warning"`
}

type Repository struct {
	All          bool     `mapstructure:"all" toml:"all" yaml:"all"`
	AssetFilters []string `mapstructure:"asset_filters" toml:"asset_filters" yaml:"asset_filters"`
	Ignore       []string `mapstructure:"ignore" toml:"ignore" yaml:"ignore"`
	DownloadOnly bool     `mapstructure:"download_only" toml:"download_only" yaml:"download_only"`
	File         string   `mapstructure:"file" toml:"file" yaml:"file"`
	Name         string   `mapstructure:"name" toml:"name" yaml:"name"`
	Prerelease   bool     `mapstructure:"pre_release" toml:"pre_release" yaml:"pre_release"`
	Quiet        bool     `mapstructure:"quiet" toml:"quiet" yaml:"quiet"`
	ShowHash     bool     `mapstructure:"show_hash" toml:"show_hash" yaml:"show_hash"`
	Source       bool     `mapstructure:"download_source" toml:"download_source" yaml:"download_source"`
	System       string   `mapstructure:"system" toml:"system" yaml:"system"`
	Tag          string   `mapstructure:"tag" toml:"tag" yaml:"tag"`
	Target       string   `mapstructure:"target" toml:"target" yaml:"target"`
	UpgradeOnly  bool     `mapstructure:"upgrade_only" toml:"upgrade_only" yaml:"upgrade_only"`
	Verify       string   `mapstructure:"verify_sha256" toml:"verify_sha256" yaml:"verify_sha256"`
	DisableSSL   bool     `mapstructure:"disable_ssl" toml:"disable_ssl" yaml:"disable_ssl"`
	SourceType   string   `mapstructure:"source" toml:"source" yaml:"source"`
}

type Source struct {
	Name                string   `mapstructure:"-" toml:"-" yaml:"-"`
	Type                string   `mapstructure:"type" toml:"type" yaml:"type"`
	Host                string   `mapstructure:"host" toml:"host" yaml:"host"`
	APIURL              string   `mapstructure:"api_url" toml:"api_url" yaml:"api_url"`
	TokenEnv            []string `mapstructure:"token_env" toml:"token_env" yaml:"token_env"`
	Token               string   `mapstructure:"token" toml:"token" yaml:"token"`
	DisableTokenWarning bool     `mapstructure:"disable_token_warning" toml:"disable_token_warning" yaml:"disable_token_warning"`
}

type Config struct {
	Path         string
	Global       Global
	Repositories map[string]Repository
	Sources      map[string]Source
}

func Default() *Config {
	return &Config{
		Global: Global{
			All:          false,
			Ignore:       []string{},
			DownloadOnly: false,
			GithubToken:  "",
			Quiet:        false,
			ShowHash:     false,
			Source:       false,
			UpgradeOnly:  false,
			DisableSSL:   false,
		},
		Repositories: map[string]Repository{},
		Sources: map[string]Source{
			"github": defaultSource("github", "github"),
			"gitlab": defaultSource("gitlab", "gitlab"),
		},
	}
}

func defaultSource(name, sourceType string) Source {
	source := Source{Name: name, Type: strings.ToLower(sourceType)}
	switch source.Type {
	case "github":
		source.Host = "github.com"
		source.APIURL = "https://api.github.com"
		source.TokenEnv = []string{"XGET_GITHUB_TOKEN", "GITHUB_TOKEN", "EGET_GITHUB_TOKEN"}
	case "gitlab":
		source.Host = "gitlab.com"
		source.APIURL = "https://gitlab.com/api/v4"
		source.TokenEnv = []string{"XGET_GITLAB_TOKEN", "GITLAB_TOKEN"}
	}
	return source
}

func normalizeSource(name string, configured Source) (Source, error) {
	sourceType := strings.ToLower(configured.Type)
	if sourceType == "" && (strings.EqualFold(name, "github") || strings.EqualFold(name, "gitlab")) {
		sourceType = strings.ToLower(name)
	}
	if sourceType != "github" && sourceType != "gitlab" {
		return Source{}, fmt.Errorf("source %q has unsupported type %q (expected github or gitlab)", name, configured.Type)
	}

	resolved := defaultSource(name, sourceType)
	if configured.Host != "" {
		resolved.Host = configured.Host
		if configured.APIURL == "" {
			if sourceType == "github" {
				resolved.APIURL = "https://" + configured.Host + "/api/v3"
			} else {
				resolved.APIURL = "https://" + configured.Host + "/api/v4"
			}
		}
	}
	if configured.APIURL != "" {
		resolved.APIURL = strings.TrimRight(configured.APIURL, "/")
	}
	if configured.TokenEnv != nil {
		resolved.TokenEnv = configured.TokenEnv
	}
	resolved.Token = configured.Token
	resolved.DisableTokenWarning = configured.DisableTokenWarning
	return resolved, nil
}

func (c *Config) ResolveSource(name string) (Source, error) {
	if name == "" {
		name = "github"
	}
	name = strings.ToLower(name)
	source, ok := c.Sources[name]
	if !ok && (name == "github" || name == "gitlab") {
		return defaultSource(name, name), nil
	}
	if !ok {
		return Source{}, fmt.Errorf("source profile %q is not configured", name)
	}
	return source, nil
}

func GetOSConfigPath(homePath string, ext string) string {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("LOCALAPPDATA")
		if configDir == "" {
			configDir = filepath.Join(homePath, "AppData", "Local")
		}
	default:
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(homePath, ".config")
		}
	}

	return filepath.Join(configDir, "xget", ".xget."+ext)
}

func configuredPath() string {
	if custom, ok := os.LookupEnv("XGET_CONFIG"); ok && custom != "" {
		return custom
	}
	if custom, ok := os.LookupEnv("EGET_CONFIG"); ok && custom != "" {
		return custom
	}
	return ""
}

func candidatePaths(homePath string) []string {
	candidates := []string{}

	for _, base := range []string{".xget", ".eget"} {
		for _, ext := range []string{"toml", "yml", "yaml"} {
			candidates = append(candidates,
				filepath.Join(".", base+"."+ext),
				filepath.Join(homePath, base+"."+ext),
				filepath.Join(homePath, ".config", "xget", base+"."+ext),
			)
			if runtime.GOOS == "windows" {
				localAppData := os.Getenv("LOCALAPPDATA")
				if localAppData == "" {
					localAppData = filepath.Join(homePath, "AppData", "Local")
				}
				candidates = append(candidates,
					filepath.Join(localAppData, "xget", base+"."+ext),
				)
			}
		}
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

	if sources := v.Sub("sources"); sources != nil {
		for name := range sources.AllSettings() {
			section := sources.Sub(name)
			if section == nil {
				continue
			}
			configured := Source{}
			if err := section.Unmarshal(&configured); err != nil {
				return nil, fmt.Errorf("parse source %q: %w", name, err)
			}
			resolved, err := normalizeSource(name, configured)
			if err != nil {
				return nil, err
			}
			cfg.Sources[strings.ToLower(name)] = resolved
		}
	}
	warnStoredTokens(cfg, os.Stderr)

	for key := range v.AllSettings() {
		if key == "global" || key == "sources" {
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
		if !section.IsSet("ignore") {
			repo.Ignore = cfg.Global.Ignore
		}
		if !section.IsSet("download_only") {
			repo.DownloadOnly = cfg.Global.DownloadOnly
		}
		if !section.IsSet("pre_release") {
			repo.Prerelease = false
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
		if !section.IsSet("source") {
			repo.SourceType = cfg.Global.SourceType
		}
		if !section.IsSet("system") {
			repo.System = cfg.Global.System
		}
		if !section.IsSet("file") {
			repo.File = cfg.Global.File
		}
		if !section.IsSet("target") {
			repo.Target = cfg.Global.Target
		}
		cfg.Repositories[key] = repo
	}
	return cfg, nil
}

func warnStoredTokens(cfg *Config, output io.Writer) {
	github := cfg.Sources["github"]
	if isPlaintextToken(cfg.Global.GithubToken) && !cfg.Global.DisableTokenWarning && !github.DisableTokenWarning {
		_, _ = fmt.Fprintln(output, "Warning: global.github_token is stored in the config file as plaintext. Prefer a token environment variable or token file. To disable this warning, set global.disable_token_warning or sources.github.disable_token_warning to true.")
	}
	for name, source := range cfg.Sources {
		if isPlaintextToken(source.Token) && !cfg.Global.DisableTokenWarning && !source.DisableTokenWarning {
			_, _ = fmt.Fprintf(output, "Warning: sources.%s.token is stored in the config file as plaintext. Prefer one of sources.%s.token_env or a token file. To disable this warning, set global.disable_token_warning or sources.%s.disable_token_warning to true.\n", name, name, name)
		}
	}
}

func isPlaintextToken(token string) bool {
	return token != "" && !strings.HasPrefix(token, "@")
}

func Load(explicitPath ...string) (*Config, error) {
	homePath, _ := os.UserHomeDir()

	if len(explicitPath) > 0 && explicitPath[0] != "" {
		cfg, err := loadFromFile(explicitPath[0])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", explicitPath[0], err)
		}
		return cfg, nil
	}

	if custom := configuredPath(); custom != "" {
		cfg, err := loadFromFile(custom)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", custom, err)
		}
		return cfg, nil
	}

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

// SubstituteTemplateVars replaces {{.OS}} and {{.Arch}} in a filter string
// with the actual OS and architecture values based on the system parameter.
func SubstituteTemplateVars(filter string, system string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// If system is specified, parse it to extract OS and Arch
	if system != "" && system != "all" {
		parts := strings.Split(system, "/")
		if len(parts) >= 2 {
			goos = parts[0]
			goarch = parts[1]
		}
	}

	result := strings.ReplaceAll(filter, "{{.OS}}", goos)
	result = strings.ReplaceAll(result, "{{.Arch}}", goarch)
	return result
}

// SubstituteTemplateVarsSlice applies SubstituteTemplateVars to every entry in filters.
func SubstituteTemplateVarsSlice(filters []string, system string) []string {
	if len(filters) == 0 {
		return filters
	}
	result := make([]string, len(filters))
	for i, filter := range filters {
		result[i] = SubstituteTemplateVars(filter, system)
	}
	return result
}
