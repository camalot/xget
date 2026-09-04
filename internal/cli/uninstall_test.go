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
