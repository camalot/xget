package cli

import (
	"reflect"
	"testing"

	"github.com/camalot/xget/internal/config"
)

func TestSplitTargetTag(t *testing.T) {
	tests := []struct {
		target string
		repo   string
		tag    string
		ok     bool
	}{
		{target: "slavaGanzin/await@2.1.0", repo: "slavaGanzin/await", tag: "2.1.0", ok: true},
		{target: "eza-community/eza@v0.23.5", repo: "eza-community/eza", tag: "v0.23.5", ok: true},
		{target: "eza-community/eza@latest", repo: "eza-community/eza", tag: "latest", ok: true},
		{target: "https://user@example.com/asset", repo: "https://user@example.com/asset"},
		{target: "owner/repo@", repo: "owner/repo@"},
	}

	for _, test := range tests {
		repo, tag, ok := splitTargetTag(test.target)
		if repo != test.repo || tag != test.tag || ok != test.ok {
			t.Errorf("splitTargetTag(%q) = (%q, %q, %t), want (%q, %q, %t)", test.target, repo, tag, ok, test.repo, test.tag, test.ok)
		}
	}
}

func TestRootCommandIncludesCompletionSubcommand(t *testing.T) {
	cmd := newRootCommand()
	if got := cmd.Commands(); len(got) == 0 {
		t.Fatal("expected root command to define subcommands")
	}

	for _, sub := range cmd.Commands() {
		if sub.Name() == "completion" {
			return
		}
	}

	t.Fatal("expected root command to include a completion subcommand")
}

func TestRootCommandIncludesConfigFlag(t *testing.T) {
	cmd := newRootCommand()
	if cmd.Flags().Lookup("config") == nil {
		t.Fatal("expected root command to include a --config flag")
	}
}

func TestRootCommandIncludesListSubcommand(t *testing.T) {
	cmd := newRootCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "list" {
			if sub.Flags().Lookup("installed") == nil {
				t.Fatal("expected list command to include an --installed flag")
			}
			return
		}
	}
	t.Fatal("expected root command to include a list subcommand")
}

func TestRootCommandIncludesInstallSubcommandWithInstallFlags(t *testing.T) {
	cmd := newRootCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() != "install" {
			continue
		}
		for _, name := range []string{"tag", "to", "asset", "ignore"} {
			if sub.Flags().Lookup(name) == nil {
				t.Fatalf("expected install command to include --%s", name)
			}
		}
		return
	}
	t.Fatal("expected root command to include an install subcommand")
}

func TestOptionsForTarget_UsesConfigIgnoreAndOverrides(t *testing.T) {
	cmd := newRootCommand()
	flags := &rootFlags{}

	cfg := &config.Config{
		Global: config.Global{Ignore: []string{"global-ignore"}},
		Repositories: map[string]config.Repository{
			"owner/repo": {
				Ignore: []string{"repo-ignore"},
			},
		},
	}

	opts, err := optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"repo-ignore"}) {
		t.Fatalf("expected repo ignore, got %#v", opts.Ignore)
	}

	opts, err = optionsForTarget(cfg, cmd, flags, "other/repo")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"global-ignore"}) {
		t.Fatalf("expected global ignore, got %#v", opts.Ignore)
	}

	flags.ignore = []string{"cli-ignore"}
	if err := cmd.Flags().Set("ignore", "cli-ignore"); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	opts, err = optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"cli-ignore"}) {
		t.Fatalf("expected cli ignore override, got %#v", opts.Ignore)
	}
}

func TestOptionsForTarget_SubstitutesTemplateVarsInConfigAssetAndIgnore(t *testing.T) {
	cmd := newRootCommand()
	flags := &rootFlags{}

	cfg := &config.Config{
		Global: config.Global{Ignore: []string{"{{.OS}}-global-ignore"}},
		Repositories: map[string]config.Repository{
			"owner/repo": {
				System:       "linux/arm64",
				AssetFilters: []string{"{{.OS}}_{{.Arch}}.tar.gz"},
				Ignore:       []string{"{{.OS}}_{{.Arch}}.sbom.json"},
			},
		},
	}

	opts, err := optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !reflect.DeepEqual(opts.Asset, []string{"linux_arm64.tar.gz"}) {
		t.Fatalf("expected substituted asset filters, got %#v", opts.Asset)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"linux_arm64.sbom.json"}) {
		t.Fatalf("expected substituted ignore filters, got %#v", opts.Ignore)
	}
}

func TestOptionsForTarget_DoesNotSubstituteCliAssetAndIgnore(t *testing.T) {
	cmd := newRootCommand()
	flags := &rootFlags{}

	cfg := &config.Config{
		Repositories: map[string]config.Repository{
			"owner/repo": {
				System: "linux/arm64",
			},
		},
	}

	flags.asset = []string{"{{.OS}}_{{.Arch}}.tar.gz"}
	if err := cmd.Flags().Set("asset", "{{.OS}}_{{.Arch}}.tar.gz"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	flags.ignore = []string{"{{.OS}}_{{.Arch}}.sbom.json"}
	if err := cmd.Flags().Set("ignore", "{{.OS}}_{{.Arch}}.sbom.json"); err != nil {
		t.Fatalf("set flag: %v", err)
	}

	opts, err := optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !reflect.DeepEqual(opts.Asset, []string{"{{.OS}}_{{.Arch}}.tar.gz"}) {
		t.Fatalf("expected cli asset filters unsubstituted, got %#v", opts.Asset)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"{{.OS}}_{{.Arch}}.sbom.json"}) {
		t.Fatalf("expected cli ignore filters unsubstituted, got %#v", opts.Ignore)
	}
}
