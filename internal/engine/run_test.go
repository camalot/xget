package engine

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/camalot/xget/internal/options"
)

func TestGetFinderUsesLatestReleaseForLatestTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		prerelease bool
		wantTag    string
	}{
		{name: "omitted", wantTag: "latest"},
		{name: "explicit", tag: "latest", wantTag: "latest"},
		{name: "with prereleases", tag: "latest", prerelease: true, wantTag: "latest"},
		{name: "specific version", tag: "v0.23.5", wantTag: "tags/v0.23.5"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			finder, _, err := getFinder("eza-community/eza", &options.Flags{
				Tag:        test.tag,
				Prerelease: test.prerelease,
			})
			if err != nil {
				t.Fatal(err)
			}
			githubFinder, ok := finder.(*GithubAssetFinder)
			if !ok {
				t.Fatalf("finder = %T, want *GithubAssetFinder", finder)
			}
			if githubFinder.Tag != test.wantTag {
				t.Errorf("tag = %q, want %q", githubFinder.Tag, test.wantTag)
			}
			if githubFinder.Prerelease != test.prerelease {
				t.Errorf("prerelease = %t, want %t", githubFinder.Prerelease, test.prerelease)
			}
		})
	}
}

func TestResolvedInstallLocation(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)

	location, err := resolvedInstallLocation([]string{"eza.exe"})
	if err != nil {
		t.Fatal(err)
	}
	if location != workingDirectory {
		t.Fatalf("location = %q, want %q", location, workingDirectory)
	}

	filePath := filepath.Join(workingDirectory, "bin", "eza.exe")
	location, err = resolvedInstallLocation([]string{filePath})
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(workingDirectory, "bin"); location != want {
		t.Fatalf("location = %q, want %q", location, want)
	}
}

// Regression: --non-interactive must still print the candidate list to stderr
// before failing, so users can see why detection was ambiguous.
func TestUserSelectNonInteractivePrintsChoicesBeforeError(t *testing.T) {
	SetNonInteractive(true)
	defer SetNonInteractive(false)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	_, selErr := userSelect([]interface{}{"a", "b", "c"})

	_ = w.Close()
	os.Stderr = origStderr
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if !errors.Is(selErr, ErrNonInteractive) {
		t.Fatalf("err = %v, want ErrNonInteractive", selErr)
	}
	for _, want := range []string{"(1) a", "(2) b", "(3) c"} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("stderr = %q, want to contain %q", out, want)
		}
	}
}

// Regression: an unquoted "~" shell-expanded ignore matcher (e.g. PowerShell
// resolving "~\.sha512$" to "C:\Users\<you>\.sha512$") must not silently pass
// through checksum/signature assets. This locks down that quoted "~" regex
// ignore matchers correctly narrow candidates to just the real archives.
func TestGetDetectorIgnoreExcludesChecksumAndSignatureAssets(t *testing.T) {
	opts := &options.Flags{
		System: "linux/amd64",
		Asset:  []string{".tar.gz"},
		Ignore: []string{`~\.sha512$`, `~\.sha256$`, `~\.sig$`},
	}
	detector, err := getDetector(opts)
	if err != nil {
		t.Fatal(err)
	}

	assets := []string{
		"git-cliff-2.14.1-x86_64-unknown-linux-gnu.tar.gz",
		"git-cliff-2.14.1-x86_64-unknown-linux-gnu.tar.gz.sha512",
		"git-cliff-2.14.1-x86_64-unknown-linux-gnu.tar.gz.sig",
		"git-cliff-2.14.1-x86_64-unknown-linux-musl.tar.gz",
		"git-cliff-2.14.1-x86_64-unknown-linux-musl.tar.gz.sha512",
		"git-cliff-2.14.1-x86_64-unknown-linux-musl.tar.gz.sig",
	}

	_, candidates, err := detector.Detect(assets)
	if err == nil {
		t.Fatal("expected error for multiple remaining candidates")
	}
	want := []string{
		"git-cliff-2.14.1-x86_64-unknown-linux-gnu.tar.gz",
		"git-cliff-2.14.1-x86_64-unknown-linux-musl.tar.gz",
	}
	if !reflect.DeepEqual(candidates, want) {
		t.Fatalf("candidates = %v, want %v", candidates, want)
	}
}
