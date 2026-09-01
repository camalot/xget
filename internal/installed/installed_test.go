package installed

import (
	"path/filepath"
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
		Name:             "repo",
		Repo:             "owner/repo",
		InstalledAt:      installedAt,
		RefreshedAt:      installedAt,
		DownloadURL:      "https://example.com/repo.zip",
		Asset:            "repo.zip",
		ExtractedFiles:   []string{"/tmp/repo"},
		InstalledTag:     "v1.0.0",
		InstalledVersion: "v1.0.0",
		CurrentTag:       "v1.0.0",
		CurrentVersion:   "v1.0.0",
		Source:           "GitHub",
		SHA256:           "abc123",
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
	got := store.Packages["repo"]
	if got.Asset != "repo-v2.zip" || got.InstalledTag != "v2.0.0" {
		t.Fatalf("record was not updated: %#v", got)
	}
	if !got.InstalledAt.Equal(installedAt) {
		t.Fatalf("expected InstalledAt to be preserved, got %s", got.InstalledAt)
	}
	if !got.RefreshedAt.Equal(installedAt.Add(time.Hour)) {
		t.Fatalf("expected RefreshedAt to update, got %s", got.RefreshedAt)
	}
}
