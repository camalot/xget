package cli

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
	"github.com/camalot/xget/internal/home"
	"github.com/camalot/xget/internal/options"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	tag         string
	prerelease  bool
	source      bool
	output      string
	system      string
	extractFile string
	all         bool
	quiet       bool
	dlOnly      bool
	upgradeOnly bool
	asset       []string
	ignore      []string
	hash        bool
	verify      string
	rate        bool
	remove      bool
	from        string
	downloadAll bool
	disableSSL  bool
	config      string
}

var getRateLimit = engine.GetRateLimit

func Execute() error {
	root := newRootCommand()
	return root.Execute()
}

func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	return 1
}

// splitTargetTag separates the repo and release tag in owner/repo@tag targets.
// URLs and malformed repository targets are left unchanged for their respective
// handlers to validate.
func splitTargetTag(target string) (repo, tag string, ok bool) {
	if strings.Contains(target, "://") {
		return target, "", false
	}
	repo, tag, ok = strings.Cut(target, "@")
	if !ok || tag == "" || strings.Count(repo, "/") != 1 {
		return target, "", false
	}
	owner, name, valid := strings.Cut(repo, "/")
	if !valid || owner == "" || name == "" {
		return target, "", false
	}
	return repo, tag, true
}

func newRootCommand() *cobra.Command {
	f := &rootFlags{}

	cmd := &cobra.Command{
		Use:           "xget [TARGET]",
		Short:         "Download pre-built binaries from GitHub releases",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		RunE:          installRunE(f),
	}

	addInstallFlags(cmd, f)

	cmd.InitDefaultCompletionCmd("completion")
	cmd.AddCommand(newVersionCommand(), newListCommand(), newConfigCommand(), newUpgradeCommand(), newInstallCommand(f), newUninstallCommand(f), newRateCommand(f))

	return cmd
}

func newRateCommand(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "rate",
		Short:         "Show GitHub API rate limiting information",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          rateRunE(f),
	}
	cmd.Flags().BoolVarP(&f.disableSSL, "disable-ssl", "k", false, "disable SSL verification for download requests")
	cmd.Flags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
	return cmd
}

func newInstallCommand(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "install TARGET",
		Short:         "Download and install a pre-built binary from GitHub releases",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          installRunE(f),
	}
	addInstallFlags(cmd, f)
	return cmd
}

func addInstallFlags(cmd *cobra.Command, f *rootFlags) {
	cmd.Flags().StringVarP(&f.tag, "tag", "t", "", "tagged release to use instead of latest")
	cmd.Flags().BoolVar(&f.prerelease, "pre-release", false, "include pre-releases when fetching the latest version")
	cmd.Flags().BoolVar(&f.source, "source", false, "download the source code for the target repo instead of a release")
	cmd.Flags().StringVar(&f.output, "to", "", "move to given location after extracting")
	cmd.Flags().StringVarP(&f.system, "system", "s", "", "target system to download for (use all for all choices)")
	cmd.Flags().StringVarP(&f.extractFile, "file", "f", "", "glob to select files for extraction")
	cmd.Flags().BoolVar(&f.all, "all", false, "extract all candidate files")
	cmd.Flags().BoolVarP(&f.quiet, "quiet", "q", false, "only print essential output")
	cmd.Flags().BoolVarP(&f.dlOnly, "download-only", "d", false, "stop after downloading the asset (no extraction)")
	cmd.Flags().BoolVar(&f.upgradeOnly, "upgrade-only", false, "only download if release is more recent than current version")
	cmd.Flags().StringSliceVarP(&f.asset, "asset", "a", nil, "filter assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not:, escapes: ~~ and ^^, explicit literal: text: (for example ^musl, not:~.*\\.sbom\\.json$, text:~literal)")
	cmd.Flags().StringSliceVar(&f.ignore, "ignore", nil, "exclude assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not: (inverts ignore), escapes: ~~ and ^^, explicit literal: text:; can be specified multiple times")
	cmd.Flags().BoolVar(&f.hash, "sha256", false, "show the SHA-256 hash of the downloaded asset")
	cmd.Flags().StringVar(&f.verify, "verify-sha256", "", "verify the downloaded asset checksum against the one provided")
	cmd.Flags().StringVar(&f.verify, "verify", "", "verify the downloaded asset checksum; pass a hash or use --verify with no value to use GitHub's published SHA256 when available")
	cmd.Flags().Lookup("verify").NoOptDefVal = "auto"
	cmd.Flags().BoolVar(&f.rate, "rate", false, "show GitHub API rate limiting information")
	cmd.Flags().BoolVarP(&f.remove, "remove", "r", false, "uninstall the target package")
	cmd.Flags().StringVar(&f.from, "from", "", "directory to remove an untracked target from")
	cmd.Flags().BoolVarP(&f.downloadAll, "download-all", "D", false, "download all projects defined in the config file")
	cmd.Flags().BoolVarP(&f.disableSSL, "disable-ssl", "k", false, "disable SSL verification for download requests")
	cmd.Flags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
}

func installRunE(f *rootFlags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(f.config)
		if err != nil {
			return err
		}

		if err := configureGithubToken(cfg); err != nil {
			return err
		}

		disableSSL := cfg.Global.DisableSSL
		if cmd.Flags().Changed("disable-ssl") {
			disableSSL = f.disableSSL
		}
		engine.SetDisableSSL(disableSSL)

		if f.rate {
			return printRateLimit(cmd)
		}

		if f.downloadAll {
			return engine.DownloadConfigRepositories(cfg)
		}

		if len(args) == 0 {
			fmt.Println("no target given")
			return cmd.Help()
		}
		if f.remove {
			return uninstallPackage(cmd, args[0], f.from)
		}

		target, inlineTag, hasInlineTag := splitTargetTag(args[0])
		opts, err := optionsForTarget(cfg, cmd, f, target)
		if err != nil {
			return err
		}
		if hasInlineTag && !cmd.Flags().Changed("tag") {
			opts.Tag = inlineTag
		}
		return engine.Run(target, opts)
	}
}

func rateRunE(f *rootFlags) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(f.config)
		if err != nil {
			return err
		}
		if err := configureGithubToken(cfg); err != nil {
			return err
		}
		disableSSL := cfg.Global.DisableSSL
		if cmd.Flags().Changed("disable-ssl") {
			disableSSL = f.disableSSL
		}
		engine.SetDisableSSL(disableSSL)
		return printRateLimit(cmd)
	}
}

func configureGithubToken(cfg *config.Config) error {
	if cfg.Global.GithubToken != "" && os.Getenv("XGET_GITHUB_TOKEN") == "" {
		return os.Setenv("XGET_GITHUB_TOKEN", cfg.Global.GithubToken)
	}
	return nil
}

func printRateLimit(cmd *cobra.Command) error {
	rdat, err := getRateLimit()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), rdat)
	if !engine.GithubTokenConfigured() {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Tip: set XGET_GITHUB_TOKEN, GITHUB_TOKEN, or EGET_GITHUB_TOKEN to a GitHub token for higher API rate limits.")
	}
	return nil
}

// configOptionsForTarget resolves the global section followed by the matching
// repository section, without applying any CLI flags.
func configOptionsForTarget(cfg *config.Config, target string) (options.Flags, error) {
	opts := options.Flags{
		Source:      cfg.Global.Source,
		System:      cfg.Global.System,
		All:         cfg.Global.All,
		Ignore:      cfg.Global.Ignore,
		Quiet:       cfg.Global.Quiet,
		DLOnly:      cfg.Global.DownloadOnly,
		UpgradeOnly: cfg.Global.UpgradeOnly,
		Hash:        cfg.Global.ShowHash,
		DisableSSL:  cfg.Global.DisableSSL,
		SourceType:  cfg.Global.SourceType,
	}

	if cfg.Global.File != "" {
		opts.ExtractFile = cfg.Global.File
	}
	if cfg.Global.Target != "" {
		expanded, err := home.Expand(cfg.Global.Target)
		if err != nil {
			return options.Flags{}, err
		}
		opts.Output = expanded
	}

	if repo, ok := cfg.Repositories[target]; ok {
		opts.All = repo.All
		opts.Asset = repo.AssetFilters
		opts.Ignore = repo.Ignore
		opts.DLOnly = repo.DownloadOnly
		opts.ExtractFile = repo.File
		opts.Hash = repo.ShowHash
		opts.Prerelease = repo.Prerelease
		opts.Quiet = repo.Quiet
		opts.Source = repo.Source
		opts.System = repo.System
		opts.Tag = repo.Tag
		opts.UpgradeOnly = repo.UpgradeOnly
		opts.Verify = repo.Verify
		opts.DisableSSL = repo.DisableSSL
		opts.SourceType = repo.SourceType

		expanded, err := home.Expand(repo.Target)
		if err != nil {
			return options.Flags{}, err
		}
		opts.Output = expanded
	}

	return opts, nil
}

func optionsForTarget(cfg *config.Config, cmd *cobra.Command, f *rootFlags, target string) (options.Flags, error) {
	opts, err := configOptionsForTarget(cfg, target)
	if err != nil {
		return options.Flags{}, err
	}

	if cmd.Flags().Changed("tag") {
		opts.Tag = f.tag
	}
	if cmd.Flags().Changed("pre-release") {
		opts.Prerelease = f.prerelease
	}
	if cmd.Flags().Changed("source") {
		opts.Source = f.source
	}
	if cmd.Flags().Changed("to") {
		expanded, err := home.Expand(f.output)
		if err != nil {
			return options.Flags{}, err
		}
		opts.Output = expanded
	}
	if cmd.Flags().Changed("system") {
		opts.System = f.system
	}
	if cmd.Flags().Changed("file") {
		opts.ExtractFile = f.extractFile
	}
	if cmd.Flags().Changed("all") {
		opts.All = f.all
	}
	if cmd.Flags().Changed("quiet") {
		opts.Quiet = f.quiet
	}
	if cmd.Flags().Changed("download-only") {
		opts.DLOnly = f.dlOnly
	}
	if cmd.Flags().Changed("upgrade-only") {
		opts.UpgradeOnly = f.upgradeOnly
	}
	if cmd.Flags().Changed("asset") {
		opts.Asset = f.asset
	}
	if cmd.Flags().Changed("ignore") {
		opts.Ignore = f.ignore
	}
	if cmd.Flags().Changed("sha256") {
		opts.Hash = f.hash
	}
	if cmd.Flags().Changed("verify") || cmd.Flags().Changed("verify-sha256") {
		opts.Verify = f.verify
	}
	if cmd.Flags().Changed("remove") {
		opts.Remove = f.remove
	}
	if cmd.Flags().Changed("disable-ssl") {
		opts.DisableSSL = f.disableSSL
	}

	// Template substitution only applies to asset/ignore matchers sourced from the
	// config file; explicit --asset/--ignore flags are used verbatim.
	systemForTemplate := opts.System
	if systemForTemplate == "" {
		systemForTemplate = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if !cmd.Flags().Changed("asset") {
		opts.Asset = config.SubstituteTemplateVarsSlice(opts.Asset, systemForTemplate)
	}
	if !cmd.Flags().Changed("ignore") {
		opts.Ignore = config.SubstituteTemplateVarsSlice(opts.Ignore, systemForTemplate)
	}

	return opts, nil
}
