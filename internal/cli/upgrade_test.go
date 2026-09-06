package cli

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/camalot/xget/internal/installed"
	"github.com/camalot/xget/internal/options"
)

func useTempInstalledStore(t *testing.T, packages ...installed.Package) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")

	storePath := filepath.Join(home, ".config", "xget", ".xget.installed.yml")
	store := &installed.Store{Packages: map[string][]installed.Package{}}
	for _, pkg := range packages {
		store.Set(pkg)
	}
	if err := installed.Save(storePath, store); err != nil {
		t.Fatal(err)
	}
	return storePath
}

func storedPackages(t *testing.T, storePath, key string) []installed.Package {
	t.Helper()
	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	return store.Packages[key]
}

func storedPackage(t *testing.T, storePath, key string) installed.Package {
	t.Helper()
	records := storedPackages(t, storePath, key)
	if len(records) != 1 {
		t.Fatalf("expected one record for %s, got %d", key, len(records))
	}
	return records[0]
}

func writeUpgradeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".xget.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func stubRefresh(t *testing.T, latest map[string]string) *[]options.Flags {
	t.Helper()
	original := refreshPackage
	t.Cleanup(func() { refreshPackage = original })

	seen := &[]options.Flags{}
	refreshPackage = func(pkg installed.Package, opts options.Flags) (installed.Package, error) {
		*seen = append(*seen, opts)
		if tag, ok := latest[pkg.Name]; ok {
			pkg.CurrentTag = tag
		}
		pkg.RefreshedAt = time.Now().UTC()
		return pkg, nil
	}
	return seen
}

func collapseSpaces(s string) string {
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// stubEngine records upgrade calls and mimics the store update engine.Run performs.
func stubEngine(t *testing.T, storePath string, fail map[string]error) *[]struct {
	Target string
	Opts   options.Flags
} {
	t.Helper()
	original := runEngine
	t.Cleanup(func() { runEngine = original })

	calls := &[]struct {
		Target string
		Opts   options.Flags
	}{}
	runEngine = func(target string, opts options.Flags) error {
		*calls = append(*calls, struct {
			Target string
			Opts   options.Flags
		}{target, opts})
		if err, ok := fail[target]; ok {
			return err
		}
		store, err := installed.Load(storePath)
		if err != nil {
			return err
		}
		for _, records := range store.Packages {
			for index, pkg := range records {
				if pkg.Name != target {
					continue
				}
				if opts.Output != "" && !installed.SamePath(pkg.InstallLocation, opts.Output) {
					continue
				}
				pkg.InstalledTag = opts.Tag
				pkg.CurrentTag = opts.Tag
				pkg.Options = installed.Options{
					Tag:          opts.Tag,
					Prerelease:   opts.Prerelease,
					Output:       opts.Output,
					ExtractFile:  opts.ExtractFile,
					All:          opts.All,
					DownloadOnly: opts.DLOnly,
					UpgradeOnly:  opts.UpgradeOnly,
					Asset:        opts.Asset,
					Ignore:       opts.Ignore,
					Verify:       opts.Verify,
				}
				records[index] = pkg
			}
		}
		return installed.Save(storePath, store)
	}
	return calls
}

func samplePackage(name, installedTag string) installed.Package {
	return installed.Package{
		Name:            name,
		InstallLocation: "/home/user/.local/bin",
		InstalledTag:    installedTag,
		CurrentTag:      installedTag,
		Source:          "GitHub",
		Options: installed.Options{
			Output:      "/home/user/.local/bin",
			ExtractFile: "*.exe",
			Asset:       []string{".tar.gz"},
			Ignore:      []string{".sbom.json"},
		},
	}
}

func samplePackageAt(name, installedTag, location string) installed.Package {
	pkg := samplePackage(name, installedTag)
	pkg.InstallLocation = location
	pkg.Options.Output = location
	return pkg
}

func upgradeOutputs(calls []struct {
	Target string
	Opts   options.Flags
}) []string {
	outputs := make([]string, 0, len(calls))
	for _, call := range calls {
		outputs = append(outputs, call.Opts.Output)
	}
	sort.Strings(outputs)
	return outputs
}

func TestUpgradeAllUpgradesEveryOutOfDateLocation(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.10", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("engine calls = %+v, want one per location", *calls)
	}
	if got := upgradeOutputs(*calls); !reflect.DeepEqual(got, []string{"/mnt/test/bin", "/opt/local/bin"}) {
		t.Fatalf("outputs = %#v", got)
	}
	if !strings.Contains(out, "in /mnt/test/bin") || !strings.Contains(out, "in /opt/local/bin") {
		t.Fatalf("expected a progress line per location:\n%s", out)
	}

	for _, pkg := range storedPackages(t, storePath, "github:jgm/pandoc") {
		if pkg.InstalledTag != "3.11" {
			t.Fatalf("%s installed_tag = %q, want 3.11", pkg.InstallLocation, pkg.InstalledTag)
		}
	}
}

func TestUpgradeListsEveryInstallLocation(t *testing.T) {
	useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.10", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})

	out, err := runCLI(t, "upgrade", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	collapsed := collapseSpaces(out)
	if !strings.Contains(collapsed, "jgm/pandoc 3.10 3.11 /mnt/test/bin GitHub") ||
		!strings.Contains(collapsed, "jgm/pandoc 3.10 3.11 /opt/local/bin GitHub") {
		t.Fatalf("expected one row per location:\n%s", out)
	}
	if !strings.Contains(out, "2 upgrades available.") {
		t.Fatalf("expected both locations counted:\n%s", out)
	}
}

func TestUpgradeNamedUpgradesEveryLocation(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.10", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "jgm/pandoc"); err != nil {
		t.Fatal(err)
	}
	if got := upgradeOutputs(*calls); !reflect.DeepEqual(got, []string{"/mnt/test/bin", "/opt/local/bin"}) {
		t.Fatalf("outputs = %#v", got)
	}
}

func TestUpgradeToSelectsASingleLocation(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.10", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "jgm/pandoc", "--to", "/opt/local/bin"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].Opts.Output != "/opt/local/bin" {
		t.Fatalf("calls = %+v", *calls)
	}
	for _, pkg := range storedPackages(t, storePath, "github:jgm/pandoc") {
		want := "3.10"
		if pkg.InstallLocation == "/opt/local/bin" {
			want = "3.11"
		}
		if pkg.InstalledTag != want {
			t.Fatalf("%s installed_tag = %q, want %q", pkg.InstallLocation, pkg.InstalledTag, want)
		}
	}
}

func TestUpgradeAllHonorsToFilter(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.10", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "--all", "--to", "/mnt/test/bin"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].Opts.Output != "/mnt/test/bin" {
		t.Fatalf("calls = %+v", *calls)
	}
}

func TestUpgradeToUntrackedLocationErrors(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackageAt("jgm/pandoc", "3.10", "/mnt/test/bin"))
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	_, err := runCLI(t, "upgrade", "jgm/pandoc", "--to", "/nowhere/bin")
	if err == nil || err.Error() != "package jgm/pandoc is not installed to /nowhere/bin" {
		t.Fatalf("error = %v", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("unexpected calls: %+v", *calls)
	}
}

func TestUpgradeInlineTagInstallsRequestedVersion(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackageAt("jgm/pandoc", "3.11", "/mnt/test/bin"))
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "jgm/pandoc@3.10")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].Opts.Tag != "3.10" {
		t.Fatalf("calls = %+v, want a 3.10 install", *calls)
	}
	if !strings.Contains(out, "Installing jgm/pandoc 3.10 in /mnt/test/bin") {
		t.Fatalf("output:\n%s", out)
	}
	if got := storedPackage(t, storePath, "github:jgm/pandoc").InstalledTag; got != "3.10" {
		t.Fatalf("installed_tag = %q, want 3.10", got)
	}
}

func TestUpgradeTagFlagInstallsRequestedVersionInEveryLocation(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.11", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.11", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "jgm/pandoc", "--tag", "3.10"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 2 {
		t.Fatalf("calls = %+v", *calls)
	}
	for _, call := range *calls {
		if call.Opts.Tag != "3.10" {
			t.Fatalf("tag = %q, want 3.10", call.Opts.Tag)
		}
	}
}

func TestUpgradeTagWithoutPackageErrors(t *testing.T) {
	useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, nil)

	if _, err := runCLI(t, "upgrade", "--tag", "v1.0.0"); err == nil || !strings.Contains(err.Error(), "--tag requires a package name") {
		t.Fatalf("error = %v", err)
	}
}

func TestUpgradeNamedReportsEachUpToDateLocation(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackageAt("jgm/pandoc", "3.11", "/mnt/test/bin"),
		samplePackageAt("jgm/pandoc", "3.11", "/opt/local/bin"),
	)
	stubRefresh(t, map[string]string{"jgm/pandoc": "3.11"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "jgm/pandoc")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("unexpected calls: %+v", *calls)
	}
	if !strings.Contains(out, "jgm/pandoc is already up to date (3.11) in /mnt/test/bin.") ||
		!strings.Contains(out, "jgm/pandoc is already up to date (3.11) in /opt/local/bin.") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestRootCommandIncludesUpgradeSubcommand(t *testing.T) {
	cmd := newRootCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "upgrade" {
			if sub.Flags().Lookup("all") == nil {
				t.Fatal("upgrade command missing --all flag")
			}
			return
		}
	}
	t.Fatal("upgrade subcommand not registered")
}

func TestUpgradeListsAvailableUpgrades(t *testing.T) {
	useTempInstalledStore(t,
		samplePackage("bschaatsbergen/cidr", "v2.2.0"),
		samplePackage("camalot/xget", "v2.0.0-beta"),
	)
	stubRefresh(t, map[string]string{
		"bschaatsbergen/cidr": "v2.3.0",
		"camalot/xget":        "v2.0.0-beta",
	})

	out, err := runCLI(t, "upgrade", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Available") || !strings.Contains(out, "Location") || !strings.Contains(out, "Source") {
		t.Fatalf("missing headers:\n%s", out)
	}
	if !strings.Contains(collapseSpaces(out), "bschaatsbergen/cidr v2.2.0 v2.3.0 /home/user/.local/bin GitHub") {
		t.Fatalf("missing upgrade row:\n%s", out)
	}
	if strings.Contains(out, "camalot/xget") {
		t.Fatalf("up-to-date package should not be listed:\n%s", out)
	}
	if !strings.Contains(out, "1 upgrade available.") {
		t.Fatalf("missing count line:\n%s", out)
	}
}

func TestUpgradeColorsAvailableUpgradeUnlessNoColor(t *testing.T) {
	useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, map[string]string{"a/one": "v1.1.0"})

	colored, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(colored, "\x1b[33ma/one") {
		t.Fatalf("expected yellow upgrade row, got:\n%s", colored)
	}

	plain, err := runCLI(t, "upgrade", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("expected no ANSI color codes, got:\n%s", plain)
	}
}

func TestUpgradePluralizesCount(t *testing.T) {
	useTempInstalledStore(t,
		samplePackage("a/one", "v1.0.0"),
		samplePackage("b/two", "v1.0.0"),
	)
	stubRefresh(t, map[string]string{"a/one": "v1.1.0", "b/two": "v2.0.0"})

	out, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2 upgrades available.") {
		t.Fatalf("expected plural count:\n%s", out)
	}
}

func TestUpgradeRefreshesAndPersistsLatestTag(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("bschaatsbergen/cidr", "v2.2.0"))
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})

	if _, err := runCLI(t, "upgrade"); err != nil {
		t.Fatal(err)
	}

	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	pkg := store.Packages["github:bschaatsbergen/cidr"][0]
	if pkg.CurrentTag != "v2.3.0" {
		t.Fatalf("current_tag = %q, want v2.3.0", pkg.CurrentTag)
	}
	if pkg.InstalledTag != "v2.2.0" {
		t.Fatalf("installed_tag = %q, want v2.2.0", pkg.InstalledTag)
	}
}

func TestUpgradeReportsNoUpgrades(t *testing.T) {
	useTempInstalledStore(t, samplePackage("camalot/xget", "v2.0.0"))
	stubRefresh(t, map[string]string{"camalot/xget": "v2.0.0"})

	out, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "No available upgrades." {
		t.Fatalf("output = %q", out)
	}
}

func TestUpgradeReportsEmptyStore(t *testing.T) {
	useTempInstalledStore(t)
	stubRefresh(t, nil)

	out, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "No available upgrades." {
		t.Fatalf("output = %q", out)
	}
}

func TestUpgradeSkipsNonGitHubSources(t *testing.T) {
	pkg := samplePackage("some-tool", "v1.0.0")
	pkg.Source = "URL"
	pkg.CurrentTag = "v2.0.0"
	useTempInstalledStore(t, pkg)
	stubRefresh(t, nil)

	out, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "No available upgrades." {
		t.Fatalf("output = %q", out)
	}
}

func TestUpgradeListsPinnedPackagesSeparately(t *testing.T) {
	pinned := samplePackage("bschaatsbergen/cidr", "v2.2.0")
	pinned.Options.Tag = "v2.2.0"
	useTempInstalledStore(t, pinned, samplePackage("camalot/xget", "v1.0.0"))
	stubRefresh(t, map[string]string{
		"bschaatsbergen/cidr": "v2.3.0",
		"camalot/xget":        "v1.1.0",
	})

	out, err := runCLI(t, "upgrade")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1 upgrade available.") {
		t.Fatalf("pinned package must not be counted:\n%s", out)
	}
	if !strings.Contains(out, "require explicit targeting for upgrade:") {
		t.Fatalf("missing pinned section:\n%s", out)
	}
	pinnedIdx := strings.Index(out, "require explicit targeting")
	if strings.Index(out, "bschaatsbergen/cidr") < pinnedIdx {
		t.Fatalf("pinned package listed in the main table:\n%s", out)
	}
}

func TestUpgradeAllRunsEngineWithStoredOptions(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("bschaatsbergen/cidr", "v2.2.0"))
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Upgrading bschaatsbergen/cidr from v2.2.0 to v2.3.0") {
		t.Fatalf("missing progress line:\n%s", out)
	}
	if len(*calls) != 1 {
		t.Fatalf("engine calls = %d", len(*calls))
	}
	call := (*calls)[0]
	if call.Target != "bschaatsbergen/cidr" {
		t.Fatalf("target = %q", call.Target)
	}
	if call.Opts.Tag != "v2.3.0" {
		t.Fatalf("tag = %q, want v2.3.0", call.Opts.Tag)
	}
	if call.Opts.UpgradeOnly {
		t.Fatal("upgrade_only must not be forwarded")
	}
	if call.Opts.Output != "/home/user/.local/bin" {
		t.Fatalf("output = %q", call.Opts.Output)
	}
	if call.Opts.ExtractFile != "*.exe" || call.Opts.SourceType != "GitHub" {
		t.Fatalf("options = %+v", call.Opts)
	}
	if len(call.Opts.Asset) != 1 || call.Opts.Asset[0] != ".tar.gz" {
		t.Fatalf("asset = %v", call.Opts.Asset)
	}
	if len(call.Opts.Ignore) != 1 || call.Opts.Ignore[0] != ".sbom.json" {
		t.Fatalf("ignore = %v", call.Opts.Ignore)
	}
}

func TestUpgradeAllLeavesUnpinnedPackagesUnpinned(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("bschaatsbergen/cidr", "v2.2.0"))
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})
	stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "--all"); err != nil {
		t.Fatal(err)
	}

	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	pkg := store.Packages["github:bschaatsbergen/cidr"][0]
	if pkg.Options.Tag != "" {
		t.Fatalf("options.tag = %q, want empty", pkg.Options.Tag)
	}
	if pkg.InstalledTag != "v2.3.0" {
		t.Fatalf("installed_tag = %q", pkg.InstalledTag)
	}
}

func TestUpgradeAllUsesInstallLocationWhenOutputIsUnset(t *testing.T) {
	pkg := samplePackage("a/one", "v1.0.0")
	pkg.Options.Output = ""
	pkg.InstallLocation = "/opt/bin"
	storePath := useTempInstalledStore(t, pkg)
	stubRefresh(t, map[string]string{"a/one": "v1.1.0"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "--all"); err != nil {
		t.Fatal(err)
	}
	if (*calls)[0].Opts.Output != "/opt/bin" {
		t.Fatalf("output = %q", (*calls)[0].Opts.Output)
	}
}

func TestUpgradeAllSkipsPinnedAndListsThem(t *testing.T) {
	pinned := samplePackage("bschaatsbergen/cidr", "v2.2.0")
	pinned.Options.Tag = "v2.2.0"
	storePath := useTempInstalledStore(t, pinned, samplePackage("camalot/xget", "v1.0.0"))
	stubRefresh(t, map[string]string{
		"bschaatsbergen/cidr": "v2.3.0",
		"camalot/xget":        "v1.1.0",
	})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].Target != "camalot/xget" {
		t.Fatalf("calls = %+v", *calls)
	}
	if !strings.Contains(out, "require explicit targeting for upgrade:") {
		t.Fatalf("missing pinned notice:\n%s", out)
	}
	if !strings.Contains(collapseSpaces(out), "bschaatsbergen/cidr v2.2.0 v2.3.0 /home/user/.local/bin GitHub") {
		t.Fatalf("missing pinned row:\n%s", out)
	}
}

func TestUpgradeAllWithOnlyPinnedPackages(t *testing.T) {
	pinned := samplePackage("bschaatsbergen/cidr", "v2.2.0")
	pinned.Options.Tag = "v2.2.0"
	storePath := useTempInstalledStore(t, pinned)
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("pinned package must not be upgraded: %+v", *calls)
	}
	// Notice, then heading, then table: all output must share one stream.
	want := []string{
		"No packages can be upgraded automatically.",
		"The following packages have an upgrade available, but require explicit targeting for upgrade:",
		"Name",
		"---",
		"bschaatsbergen/cidr",
	}
	previous := -1
	for _, fragment := range want {
		index := strings.Index(out, fragment)
		if index <= previous {
			t.Fatalf("output out of order at %q:\n%s", fragment, out)
		}
		previous = index
	}
}

func TestUpgradeAllWithNothingToDo(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, map[string]string{"a/one": "v1.0.0"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "--all")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("unexpected calls: %+v", *calls)
	}
	if strings.TrimSpace(out) != "No available upgrades." {
		t.Fatalf("output = %q", out)
	}
}

func TestUpgradeAllContinuesAfterFailure(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackage("a/one", "v1.0.0"),
		samplePackage("b/two", "v1.0.0"),
	)
	stubRefresh(t, map[string]string{"a/one": "v1.1.0", "b/two": "v1.1.0"})
	calls := stubEngine(t, storePath, map[string]error{"a/one": errors.New("boom")})

	out, err := runCLI(t, "upgrade", "--all")
	if err == nil || !strings.Contains(err.Error(), "a/one") {
		t.Fatalf("expected failure error, got %v", err)
	}
	if len(*calls) != 2 {
		t.Fatalf("expected both packages attempted, got %+v", *calls)
	}
	if !strings.Contains(out, "failed to upgrade a/one: boom") {
		t.Fatalf("missing failure message:\n%s", out)
	}

	if storedPackage(t, storePath, "github:b/two").InstalledTag != "v1.1.0" {
		t.Fatal("second package should still have been upgraded")
	}
}

func TestUpgradeNamedPinnedPackageRepins(t *testing.T) {
	pinned := samplePackage("bschaatsbergen/cidr", "v2.2.0")
	pinned.Options.Tag = "v2.2.0"
	storePath := useTempInstalledStore(t, pinned)
	stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "bschaatsbergen/cidr"); err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || (*calls)[0].Opts.Tag != "v2.3.0" {
		t.Fatalf("calls = %+v", *calls)
	}

	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	pkg := store.Packages["github:bschaatsbergen/cidr"][0]
	if pkg.Options.Tag != "v2.3.0" {
		t.Fatalf("options.tag = %q, want v2.3.0", pkg.Options.Tag)
	}
	if pkg.InstalledTag != "v2.3.0" {
		t.Fatalf("installed_tag = %q", pkg.InstalledTag)
	}
}

func TestUpgradeNamedPreservesUpgradeOnlyOption(t *testing.T) {
	pkg := samplePackage("a/one", "v1.0.0")
	pkg.Options.UpgradeOnly = true
	storePath := useTempInstalledStore(t, pkg)
	stubRefresh(t, map[string]string{"a/one": "v1.1.0"})
	calls := stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "a/one"); err != nil {
		t.Fatal(err)
	}
	if (*calls)[0].Opts.UpgradeOnly {
		t.Fatal("upgrade_only must not be forwarded to the engine")
	}

	if !storedPackage(t, storePath, "github:a/one").Options.UpgradeOnly {
		t.Fatal("stored upgrade_only option should be preserved")
	}
}

func TestUpgradeNamedAcceptsShortAndKeyNames(t *testing.T) {
	for _, target := range []string{"cidr", "github:bschaatsbergen/cidr"} {
		storePath := useTempInstalledStore(t, samplePackage("bschaatsbergen/cidr", "v2.2.0"))
		stubRefresh(t, map[string]string{"bschaatsbergen/cidr": "v2.3.0"})
		calls := stubEngine(t, storePath, nil)

		if _, err := runCLI(t, "upgrade", target); err != nil {
			t.Fatalf("%s: %v", target, err)
		}
		if len(*calls) != 1 {
			t.Fatalf("%s: calls = %+v", target, *calls)
		}
	}
}

func TestUpgradeNamedAlreadyUpToDate(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, map[string]string{"a/one": "v1.0.0"})
	calls := stubEngine(t, storePath, nil)

	out, err := runCLI(t, "upgrade", "a/one")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 0 {
		t.Fatalf("unexpected calls: %+v", *calls)
	}
	if !strings.Contains(out, "a/one is already up to date (v1.0.0).") {
		t.Fatalf("output:\n%s", out)
	}
}

func TestUpgradeNamedRefreshesOnlyRequestedPackage(t *testing.T) {
	storePath := useTempInstalledStore(t,
		samplePackage("a/one", "v1.0.0"),
		samplePackage("invalid", "v1.0.0"),
	)
	original := refreshPackage
	t.Cleanup(func() { refreshPackage = original })
	refreshed := []string{}
	refreshPackage = func(pkg installed.Package, opts options.Flags) (installed.Package, error) {
		refreshed = append(refreshed, pkg.Name)
		if pkg.Name == "invalid" {
			return pkg, errors.New("invalid argument (must be of the form user/repo)")
		}
		pkg.CurrentTag = "v1.0.0"
		pkg.RefreshedAt = time.Now().UTC()
		return pkg, nil
	}
	stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "a/one"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(refreshed, []string{"a/one"}) {
		t.Fatalf("refreshed = %#v, want only a/one", refreshed)
	}
}

func TestUpgradeNamedNotInstalled(t *testing.T) {
	useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, nil)

	if _, err := runCLI(t, "upgrade", "z/nine"); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Fatalf("expected not-installed error, got %v", err)
	}
}

func TestUpgradeNamedNonGitHubSource(t *testing.T) {
	pkg := samplePackage("some-tool", "v1.0.0")
	pkg.Source = "URL"
	useTempInstalledStore(t, pkg)
	stubRefresh(t, nil)

	_, err := runCLI(t, "upgrade", "some-tool")
	if err == nil || !strings.Contains(err.Error(), "no upgrade can be determined") {
		t.Fatalf("expected source error, got %v", err)
	}
}

func TestUpgradeNamedPropagatesEngineFailure(t *testing.T) {
	storePath := useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	stubRefresh(t, map[string]string{"a/one": "v1.1.0"})
	stubEngine(t, storePath, map[string]error{"a/one": errors.New("boom")})

	if _, err := runCLI(t, "upgrade", "a/one"); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected engine error, got %v", err)
	}
}

func TestUpgradePropagatesRefreshFailure(t *testing.T) {
	useTempInstalledStore(t, samplePackage("a/one", "v1.0.0"))
	original := refreshPackage
	t.Cleanup(func() { refreshPackage = original })
	refreshPackage = func(installed.Package, options.Flags) (installed.Package, error) {
		return installed.Package{}, errors.New("network down")
	}

	if _, err := runCLI(t, "upgrade"); err == nil || !strings.Contains(err.Error(), "network down") {
		t.Fatalf("expected refresh error, got %v", err)
	}
}

func TestUpgradeTableAlignsColumns(t *testing.T) {
	useTempInstalledStore(t,
		samplePackage("a/short", "v1.0.0"),
		samplePackage("a/a-much-longer-name", "v1.0.0"),
	)
	stubRefresh(t, map[string]string{"a/short": "v1.1.0", "a/a-much-longer-name": "v1.1.0"})

	out, err := runCLI(t, "upgrade", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	header, separator := lines[0], lines[1]
	if len(separator) != len(header) {
		t.Fatalf("separator width %d != header width %d:\n%s", len(separator), len(header), out)
	}
	if strings.Trim(separator, "-") != "" {
		t.Fatalf("separator = %q", separator)
	}
	versionColumn := strings.Index(header, "Version")
	for _, line := range lines[2:4] {
		if !strings.HasPrefix(line[versionColumn:], "v1.0.0") {
			t.Fatalf("column misaligned in %q (expected v1.0.0 at %d)", line, versionColumn)
		}
	}
}

func TestUpgradeStoreLoadFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XGET_CONFIG", "")
	t.Setenv("EGET_CONFIG", "")
	storePath := filepath.Join(home, ".config", "xget", ".xget.installed.yml")
	if err := os.MkdirAll(filepath.Dir(storePath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(storePath, []byte("packages:\n  bad: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := runCLI(t, "upgrade"); err == nil {
		t.Fatal("expected store parse error")
	}
}

func TestUpgradeLayersGlobalRepoAndStoredOptions(t *testing.T) {
	pkg := installed.Package{
		Name:            "owner/tool",
		InstallLocation: "/opt/bin",
		InstalledTag:    "v1.0.0",
		CurrentTag:      "v1.0.0",
		Source:          "GitHub",
		Options:         installed.Options{ExtractFile: "*.exe"},
	}
	storePath := useTempInstalledStore(t, pkg)
	stubRefresh(t, map[string]string{"owner/tool": "v1.1.0"})
	calls := stubEngine(t, storePath, nil)

	cfgPath := writeUpgradeConfig(t, strings.Join([]string{
		"global:",
		"  target: /global/bin",
		"  system: linux/amd64",
		"  file: global.bin",
		"  upgrade_only: true",
		"\"owner/tool\":",
		"  target: /repo/bin",
		"  asset_filters:",
		"    - \"{{.OS}}_{{.Arch}}\"",
		"  pre_release: true",
	}, "\n")+"\n")

	if _, err := runCLI(t, "upgrade", "--all", "--config", cfgPath); err != nil {
		t.Fatal(err)
	}
	opts := (*calls)[0].Opts

	// Repo section overrides global.
	if opts.Output != "/repo/bin" {
		t.Fatalf("output = %q, want /repo/bin", opts.Output)
	}
	if !opts.Prerelease {
		t.Fatal("repo pre_release should be applied")
	}
	// Global system is inherited by the repo section.
	if opts.System != "linux/amd64" {
		t.Fatalf("system = %q, want linux/amd64", opts.System)
	}
	// Stored options override config.
	if opts.ExtractFile != "*.exe" {
		t.Fatalf("file = %q, want *.exe", opts.ExtractFile)
	}
	// upgrade_only and tag are never forwarded.
	if opts.UpgradeOnly {
		t.Fatal("upgrade_only must not be forwarded")
	}
	if opts.Tag != "v1.1.0" {
		t.Fatalf("tag = %q, want v1.1.0", opts.Tag)
	}
	// Config asset filters get template substitution.
	if len(opts.Asset) != 1 || opts.Asset[0] != "linux_amd64" {
		t.Fatalf("asset = %v, want [linux_amd64]", opts.Asset)
	}
}

func TestRefreshUsesLayeredOptions(t *testing.T) {
	pkg := samplePackage("owner/tool", "v1.0.0")
	pkg.Options = installed.Options{Tag: "v1.0.0", System: "darwin/arm64"}
	useTempInstalledStore(t, pkg)
	seen := stubRefresh(t, map[string]string{"owner/tool": "v1.1.0"})

	cfgPath := writeUpgradeConfig(t, "\"owner/tool\":\n  pre_release: true\n  system: linux/amd64\n")
	if _, err := runCLI(t, "upgrade", "--config", cfgPath); err != nil {
		t.Fatal(err)
	}
	if len(*seen) != 1 {
		t.Fatalf("refresh calls = %d", len(*seen))
	}
	opts := (*seen)[0]
	if !opts.Prerelease {
		t.Fatal("config pre_release should reach refresh")
	}
	if opts.System != "darwin/arm64" {
		t.Fatalf("system = %q, want the stored value to win", opts.System)
	}
	if opts.Tag != "" || opts.UpgradeOnly {
		t.Fatalf("tag/upgrade_only must not reach refresh: %+v", opts)
	}
}
