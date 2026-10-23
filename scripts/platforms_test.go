package scripts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type releasePlatform struct {
	Release   string `json:"release"`
	Goarch    string `json:"goarch"`
	Goarm     string `json:"goarm"`
	Container string `json:"container"`
	Debian    string `json:"debian"`
	RPM       string `json:"rpm"`
	RPMImage  string `json:"rpm_image"`
	Alpine    string `json:"alpine"`
	AUR       string `json:"aur"`
}

func readRepositoryFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestReleasePlatforms(t *testing.T) {
	var platforms []releasePlatform
	if err := json.Unmarshal(readRepositoryFile(t, ".packages/platforms.json"), &platforms); err != nil {
		t.Fatal(err)
	}
	var config struct {
		Builds []struct {
			Goos   []string
			Goarch []string
			Goarm  []string
		}
		Archives []struct {
			NameTemplate string `yaml:"name_template"`
		}
	}
	if err := yaml.Unmarshal(readRepositoryFile(t, ".goreleaser.yaml"), &config); err != nil {
		t.Fatal(err)
	}
	expected := make(map[string]bool)
	for _, build := range config.Builds {
		for _, targetOS := range build.Goos {
			if targetOS != "linux" {
				continue
			}
			for _, arch := range build.Goarch {
				if arch == "arm" {
					if len(build.Goarm) == 0 {
						t.Fatal("Linux ARM builds must explicitly set goarm")
					}
					for _, arm := range build.Goarm {
						expected["armv"+arm] = true
					}
				} else {
					expected[arch] = true
				}
			}
		}
	}
	seen := make(map[string]bool)
	packageNames := make(map[string]bool)
	for _, platform := range platforms {
		if seen[platform.Release] || !expected[platform.Release] {
			t.Errorf("duplicate or unbuilt release platform: %s", platform.Release)
		}
		seen[platform.Release] = true
		if platform.Goarch == "arm" {
			if platform.Release != "armv"+platform.Goarm || platform.Container != "linux/arm/v"+platform.Goarm {
				t.Errorf("inconsistent ARM platform: %+v", platform)
			}
		} else if platform.Release != platform.Goarch || platform.Goarm != "" || platform.Container != "linux/"+platform.Goarch {
			t.Errorf("inconsistent platform: %+v", platform)
		}
		if platform.RPM != "" && platform.RPMImage == "" {
			t.Errorf("RPM platform %s has no install-test image", platform.Release)
		}
		for format, name := range map[string]string{"debian": platform.Debian, "rpm": platform.RPM, "alpine": platform.Alpine, "aur": platform.AUR} {
			if name == "" {
				continue
			}
			key := format + ":" + name
			if packageNames[key] {
				t.Errorf("package architecture collision: %s", key)
			}
			packageNames[key] = true
		}
	}
	for release := range expected {
		if !seen[release] {
			t.Errorf("Linux release %s is missing from .packages/platforms.json", release)
		}
	}
	if len(platforms) == 0 || len(config.Archives) != 1 || !strings.Contains(config.Archives[0].NameTemplate, "{{ if .Arm }}v{{ .Arm }}{{ end }}") {
		t.Error("archive names must include the ARM variant")
	}
}

func TestPackagingWorkflowPlatforms(t *testing.T) {
	for _, name := range []string{"release", "publish-debian", "publish-rpm", "publish-alpine", "publish-aur"} {
		t.Run(name, func(t *testing.T) {
			var workflow struct {
				Jobs map[string]struct {
					Strategy struct {
						Matrix map[string]string
					}
					Steps []struct {
						Uses string
						With map[string]string
					}
				}
			}
			if err := yaml.Unmarshal(readRepositoryFile(t, ".github/workflows/"+name+".yml"), &workflow); err != nil {
				t.Fatal(err)
			}
			for _, format := range []string{"debian", "rpm", "alpine", "aur"} {
				job, exists := workflow.Jobs[format]
				if !exists {
					continue
				}
				if !strings.Contains(job.Strategy.Matrix["platform"], "fromJSON(needs.platforms.outputs.") {
					t.Errorf("%s must use the manifest-generated matrix", format)
				}
				found := false
				for _, step := range job.Steps {
					if step.Uses == "./.github/actions/"+format {
						found = true
						if step.With["architecture"] != "${{ matrix.platform.release }}" {
							t.Errorf("%s action must receive matrix.platform.release", format)
						}
					}
				}
				if !found {
					t.Fatalf("missing %s action", format)
				}
			}
		})
	}
}
