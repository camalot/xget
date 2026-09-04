package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
	"github.com/camalot/xget/internal/installed"
	"github.com/spf13/cobra"
)

func TestPrintInstalledPackagesUsesTableFormat(t *testing.T) {
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	printInstalledPackages(cmd, []installed.Package{
		{
			Name:            "nektos/act",
			InstallLocation: "~/.local/bin",
			InstalledAt:     time.Date(2026, 9, 1, 17, 10, 28, 0, time.UTC),
			RefreshedAt:     time.Date(2026, 9, 1, 17, 10, 28, 0, time.UTC),
			InstalledTag:    "v0.2.89",
			CurrentTag:      "v0.2.89",
			Source:          "GitHub",
		},
	})

	got := buf.String()
	if !strings.Contains(got, "PACKAGE") || !strings.Contains(got, "TAG/VERSION") || !strings.Contains(got, "LATEST") || !strings.Contains(got, "LOCATION") || !strings.Contains(got, "INSTALLED/UPDATED") {
		t.Fatalf("expected table header, got:\n%s", got)
	}
	if !strings.Contains(got, "github:nektos/act") || !strings.Contains(got, "v0.2.89") || !strings.Contains(got, "~/.local/bin") || !strings.Contains(got, "2026-09-01") {
		t.Fatalf("expected installed package row, got:\n%s", got)
	}
	if strings.Contains(got, "Installed at:") || strings.Contains(got, "Download URL:") {
		t.Fatalf("expected summary table, got detail output:\n%s", got)
	}
}

func TestListTargetPrintsReleaseSummariesAndPassesPrereleaseFlag(t *testing.T) {
	original := listReleases
	t.Cleanup(func() { listReleases = original })
	includePrereleases := false
	listReleases = func(target string, include bool) ([]engine.Release, error) {
		if target != "eza-community/eza" {
			t.Errorf("target = %q", target)
		}
		includePrereleases = include
		return []engine.Release{{Name: "eza v0.23.5", Tag: "v0.23.5", PublishedAt: time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)}}, nil
	}

	out, err := runCLI(t, "list", "eza-community/eza", "--pre-release")
	if err != nil {
		t.Fatal(err)
	}
	if !includePrereleases {
		t.Fatal("expected --pre-release to be passed to release lookup")
	}
	if !strings.Contains(collapseSpaces(out), "NAME TAG DATE") || !strings.Contains(collapseSpaces(out), "eza v0.23.5 v0.23.5 2026-09-04") {
		t.Fatalf("expected release table, got:\n%s", out)
	}
}

func TestPrintInstalledPackagesAlignsHeaderWithRows(t *testing.T) {
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	printInstalledPackages(cmd, []installed.Package{
		{Name: "bschaatsbergen/cidr", InstalledTag: "v2.3.0", CurrentTag: "v2.3.0", Source: "GitHub", InstallLocation: "/bin"},
		{Name: "a/b", InstalledTag: "v1.0.0", CurrentTag: "v1.1.0", Source: "GitHub", InstallLocation: "/bin"},
	})

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected header, separator, and 2 rows, got:\n%s", buf.String())
	}
	header, separator := lines[0], lines[1]
	if len(separator) != len(header) {
		t.Fatalf("separator width %d != header width %d:\n%s", len(separator), len(header), buf.String())
	}
	if strings.Trim(separator, "-") != "" {
		t.Fatalf("separator = %q", separator)
	}
	// Every row must start its version column at the header's column offset.
	column := strings.Index(header, "TAG/VERSION")
	for _, line := range lines[2:] {
		if !strings.HasPrefix(line[column:], "v") {
			t.Fatalf("column misaligned in %q (expected a version at offset %d)\n%s", line, column, buf.String())
		}
	}
}

func TestInstalledLocationUsesDirectoryForLegacyFileRecord(t *testing.T) {
	pkg := installed.Package{
		InstallLocation: filepath.Join("home", "user", ".local", "bin", "eza.exe"),
		ExtractedFiles:  []string{filepath.Join("home", "user", ".local", "bin", "eza.exe")},
	}
	if got, want := installedLocation(pkg), filepath.Join("home", "user", ".local", "bin"); got != want {
		t.Fatalf("installedLocation() = %q, want %q", got, want)
	}
}

func TestListInstalledRefreshesAndPersistsLatestTag(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("bschaatsbergen/cidr", "v2.2.0"))
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})

	out, err := runCLI(t, "list", "--installed")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(collapseSpaces(out), "github:bschaatsbergen/cidr v2.2.0 v2.3.0") {
		t.Fatalf("expected refreshed latest tag in output:\n%s", out)
	}

	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := store.Packages["github:bschaatsbergen/cidr"].CurrentTag; got != "v2.3.0" {
		t.Fatalf("current_tag = %q, want v2.3.0", got)
	}
}

func runListConfigured(t *testing.T, cfg *config.Config) string {
	t.Helper()
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	if err := listConfigured(cmd, cfg); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestListConfiguredWithoutConfigFilePointsAtInstalled(t *testing.T) {
	got := runListConfigured(t, config.Default())
	if !strings.Contains(got, "no config file found") {
		t.Fatalf("expected missing-config message, got %q", got)
	}
	if !strings.Contains(got, "xget list --installed") {
		t.Fatalf("expected a pointer to --installed, got %q", got)
	}
}

func TestListConfiguredWithEmptyConfigFileNamesThePath(t *testing.T) {
	cfg := config.Default()
	cfg.Path = "/home/user/.config/xget/.xget.yml"

	got := runListConfigured(t, cfg)
	if !strings.Contains(got, cfg.Path) {
		t.Fatalf("expected the config path in the message, got %q", got)
	}
	if !strings.Contains(got, "xget list --installed") {
		t.Fatalf("expected a pointer to --installed, got %q", got)
	}
}

func TestListConfiguredPrintsSortedRepositories(t *testing.T) {
	cfg := config.Default()
	cfg.Path = "/tmp/.xget.yml"
	cfg.Repositories["zyedidia/micro"] = config.Repository{Name: "zyedidia/micro"}
	cfg.Repositories["a/b"] = config.Repository{Name: "a/b"}

	got := runListConfigured(t, cfg)
	if got != "a/b\nzyedidia/micro\n" {
		t.Fatalf("output = %q", got)
	}
}
