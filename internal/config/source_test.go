package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultSources(t *testing.T) {
	cfg := Default()

	github, err := cfg.ResolveSource("")
	if err != nil {
		t.Fatal(err)
	}
	if github.Type != "github" || github.Host != "github.com" || github.APIURL != "https://api.github.com" {
		t.Fatalf("unexpected GitHub defaults: %#v", github)
	}
	if strings.Join(github.TokenEnv, ",") != "XGET_GITHUB_TOKEN,GITHUB_TOKEN,EGET_GITHUB_TOKEN" {
		t.Fatalf("unexpected GitHub token environment order: %#v", github.TokenEnv)
	}

	gitlab, err := cfg.ResolveSource("gitlab")
	if err != nil {
		t.Fatal(err)
	}
	if gitlab.Type != "gitlab" || gitlab.Host != "gitlab.com" || gitlab.APIURL != "https://gitlab.com/api/v4" {
		t.Fatalf("unexpected GitLab defaults: %#v", gitlab)
	}
}

func TestLoadNamedSourcesAndDoesNotTreatSourcesAsRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.toml")
	content := `[global]
source = "work"

[sources.work]
type = "github"
host = "github.example.com"
token_env = ["WORK_GITHUB_TOKEN"]
disable_token_warning = true

["group/project"]
source = "work"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := cfg.Repositories["sources"]; exists {
		t.Fatal("sources must not be parsed as a repository")
	}
	work, err := cfg.ResolveSource("work")
	if err != nil {
		t.Fatal(err)
	}
	if work.Type != "github" || work.Host != "github.example.com" || work.APIURL != "https://github.example.com/api/v3" {
		t.Fatalf("unexpected work source: %#v", work)
	}
	if strings.Join(work.TokenEnv, ",") != "WORK_GITHUB_TOKEN" || !work.DisableTokenWarning {
		t.Fatalf("unexpected work token settings: %#v", work)
	}
	if cfg.Repositories["group/project"].SourceType != "work" {
		t.Fatalf("repository source = %q", cfg.Repositories["group/project"].SourceType)
	}
}

func TestLoadRejectsUnsupportedSourceType(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	if err := os.WriteFile(path, []byte("sources:\n  custom:\n    type: bitbucket\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("Load() error = %v, want unsupported type", err)
	}
}

func TestLoadNamedSourceFromYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.yml")
	content := `global:
  source: company
sources:
  company:
    type: gitlab
    host: gitlab.example.com
    token_env:
      - COMPANY_GITLAB_TOKEN
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	source, err := cfg.ResolveSource(cfg.Global.SourceType)
	if err != nil {
		t.Fatal(err)
	}
	if source.Type != "gitlab" || source.APIURL != "https://gitlab.example.com/api/v4" {
		t.Fatalf("unexpected YAML source: %#v", source)
	}
}

func TestResolveSourceIsCaseInsensitive(t *testing.T) {
	source, err := Default().ResolveSource("GitHub")
	if err != nil {
		t.Fatal(err)
	}
	if source.Name != "github" {
		t.Fatalf("source name = %q, want github", source.Name)
	}
}

func TestWarnStoredTokensIncludesSuppressionInstructions(t *testing.T) {
	cfg := Default()
	cfg.Global.GithubToken = "legacy"
	cfg.Sources["work"] = Source{Name: "work", Type: "gitlab", Token: "secret"}
	output := &bytes.Buffer{}

	warnStoredTokens(cfg, output)

	warning := output.String()
	for _, expected := range []string{
		"sources.github.disable_token_warning to true",
		"sources.work.disable_token_warning to true",
	} {
		if !strings.Contains(warning, expected) {
			t.Fatalf("warning %q does not contain %q", warning, expected)
		}
	}

	github := cfg.Sources["github"]
	github.DisableTokenWarning = true
	cfg.Sources["github"] = github
	work := cfg.Sources["work"]
	work.DisableTokenWarning = true
	cfg.Sources["work"] = work
	output.Reset()
	warnStoredTokens(cfg, output)
	if output.Len() != 0 {
		t.Fatalf("warnings were not disabled: %q", output.String())
	}
}
