package cli

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
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

func TestSplitTargetProvider(t *testing.T) {
	tests := []struct {
		target   string
		repo     string
		provider string
		ok       bool
	}{
		{target: "gitlab:gitlab-org/cli", repo: "gitlab-org/cli", provider: "gitlab", ok: true},
		{target: "work:group/subgroup/project@v1", repo: "group/subgroup/project@v1", provider: "work", ok: true},
		{target: "https://gitlab.com/group/project", repo: "https://gitlab.com/group/project"},
		{target: `C:\tools\archive.zip`, repo: `C:\tools\archive.zip`},
		{target: "owner/repo", repo: "owner/repo"},
	}

	for _, test := range tests {
		repo, provider, ok := splitTargetProvider(test.target)
		if repo != test.repo || provider != test.provider || ok != test.ok {
			t.Errorf("splitTargetProvider(%q) = (%q, %q, %t), want (%q, %q, %t)", test.target, repo, provider, ok, test.repo, test.provider, test.ok)
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
			for _, name := range []string{"installed", "check"} {
				if sub.Flags().Lookup(name) == nil {
					t.Fatalf("expected list command to include an --%s flag", name)
				}
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
		for _, name := range []string{"tag", "to", "asset", "ignore", "source", "provider"} {
			if sub.Flags().Lookup(name) == nil {
				t.Fatalf("expected install command to include --%s", name)
			}
		}
		return
	}
	t.Fatal("expected root command to include an install subcommand")
}

func TestOptionsForTargetProviderPrecedence(t *testing.T) {
	cfg := config.Default()
	cfg.Global.SourceType = "gitlab"
	cfg.Sources["work"] = config.Source{Name: "work", Type: "github", Host: "github.example.com", APIURL: "https://github.example.com/api/v3"}
	cfg.Repositories["owner/repo"] = config.Repository{Name: "owner/repo", SourceType: "work"}

	cmd := newRootCommand()
	flags := &rootFlags{}
	opts, err := optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if opts.SourceType != "work" || opts.SourceConfig.Host != "github.example.com" {
		t.Fatalf("repository source was not resolved: %#v", opts)
	}

	flags.provider = "gitlab"
	if err := cmd.Flags().Set("provider", "gitlab"); err != nil {
		t.Fatal(err)
	}
	opts, err = optionsForTarget(cfg, cmd, flags, "owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if opts.SourceType != "gitlab" || opts.SourceConfig.Type != "gitlab" {
		t.Fatalf("CLI provider did not win: %#v", opts)
	}
	if opts.Source {
		t.Fatal("--provider must not enable source archive downloads")
	}
}

func TestOptionsForTargetInlineProvider(t *testing.T) {
	cfg := config.Default()
	cfg.Repositories["gitlab-org/cli"] = config.Repository{Name: "gitlab-org/cli", Target: "./bin"}
	cmd := newRootCommand()
	flags := &rootFlags{}

	opts, err := optionsForTargetProvider(cfg, cmd, flags, "gitlab-org/cli", "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	if opts.SourceType != "gitlab" || opts.SourceConfig.Type != "gitlab" {
		t.Fatalf("inline provider was not resolved: %#v", opts)
	}
	if !strings.HasSuffix(opts.Output, "bin") {
		t.Fatalf("canonical repository config was not applied: %#v", opts)
	}

	flags.provider = "github"
	if err := cmd.Flags().Set("provider", "github"); err != nil {
		t.Fatal(err)
	}
	if _, err := optionsForTargetProvider(cfg, cmd, flags, "gitlab-org/cli", "gitlab"); err == nil || !strings.Contains(err.Error(), "conflicting providers") {
		t.Fatalf("conflicting providers error = %v", err)
	}
}

func TestRateCommandPrintsTokenGuidanceToStderr(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), ".xget.yml")
	if err := os.WriteFile(configPath, []byte("global: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XGET_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("EGET_GITHUB_TOKEN", "")
	original := getRateLimit
	t.Cleanup(func() { getRateLimit = original })
	getRateLimit = func() (engine.RateLimit, error) {
		return engine.RateLimit{Limit: 60, Remaining: 42}, nil
	}

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{"rate", "--config", configPath})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "Limit: 60, Remaining: 42") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "XGET_GITHUB_TOKEN") {
		t.Fatalf("token guidance must not be written to stdout: %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "XGET_GITHUB_TOKEN") {
		t.Fatalf("stderr missing token guidance: %q", stderr.String())
	}
}

func TestRootCommandIncludesRateSubcommand(t *testing.T) {
	cmd := newRootCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "rate" {
			return
		}
	}
	t.Fatal("expected root command to include a rate subcommand")
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

func TestOptionsForTargetMergesGlobalArrays(t *testing.T) {
	cfg := config.Default()
	cfg.Global.Ignore = []string{".sbom.json"}
	cfg.Repositories["foo/bar"] = config.Repository{Ignore: []string{"debug"}}
	flags := &rootFlags{}
	cmd := newInstallCommand(flags)
	if err := cmd.ParseFlags([]string{"--ignore", "musl", "--clobber_global.ignore=false"}); err != nil {
		t.Fatal(err)
	}
	opts, err := optionsForTarget(cfg, cmd, flags, "foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"musl", "debug", ".sbom.json"}) {
		t.Fatalf("merged ignore = %#v", opts.Ignore)
	}
}

func TestOptionsForTargetClobberPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		global  string
		repo    string
		args    []string
		want    []string
		wantCLI []string
	}{
		{name: "default replacement", want: []string{"repo", "shared"}, wantCLI: []string{"cli", "shared"}},
		{name: "global merge", global: "false", want: []string{"repo", "shared", "global"}, wantCLI: []string{"cli", "shared", "repo", "global"}},
		{name: "repository merge", global: "true", repo: "false", want: []string{"repo", "shared", "global"}, wantCLI: []string{"cli", "shared", "repo", "global"}},
		{name: "repository replacement", global: "false", repo: "true", want: []string{"repo", "shared"}, wantCLI: []string{"cli", "shared"}},
		{name: "CLI merge", repo: "true", args: []string{"--clobber_global.ignore=false", "--clobber_global.asset_filters=false"}, want: []string{"repo", "shared", "global"}, wantCLI: []string{"cli", "shared", "repo", "global"}},
		{name: "CLI replacement", global: "false", repo: "false", args: []string{"--clobber_global.ignore=true", "--clobber_global.asset_filters=true"}, want: []string{"repo", "shared"}, wantCLI: []string{"cli", "shared"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := "global:\n  config_merge: false\n  ignore: [global, shared]\n  asset_filters: [global, shared]\n  target: global-bin\n"
			if test.global != "" {
				content += "  clobber_global:\n    ignore: " + test.global + "\n    asset_filters: " + test.global + "\n"
			}
			content += "foo/bar:\n  ignore: [repo, shared]\n  asset_filters: [repo, shared]\n  target: repo-bin\n"
			if test.repo != "" {
				content += "  clobber_global:\n    ignore: " + test.repo + "\n    asset_filters: " + test.repo + "\n"
			}
			path := filepath.Join(t.TempDir(), ".xget.yml")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			cfg, err := config.LoadQuiet(path)
			if err != nil {
				t.Fatal(err)
			}
			flags := &rootFlags{}
			cmd := newInstallCommand(flags)
			if err := cmd.ParseFlags(test.args); err != nil {
				t.Fatal(err)
			}
			opts, err := optionsForTarget(cfg, cmd, flags, "foo/bar")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opts.Ignore, test.want) || !reflect.DeepEqual(opts.Asset, test.want) || opts.Output != "repo-bin" {
				t.Fatalf("config resolution = %#v; want arrays %v and repository scalar", opts, test.want)
			}
			if err := cmd.ParseFlags([]string{"--ignore", "cli,shared", "--asset", "cli,shared", "--to", "cli-bin"}); err != nil {
				t.Fatal(err)
			}
			opts, err = optionsForTarget(cfg, cmd, flags, "foo/bar")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opts.Ignore, test.wantCLI) || !reflect.DeepEqual(opts.Asset, test.wantCLI) || opts.Output != "cli-bin" {
				t.Fatalf("CLI resolution = %#v; want arrays %v and CLI scalar", opts, test.wantCLI)
			}
			if !reflect.DeepEqual(cfg.Global.Ignore, []string{"global", "shared"}) || !reflect.DeepEqual(cfg.Repositories["foo/bar"].Ignore, []string{"repo", "shared"}) {
				t.Fatal("resolution mutated the loaded configuration")
			}
		})
	}
}

func TestOptionsForTargetClobberArraysIndependent(t *testing.T) {
	cfg := config.Default()
	cfg.Global.Ignore = []string{"global-ignore"}
	cfg.Global.AssetFilters = []string{"global-asset"}
	cfg.Global.ClobberGlobal = map[string]bool{"ignore": false, "asset_filters": false}
	cfg.Repositories["foo/bar"] = config.Repository{
		Ignore:        []string{},
		AssetFilters:  []string{},
		ClobberGlobal: map[string]bool{"ignore": true},
	}
	opts, err := configOptionsForTarget(cfg, "foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Ignore) != 0 || !reflect.DeepEqual(opts.Asset, cfg.Global.AssetFilters) {
		t.Fatalf("independent controls with empty arrays = %#v", opts)
	}
	flags := &rootFlags{}
	cmd := newInstallCommand(flags)
	if err := cmd.ParseFlags([]string{"--ignore", "cli", "--asset", "cli"}); err != nil {
		t.Fatal(err)
	}
	opts, err = optionsForTarget(cfg, cmd, flags, "unknown/repo")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"cli", "global-ignore"}) || !reflect.DeepEqual(opts.Asset, []string{"cli", "global-asset"}) {
		t.Fatalf("merge without repository section = %#v", opts)
	}
}

func TestOptionsForTarget_UntrackedFlag(t *testing.T) {
	cfg := &config.Config{}

	cmd := newRootCommand()
	flags := &rootFlags{}
	opts, err := optionsForTarget(cfg, cmd, flags, "jgm/pandoc")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if opts.Untracked {
		t.Fatal("expected installs to be tracked by default")
	}

	cmd = newRootCommand()
	flags = &rootFlags{untracked: true}
	if err := cmd.Flags().Set("untracked", "true"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	opts, err = optionsForTarget(cfg, cmd, flags, "jgm/pandoc")
	if err != nil {
		t.Fatalf("optionsForTarget returned error: %v", err)
	}
	if !opts.Untracked {
		t.Fatal("expected --untracked to set Untracked")
	}
}

func TestInstallCommandSupportsUntrackedFlag(t *testing.T) {
	root := newRootCommand()
	for _, sub := range root.Commands() {
		if sub.Name() != "install" {
			continue
		}
		if sub.Flags().Lookup("untracked") == nil {
			t.Fatal("install command missing --untracked flag")
		}
		return
	}
	t.Fatal("expected root command to include an install subcommand")
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
				File:         "{{.Arch}}/*.exe",
				Binary:       "tool-{{.OS}}-{{.Arch}}",
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
	if opts.ExtractFile != "arm64/*.exe" || opts.Binary != "tool-linux-arm64" {
		t.Fatalf("expected substituted file and binary, got file=%q binary=%q", opts.ExtractFile, opts.Binary)
	}
}

func TestOptionsForTarget_SubstitutesCliAssetButNotIgnore(t *testing.T) {
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
	if !reflect.DeepEqual(opts.Asset, []string{"linux_arm64.tar.gz"}) {
		t.Fatalf("expected substituted cli asset filters, got %#v", opts.Asset)
	}
	if !reflect.DeepEqual(opts.Ignore, []string{"{{.OS}}_{{.Arch}}.sbom.json"}) {
		t.Fatalf("expected cli ignore filters unsubstituted, got %#v", opts.Ignore)
	}
}

func TestOptionsForTargetSubstitutesCLIFileAssetAndBinary(t *testing.T) {
	for _, binaryFlag := range []string{"binary", "name"} {
		for _, system := range []string{"", "all", "windows/amd64", "linux/arm64"} {
			t.Run(binaryFlag+"/"+system, func(t *testing.T) {
				cmd := newRootCommand()
				flags := &rootFlags{
					system: system, extractFile: "{{.Arch}}/*.exe",
					binary: "tool-{{.OS}}-{{.Arch}}",
					asset:  []string{"{{.OS}}_{{.Arch}}.zip", "not:{{.Arch}}-debug"},
				}
				for name, value := range map[string]string{
					"system": system, "file": flags.extractFile,
					binaryFlag: flags.binary, "asset": strings.Join(flags.asset, ","),
				} {
					if err := cmd.Flags().Set(name, value); err != nil {
						t.Fatal(err)
					}
				}
				cfg := config.Default()
				cfg.Repositories["owner/repo"] = config.Repository{
					System: "darwin/arm64", File: "config-file", Binary: "config-binary",
					AssetFilters: []string{"config-asset"},
				}
				opts, err := optionsForTarget(cfg, cmd, flags, "owner/repo")
				if err != nil {
					t.Fatal(err)
				}
				goos, arch := runtime.GOOS, runtime.GOARCH
				if system != "" && system != "all" {
					parts := strings.Split(system, "/")
					goos, arch = parts[0], parts[1]
				}
				if opts.ExtractFile != arch+"/*.exe" || opts.Binary != "tool-"+goos+"-"+arch {
					t.Fatalf("file=%q binary=%q", opts.ExtractFile, opts.Binary)
				}
				wantAssets := []string{goos + "_" + arch + ".zip", "not:" + arch + "-debug"}
				if !reflect.DeepEqual(opts.Asset, wantAssets) {
					t.Fatalf("assets=%#v, want %#v", opts.Asset, wantAssets)
				}
			})
		}
	}
}

func TestInstallTemplatesSelectArchitectureFiles(t *testing.T) {
	useTempInstalledStore(t)
	archive := filepath.Join(t.TempDir(), "tool_windows_amd64.zip")
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, name := range []string{"amd64/tool.exe", "amd64/helper.exe", "arm64/tool.exe", "386/tool.exe", "amd64/README.md"} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archive, buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := writeUpgradeConfig(t, "global:\n  xget_update_check: false\n")
	for _, all := range []bool{false, true} {
		out := t.TempDir()
		args := []string{
			"install", archive, "--config", cfg, "--system", "windows/amd64",
			"--asset", "{{.OS}}_{{.Arch}}.zip", "--to", out, "--untracked", "--quiet",
		}
		wantNames := []string{"tool-windows-amd64.exe"}
		if all {
			args = append(args, "--file", "{{.Arch}}/*.exe", "--all")
			wantNames = []string{"helper.exe", "tool.exe"}
		} else {
			args = append(args, "--file", "{{.Arch}}/tool.exe", "--binary", "tool-{{.OS}}-{{.Arch}}")
		}
		if _, err := runCLI(t, args...); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(out)
		if err != nil {
			t.Fatal(err)
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
			original := "amd64/tool.exe"
			if entry.Name() == "helper.exe" {
				original = "amd64/helper.exe"
			}
			if got := readFileString(t, filepath.Join(out, entry.Name())); got != original {
				t.Fatalf("%s content=%q, want %q", entry.Name(), got, original)
			}
		}
		if !reflect.DeepEqual(names, wantNames) {
			t.Fatalf("all=%t: extracted names=%#v, want %#v", all, names, wantNames)
		}
	}
}
