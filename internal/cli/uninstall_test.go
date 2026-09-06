package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/camalot/xget/internal/installed"
)

func TestUninstallRemovesTrackedFilesAndPackage(t *testing.T) {
	directory := t.TempDir()
	first := filepath.Join(directory, "eza")
	second := filepath.Join(directory, "eza-helper")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	storePath := useTempInstalledStore(t, installed.Package{
		Name:            "eza-community/eza",
		InstallLocation: directory,
		ExtractedFiles:  []string{first, second},
		Source:          "GitHub",
	})

	out, err := runCLI(t, "uninstall", "eza-community/eza")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Uninstalled `eza-community/eza`") {
		t.Fatalf("output = %q", out)
	}
	for _, path := range []string{first, second} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got %v", path, err)
		}
	}
	store, err := installed.Load(storePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Packages) != 0 {
		t.Fatalf("expected package to be removed from store, got %#v", store.Packages)
	}
}

func TestUninstallFallsBackToXgetBin(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XGET_BIN", directory)
	path := filepath.Join(directory, "eza")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	useTempInstalledStore(t)

	out, err := runCLI(t, "remove", "eza")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Removed `"+path+"`") {
		t.Fatalf("output = %q", out)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected fallback file to be removed, got %v", err)
	}
}

func TestLegacyRemoveFlagUsesUninstall(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XGET_BIN", directory)
	path := filepath.Join(directory, "eza")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	useTempInstalledStore(t)

	out, err := runCLI(t, "eza", "--remove")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Removed `"+path+"`") {
		t.Fatalf("output = %q", out)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected fallback file to be removed, got %v", err)
	}
}

func TestUninstallReportsMissingPackageAndFile(t *testing.T) {
	directory := t.TempDir()
	useTempInstalledStore(t)

	_, err := runCLI(t, "uninstall", "eza-community/eza", "--from", directory)
	if err == nil || !strings.Contains(err.Error(), "is not installed") || !strings.Contains(err.Error(), "was not found") {
		t.Fatalf("error = %v, want missing package and file message", err)
	}
}

func TestUninstallRejectsTraversalFallbackTarget(t *testing.T) {
	useTempInstalledStore(t)

	_, err := runCLI(t, "uninstall", "../eza", "--from", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "invalid target") {
		t.Fatalf("error = %v, want invalid target error", err)
	}
}

// trackedAt writes a file named eza in directory and returns a record for it.
func trackedAt(t *testing.T, directory string) installed.Package {
	t.Helper()
	path := filepath.Join(directory, "eza")
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	return installed.Package{
		Name:            "eza-community/eza",
		InstallLocation: directory,
		ExtractedFiles:  []string{path},
		Source:          "GitHub",
	}
}

func TestUninstallRequiresLocationWhenTrackedInSeveralPlaces(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	firstPkg, secondPkg := trackedAt(t, first), trackedAt(t, second)
	useTempInstalledStore(t, firstPkg, secondPkg)

	_, err := runCLI(t, "uninstall", "eza-community/eza")
	if err == nil || !strings.Contains(err.Error(), "installed to multiple locations") {
		t.Fatalf("error = %v, want ambiguous location error", err)
	}
	if !strings.Contains(err.Error(), first) || !strings.Contains(err.Error(), second) {
		t.Fatalf("error should list both locations: %v", err)
	}
	for _, path := range []string{firstPkg.ExtractedFiles[0], secondPkg.ExtractedFiles[0]} {
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("expected %s to be left in place, got %v", path, statErr)
		}
	}
}

func TestUninstallFromRemovesOnlyTheSelectedLocation(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	firstPkg, secondPkg := trackedAt(t, first), trackedAt(t, second)
	storePath := useTempInstalledStore(t, firstPkg, secondPkg)

	out, err := runCLI(t, "uninstall", "eza-community/eza", "--from", second)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Uninstalled `eza-community/eza`") {
		t.Fatalf("output = %q", out)
	}
	if _, statErr := os.Stat(secondPkg.ExtractedFiles[0]); !os.IsNotExist(statErr) {
		t.Fatalf("expected selected copy to be removed, got %v", statErr)
	}
	if _, statErr := os.Stat(firstPkg.ExtractedFiles[0]); statErr != nil {
		t.Fatalf("expected other copy to remain, got %v", statErr)
	}

	records := storedPackages(t, storePath, "github:eza-community/eza")
	if len(records) != 1 || records[0].InstallLocation != first {
		t.Fatalf("records = %#v, want only %s", records, first)
	}
}

func TestUninstallFromUntrackedLocationErrors(t *testing.T) {
	useTempInstalledStore(t, trackedAt(t, t.TempDir()))

	_, err := runCLI(t, "uninstall", "eza-community/eza", "--from", "/nowhere/bin")
	if err == nil || err.Error() != "package eza-community/eza is not installed to /nowhere/bin" {
		t.Fatalf("error = %v", err)
	}
}

func TestUninstallAllRemovesEveryLocation(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	firstPkg, secondPkg := trackedAt(t, first), trackedAt(t, second)
	storePath := useTempInstalledStore(t, firstPkg, secondPkg)

	if _, err := runCLI(t, "uninstall", "eza-community/eza", "--all"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{firstPkg.ExtractedFiles[0], secondPkg.ExtractedFiles[0]} {
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("expected %s to be removed, got %v", path, statErr)
		}
	}
	if records := storedPackages(t, storePath, "github:eza-community/eza"); len(records) != 0 {
		t.Fatalf("expected no records left, got %#v", records)
	}
}
