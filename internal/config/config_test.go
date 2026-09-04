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
	baseNames := []string{".xget", ".eget"}
	exts := []string{"toml", "yml", "yaml"}

	for _, base := range baseNames {
		for _, ext := range exts {
			want = append(want,
				filepath.Join(".", base+"."+ext),
				filepath.Join(homePath, base+"."+ext),
				filepath.Join(homePath, ".config", "xget", base+"."+ext),
			)
			if runtime.GOOS == "windows" {
				localAppData := os.Getenv("LOCALAPPDATA")
				if localAppData == "" {
					localAppData = filepath.Join(homePath, "AppData", "Local")
				}
				want = append(want, filepath.Join(localAppData, "xget", base+"."+ext))
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
