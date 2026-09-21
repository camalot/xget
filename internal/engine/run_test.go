package engine

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/camalot/xget/internal/config"
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

func TestGithubSourceFinderKeepsArchiveExtension(t *testing.T) {
	source, err := config.Default().ResolveSource("github")
	if err != nil {
		t.Fatal(err)
	}
	finder := &GithubSourceFinder{Repo: "owner/project", Tag: "main", Tool: "project", Source: source}
	assets, err := finder.Find()
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 1 || !strings.HasSuffix(assets[0], "/project.tar.gz") {
		t.Fatalf("source assets = %#v", assets)
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

func TestIsDirectoryDestination(t *testing.T) {
	workingDirectory := t.TempDir()
	existingDirectory := filepath.Join(workingDirectory, "existing")
	if err := os.Mkdir(existingDirectory, 0750); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name string
		path string
		want bool
	}{
		{name: "existing directory", path: existingDirectory, want: true},
		{name: "trailing slash", path: filepath.Join(workingDirectory, "bin") + "/", want: true},
		{name: "trailing backslash", path: filepath.Join(workingDirectory, "bin") + "\\", want: true},
		{name: "file path", path: filepath.Join(workingDirectory, "xget"), want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isDirectoryDestination(test.path); got != test.want {
				t.Fatalf("isDirectoryDestination(%q) = %t, want %t", test.path, got, test.want)
			}
		})
	}
}

func TestIsRunningExecutableDestination(t *testing.T) {
	for _, test := range []struct {
		name        string
		destination string
		executable  string
		windows     bool
		want        bool
	}{
		{name: "same Windows path", destination: `C:\\Tools\\xget.exe`, executable: `C:\\Tools\\xget.exe`, windows: true, want: true},
		{name: "case-insensitive Windows path", destination: `C:\\Tools\\xget.exe`, executable: `c:\\tools\\XGET.EXE`, windows: true, want: true},
		{name: "different executable", destination: `C:\\Tools\\xget.exe`, executable: `C:\\Tools\\other.exe`, windows: true, want: false},
		{name: "non-Windows platform", destination: "/usr/local/bin/xget", executable: "/usr/local/bin/xget", windows: false, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isRunningExecutableDestination(test.destination, test.executable, test.windows); got != test.want {
				t.Fatalf("isRunningExecutableDestination(%q, %q, %t) = %t, want %t", test.destination, test.executable, test.windows, got, test.want)
			}
		})
	}
}

func TestReplaceStagedExecutable(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "xget.exe")
	if err := os.WriteFile(destination, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination+".new", []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := replaceStagedExecutable(destination); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		path string
		want string
	}{
		{path: destination, want: "new"},
		{path: destination + ".old", want: "old"},
	} {
		content, err := os.ReadFile(test.path)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(content); got != test.want {
			t.Fatalf("content of %q = %q, want %q", test.path, got, test.want)
		}
	}
	if _, err := os.Stat(destination + ".new"); !os.IsNotExist(err) {
		t.Fatalf("staged executable still exists: %v", err)
	}
}

func TestShouldRecordInstall(t *testing.T) {
	tests := []struct {
		name string
		opts options.Flags
		want bool
	}{
		{name: "default", opts: options.Flags{}, want: true},
		{name: "untracked", opts: options.Flags{Untracked: true}, want: false},
		{name: "stdout", opts: options.Flags{Output: "-"}, want: false},
		{name: "untracked with output", opts: options.Flags{Untracked: true, Output: "/usr/local/bin"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldRecordInstall(test.opts); got != test.want {
				t.Fatalf("shouldRecordInstall = %t, want %t", got, test.want)
			}
		})
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
