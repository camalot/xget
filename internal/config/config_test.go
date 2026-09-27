package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestCandidatePathsUsesDotPrefixedLocationsInOrder(t *testing.T) {
	homePath := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(homePath, 0o750); err != nil {
		t.Fatal(err)
	}

	got := candidatePaths(homePath)
	want := []string{}
	locations := []string{".", homePath, filepath.Join(homePath, ".config", "xget")}
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homePath, "AppData", "Local")
		}
		locations = append(locations, filepath.Join(localAppData, "xget"))
	}
	for _, location := range locations {
		for _, base := range []string{".xget", ".eget"} {
			for _, ext := range []string{"toml", "yml", "yaml"} {
				want = append(want, filepath.Join(location, base+"."+ext))
			}
		}
	}

	if len(got) != len(want) {
		t.Fatalf("candidatePaths length mismatch: got %d, want %d\nGot: %v\nWant: %v", len(got), len(want), got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidatePaths[%d] mismatch: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestConfiguredPathPrefersXgetAndExplicitOverride(t *testing.T) {
	oldXget := os.Getenv("XGET_CONFIG")
	oldEget := os.Getenv("EGET_CONFIG")
	defer func() {
		if err := os.Setenv("XGET_CONFIG", oldXget); err != nil {
			t.Fatalf("restore XGET_CONFIG: %v", err)
		}
		if err := os.Setenv("EGET_CONFIG", oldEget); err != nil {
			t.Fatalf("restore EGET_CONFIG: %v", err)
		}
	}()

	if err := os.Setenv("XGET_CONFIG", "/tmp/xget.toml"); err != nil {
		t.Fatalf("set XGET_CONFIG: %v", err)
	}
	if err := os.Setenv("EGET_CONFIG", "/tmp/eget.toml"); err != nil {
		t.Fatalf("set EGET_CONFIG: %v", err)
	}
	if got := configuredPath(); got != "/tmp/xget.toml" {
		t.Fatalf("configuredPath() = %q, want %q", got, "/tmp/xget.toml")
	}

	if err := os.Unsetenv("XGET_CONFIG"); err != nil {
		t.Fatalf("unset XGET_CONFIG: %v", err)
	}
	if got := configuredPath(); got != "/tmp/eget.toml" {
		t.Fatalf("configuredPath() = %q, want %q", got, "/tmp/eget.toml")
	}
}

func TestLoadMergesConfigLayers(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	t.Chdir(cwd)

	basePath := filepath.Join(home, ".config", "xget", ".xget.yml")
	if err := os.MkdirAll(filepath.Dir(basePath), 0o750); err != nil {
		t.Fatal(err)
	}
	base := "global:\n  target: ~/.local/bin\n  system: linux/amd64\n  xget_update_check: false\n  ignore: [base]\nowner/base:\n  tag: stable\nowner/shared:\n  file: base.zip\nsources:\n  github:\n    token_env: [BASE_TOKEN]\n"
	if err := os.WriteFile(basePath, []byte(base), 0o600); err != nil {
		t.Fatal(err)
	}
	homePath := filepath.Join(home, ".xget.toml")
	if err := os.WriteFile(homePath, []byte("[global]\ntarget = \"~/bin\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cwdPath := filepath.Join(cwd, ".xget.yml")
	cwdConfig := "global:\n  target: /usr/local/bin\n  ignore: [local]\nowner/shared:\n  tag: latest\nowner/local:\n  quiet: true\nsources:\n  github:\n    host: example.com\n"
	if err := os.WriteFile(cwdPath, []byte(cwdConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadQuiet()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != filepath.Join(".", ".xget.yml") || cfg.Global.Target != "/usr/local/bin" || cfg.Global.System != "linux/amd64" || cfg.Global.XgetUpdateCheck || !cfg.Global.ConfigMerge {
		t.Fatalf("unexpected merged global or path: %#v", cfg)
	}
	if !reflect.DeepEqual(cfg.Global.Ignore, []string{"local"}) || cfg.Repositories["owner/base"].Tag != "stable" || cfg.Repositories["owner/shared"].File != "base.zip" || cfg.Repositories["owner/shared"].Tag != "latest" || cfg.Repositories["owner/local"].Target != "/usr/local/bin" {
		t.Fatalf("unexpected merged repositories or list: %#v", cfg)
	}
	if cfg.Sources["github"].Host != "example.com" || !reflect.DeepEqual(cfg.Sources["github"].TokenEnv, []string{"BASE_TOKEN"}) {
		t.Fatalf("unexpected merged source: %#v", cfg.Sources["github"])
	}

	explicitPath := filepath.Join(t.TempDir(), ".xget.toml")
	if err := os.WriteFile(explicitPath, []byte("[global]\ntarget = \"/opt/bin\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	explicit, err := LoadQuiet(explicitPath)
	if err != nil {
		t.Fatal(err)
	}
	if explicit.Path != explicitPath || explicit.Global.Target != "/opt/bin" || explicit.Global.System != "linux/amd64" {
		t.Fatalf("unexpected explicit override: %#v", explicit)
	}
}

func TestLoadHighestPriorityCanDisableMerge(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	t.Chdir(cwd)

	basePath := filepath.Join(home, ".config", "xget", ".xget.toml")
	if err := os.MkdirAll(filepath.Dir(basePath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(basePath, []byte("[global]\ntarget = \"~/bin\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cwdPath := filepath.Join(cwd, ".xget.yml")
	if err := os.WriteFile(cwdPath, []byte("global:\n  config_merge: false\n  quiet: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadQuiet()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Global.ConfigMerge || !cfg.Global.Quiet || cfg.Global.Target != "" {
		t.Fatalf("expected cwd config alone, got %#v", cfg.Global)
	}

	explicitPath := filepath.Join(t.TempDir(), ".xget.toml")
	if err := os.WriteFile(explicitPath, []byte("[global]\ntarget = \"/opt/bin\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	explicit, err := LoadQuiet(explicitPath)
	if err != nil {
		t.Fatal(err)
	}
	if explicit.Global.Target != "/opt/bin" || !explicit.Global.Quiet || explicit.Global.ConfigMerge {
		t.Fatalf("lower-priority opt-out should not stop merging: %#v", explicit.Global)
	}

	if err := os.WriteFile(basePath, []byte("[global\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadQuiet(); err != nil {
		t.Fatalf("disabled merge should skip lower-priority invalid files: %v", err)
	}
}

func TestLoadSupportsIgnoreArrayInGlobalAndRepository(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".xget.toml")
	content := `[global]
ignore = ["~\\.sbom\\.json$", "not:debug"]

["owner/repo"]
ignore = ["~\\.sig$", "notes"]
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !reflect.DeepEqual(cfg.Global.Ignore, []string{"~\\.sbom\\.json$", "not:debug"}) {
		t.Fatalf("unexpected global ignore: %#v", cfg.Global.Ignore)
	}

	repo := cfg.Repositories["owner/repo"]
	if !reflect.DeepEqual(repo.Ignore, []string{"~\\.sig$", "notes"}) {
		t.Fatalf("unexpected repo ignore: %#v", repo.Ignore)
	}
}

func TestLoadRepositoryIgnoreFallsBackToGlobal(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".xget.toml")
	content := `[global]
ignore = ["~\\.sbom\\.json$", "^^caret"]

["owner/repo"]
asset_filters = [".zip"]
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	repo := cfg.Repositories["owner/repo"]
	if !reflect.DeepEqual(repo.Ignore, []string{"~\\.sbom\\.json$", "^^caret"}) {
		t.Fatalf("expected repo ignore to fall back to global, got %#v", repo.Ignore)
	}
}

func TestLoadRepositorySourceTypeFallsBackToGlobal(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".xget.toml")
	content := `[global]
source = "GitHub"

["owner/repo"]
asset_filters = [".zip"]
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	repo := cfg.Repositories["owner/repo"]
	if repo.SourceType != "GitHub" {
		t.Fatalf("expected repo source to fall back to global, got %q", repo.SourceType)
	}
}

func TestLoadRepositorySystemAndFileFallBackToGlobal(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".xget.toml")
	content := `[global]
system = "linux/amd64"
file = "global.bin"

["owner/inherits"]
asset_filters = [".zip"]

["owner/overrides"]
system = "darwin/arm64"
file = "*.exe"
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	inherits := cfg.Repositories["owner/inherits"]
	if inherits.System != "linux/amd64" {
		t.Fatalf("expected repo system to fall back to global, got %q", inherits.System)
	}
	if inherits.File != "global.bin" {
		t.Fatalf("expected repo file to fall back to global, got %q", inherits.File)
	}

	overrides := cfg.Repositories["owner/overrides"]
	if overrides.System != "darwin/arm64" {
		t.Fatalf("expected repo system to win, got %q", overrides.System)
	}
	if overrides.File != "*.exe" {
		t.Fatalf("expected repo file to win, got %q", overrides.File)
	}
}

func TestLoadRepositorySystemAndFileStayEmptyWithoutGlobal(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".xget.toml")
	content := `["owner/repo"]
asset_filters = [".zip"]
`
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	repo := cfg.Repositories["owner/repo"]
	if repo.System != "" || repo.File != "" {
		t.Fatalf("expected empty system/file, got %q/%q", repo.System, repo.File)
	}
}

func TestSubstituteTemplateVarsUsesGivenSystem(t *testing.T) {
	got := SubstituteTemplateVars("{{.OS}}_{{.Arch}}.tar.gz", "linux/arm64")
	want := "linux_arm64.tar.gz"
	if got != want {
		t.Fatalf("SubstituteTemplateVars() = %q, want %q", got, want)
	}
}

func TestSubstituteTemplateVarsFallsBackToRuntimeWhenSystemEmptyOrAll(t *testing.T) {
	want := runtime.GOOS + "_" + runtime.GOARCH

	if got := SubstituteTemplateVars("{{.OS}}_{{.Arch}}", ""); got != want {
		t.Fatalf("SubstituteTemplateVars() with empty system = %q, want %q", got, want)
	}
	if got := SubstituteTemplateVars("{{.OS}}_{{.Arch}}", "all"); got != want {
		t.Fatalf("SubstituteTemplateVars() with all system = %q, want %q", got, want)
	}
}

func TestSubstituteTemplateVarsLeavesNonTemplateFiltersUnchanged(t *testing.T) {
	got := SubstituteTemplateVars("~\\.sbom\\.json$", "linux/amd64")
	want := "~\\.sbom\\.json$"
	if got != want {
		t.Fatalf("SubstituteTemplateVars() = %q, want %q", got, want)
	}
}

func TestSubstituteTemplateVarsSliceAppliesToEveryEntry(t *testing.T) {
	got := SubstituteTemplateVarsSlice([]string{"{{.OS}}_{{.Arch}}.zip", "not:debug"}, "windows/amd64")
	want := []string{"windows_amd64.zip", "not:debug"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SubstituteTemplateVarsSlice() = %#v, want %#v", got, want)
	}
}

func TestSubstituteTemplateVarsSliceHandlesEmptyAndNil(t *testing.T) {
	if got := SubstituteTemplateVarsSlice(nil, "linux/amd64"); got != nil {
		t.Fatalf("expected nil for nil input, got %#v", got)
	}
	if got := SubstituteTemplateVarsSlice([]string{}, "linux/amd64"); len(got) != 0 {
		t.Fatalf("expected empty slice for empty input, got %#v", got)
	}
}
