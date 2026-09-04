package engine

import (
	"path/filepath"
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
