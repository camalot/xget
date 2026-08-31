package cli

import (
	"reflect"
	"testing"

	"github.com/camalot/xget/internal/config"
)

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
