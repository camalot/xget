package installed

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadMissingReturnsEmptyStore(t *testing.T) {
	store, err := Load(filepath.Join(t.TempDir(), ".xget.installed.yml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if store == nil || len(store.Packages) != 0 {
		t.Fatalf("expected empty store, got %#v", store)
	}
}

func TestUpsertCreatesAndUpdatesPackageRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".config", "xget", ".xget.installed.yml")
	installedAt := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	first := Package{
		Name:           "owner/repo",
		InstalledAt:    installedAt,
		RefreshedAt:    installedAt,
		DownloadURL:    "https://example.com/repo.zip",
		Asset:          "repo.zip",
		ExtractedFiles: []string{"/tmp/repo"},
		InstalledTag:   "v1.0.0",
		CurrentTag:     "v1.0.0",
		Source:         "GitHub",
		SHA256:         "abc123",
	}
	if err := Upsert(path, first); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	second := first
	second.InstalledAt = installedAt.Add(time.Hour)
	second.RefreshedAt = installedAt.Add(time.Hour)
	second.DownloadURL = "https://example.com/repo-v2.zip"
	second.Asset = "repo-v2.zip"
	second.InstalledTag = "v2.0.0"
	second.CurrentTag = "v2.0.0"
	if err := Upsert(path, second); err != nil {
		t.Fatalf("second Upsert returned error: %v", err)
	}

	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(store.Packages) != 1 {
		t.Fatalf("expected one package, got %d", len(store.Packages))
	}
	got := store.Packages["github:owner/repo"]
	if got.Asset != "repo-v2.zip" || got.InstalledTag != "v2.0.0" {
		t.Fatalf("record was not updated: %#v", got)
	}
	if !got.InstalledAt.Equal(installedAt.Add(time.Hour)) {
		t.Fatalf("expected InstalledAt to update, got %s", got.InstalledAt)
	}
	if !got.RefreshedAt.Equal(installedAt.Add(time.Hour)) {
		t.Fatalf("expected RefreshedAt to update, got %s", got.RefreshedAt)
	}

	// #nosec G304 -- test reads a temporary file path created by the test.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read installed store: %v", err)
	}
	if !strings.Contains(string(data), `"github:owner/repo":`) {
		t.Fatalf("expected source-qualified package key to be quoted, got:\n%s", data)
	}
}

func TestLoadMigratesNameKeyedRepoRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.installed.yml")
	content := `packages:
  act:
    name: act
    repo: nektos/act
    installed_at: 2026-09-01T17:10:28Z
    download_url: https://example.com/act.zip
    asset: act.zip
    refreshed_at: 2026-09-01T17:10:28Z
    current_version: v0.2.89
    current_tag: v0.2.89
    installed_version: v0.2.89
    installed_tag: v0.2.89
    source: GitHub
    sha256: abc123
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(store.Packages) != 1 {
		t.Fatalf("expected one package, got %d", len(store.Packages))
	}
	got, ok := store.Packages["github:nektos/act"]
	if !ok {
		t.Fatalf("expected migrated github key, got %#v", store.Packages)
	}
	if got.Name != "nektos/act" || got.Repo != "" || got.CurrentTag != "v0.2.89" || got.InstalledTag != "v0.2.89" {
		t.Fatalf("unexpected migrated package: %#v", got)
	}
}
