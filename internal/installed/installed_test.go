package installed

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestSelfUpdateCheckRoundTripsOutsidePackages(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.installed.yml")
	checked := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	store := &Store{Packages: map[string][]Package{}, SelfUpdate: SelfUpdateCheck{LastChecked: checked}}
	if err := Save(path, store); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.SelfUpdate.LastChecked.Equal(checked) || len(loaded.Packages) != 0 {
		t.Fatalf("loaded = %#v", loaded)
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
	records := store.Packages["github:owner/repo"]
	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
	got := records[0]
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
	records, ok := store.Packages["github:nektos/act"]
	if !ok || len(records) != 1 {
		t.Fatalf("expected migrated github key, got %#v", store.Packages)
	}
	got := records[0]
	if got.Name != "nektos/act" || got.Repo != "" || got.CurrentTag != "v0.2.89" || got.InstalledTag != "v0.2.89" {
		t.Fatalf("unexpected migrated package: %#v", got)
	}
}

func multiLocationPackage(location, tag string) Package {
	return Package{
		Name:            "jgm/pandoc",
		InstallLocation: location,
		InstalledTag:    tag,
		CurrentTag:      tag,
		Source:          "GitHub",
	}
}

func TestUpsertTracksEachInstallLocationSeparately(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.installed.yml")

	if err := Upsert(path, multiLocationPackage(filepath.FromSlash("/mnt/test/bin"), "3.10")); err != nil {
		t.Fatal(err)
	}
	if err := Upsert(path, multiLocationPackage(filepath.FromSlash("/opt/local/bin"), "3.10")); err != nil {
		t.Fatal(err)
	}

	store, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	records := store.Packages["github:jgm/pandoc"]
	if len(records) != 2 {
		t.Fatalf("expected both locations to be tracked, got %#v", records)
	}

	// Reinstalling one location must not disturb the other.
	if err := Upsert(path, multiLocationPackage(filepath.FromSlash("/opt/local/bin"), "3.11")); err != nil {
		t.Fatal(err)
	}
	store, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	records = store.Packages["github:jgm/pandoc"]
	if len(records) != 2 {
		t.Fatalf("expected two records, got %#v", records)
	}
	for _, pkg := range records {
		want := "3.10"
		if SamePath(pkg.InstallLocation, filepath.FromSlash("/opt/local/bin")) {
			want = "3.11"
		}
		if pkg.InstalledTag != want {
			t.Fatalf("%s installed_tag = %q, want %q", pkg.InstallLocation, pkg.InstalledTag, want)
		}
	}
}

func TestRemoveDropsOnlyTheMatchingLocation(t *testing.T) {
	store := &Store{Packages: map[string][]Package{}}
	first := multiLocationPackage(filepath.FromSlash("/mnt/test/bin"), "3.10")
	second := multiLocationPackage(filepath.FromSlash("/opt/local/bin"), "3.10")
	store.Set(first)
	store.Set(second)

	store.Remove(first)
	records := store.Packages["github:jgm/pandoc"]
	if len(records) != 1 || !SamePath(records[0].InstallLocation, second.InstallLocation) {
		t.Fatalf("records = %#v, want only %s", records, second.InstallLocation)
	}

	store.Remove(second)
	if _, ok := store.Packages["github:jgm/pandoc"]; ok {
		t.Fatalf("expected the key to be dropped, got %#v", store.Packages)
	}
}

func TestLoadReadsSingleRecordAndListLayouts(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".xget.installed.yml")
	content := `packages:
  "github:jgm/pandoc":
    name: jgm/pandoc
    install_location: /mnt/test/bin
    installed_tag: "3.10"
    current_tag: "3.10"
    source: GitHub
  "github:nektos/act":
    - name: nektos/act
      install_location: /opt/local/bin
      installed_tag: v0.2.89
      current_tag: v0.2.89
      source: GitHub
    - name: nektos/act
      install_location: /usr/local/bin
      installed_tag: v0.2.88
      current_tag: v0.2.89
      source: GitHub
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if got := store.Packages["github:jgm/pandoc"]; len(got) != 1 || got[0].InstalledTag != "3.10" {
		t.Fatalf("single record was not migrated to a list: %#v", got)
	}
	if got := store.Packages["github:nektos/act"]; len(got) != 2 {
		t.Fatalf("expected both act records, got %#v", got)
	}
}

func TestSortedPackagesReturnsEveryLocation(t *testing.T) {
	store := &Store{Packages: map[string][]Package{}}
	store.Set(multiLocationPackage(filepath.FromSlash("/opt/local/bin"), "3.10"))
	store.Set(multiLocationPackage(filepath.FromSlash("/mnt/test/bin"), "3.10"))

	packages := SortedPackages(store)
	if len(packages) != 2 {
		t.Fatalf("expected two entries, got %#v", packages)
	}

	if packages[0].InstallLocation != filepath.FromSlash("/mnt/test/bin") {
		t.Fatalf("expected locations to be sorted, got %#v", packages)
	}
}

func TestLoadMergesLegacyFileAndDirectoryLocations(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bin")
	file := filepath.Join(dir, "fx.exe")
	old := Package{
		Name: "antonmedv/fx", Source: "GitHub",
		InstallLocation: file, ExtractedFiles: []string{filepath.ToSlash(file)},
		InstalledTag: "39.2.0", InstalledAt: time.Unix(1, 0),
	}
	latest := old
	latest.Source = "github"
	latest.InstallLocation = dir
	latest.InstalledTag = "40.0.0"
	latest.InstalledAt = time.Unix(2, 0)
	other := latest
	other.InstallLocation = filepath.Join(t.TempDir(), "other-bin")
	other.ExtractedFiles = []string{filepath.Join(other.InstallLocation, "fx.exe")}

	for _, records := range [][]Package{{old, latest, other}, {latest, other, old}} {
		path := filepath.Join(t.TempDir(), "installed.yml")
		store := &Store{Packages: map[string][]Package{"github:antonmedv/fx": records}}
		if err := Save(path, store); err != nil {
			t.Fatal(err)
		}
		loaded, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(loaded.Packages[old.Key()]); got != 2 {
			t.Fatalf("record count = %d, want two distinct locations", got)
		}
		got, ok := loaded.Find(old.Key(), dir)
		if !ok || got.InstalledTag != latest.InstalledTag {
			t.Fatalf("merged record = %#v, found = %t", got, ok)
		}
		if err := Save(path, loaded); err != nil {
			t.Fatal(err)
		}
		reloaded, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(loaded, reloaded) {
			t.Fatal("migration did not survive save and reload")
		}
	}
}

func TestSetAndRemoveNormalizeLegacyFileLocation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bin")
	file := filepath.Join(dir, "fx.exe")
	pkg := Package{
		Name: "antonmedv/fx", Source: "GitHub",
		InstallLocation: file, ExtractedFiles: []string{file},
	}
	store := &Store{}
	store.Set(pkg)
	if _, ok := store.Find(pkg.Key(), dir); !ok {
		t.Fatalf("expected directory location, got %#v", store.Packages)
	}
	updated := pkg
	updated.InstallLocation = dir
	updated.InstalledTag = "40.0.0"
	store.Set(updated)
	if len(store.Packages[pkg.Key()]) != 1 {
		t.Fatal("reinstall added a duplicate location")
	}
	store.Remove(pkg)
	if len(store.Packages) != 0 {
		t.Fatalf("legacy location was not removed: %#v", store.Packages)
	}
}
