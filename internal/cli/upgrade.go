package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
	"github.com/camalot/xget/internal/installed"
	"github.com/camalot/xget/internal/options"
	"github.com/camalot/xget/internal/semver"
	"github.com/spf13/cobra"
)

// Indirected for testing.
var (
	refreshPackage = engine.RefreshInstalledPackage
	runEngine      = engine.Run
)

type upgradeFlags struct {
	all     bool
	noColor bool
	config  string
}

type upgradeCandidate struct {
	pkg    installed.Package
	pinned bool
}

func newUpgradeCommand() *cobra.Command {
	f := &upgradeFlags{}

	cmd := &cobra.Command{
		Use:   "upgrade [PACKAGE]",
		Short: "List and apply available upgrades for installed packages",
		Long: "List and apply available upgrades for installed packages.\n\n" +
			"With no arguments, the latest release of every installed package is looked up,\n" +
			"the installed metadata store is refreshed, and available upgrades are listed.\n\n" +
			"Packages pinned to a tag are listed separately and are never upgraded by --all;\n" +
			"they must be named explicitly.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(f.config)
			if err != nil {
				return err
			}
			storePath, err := installed.DefaultPath()
			if err != nil {
				return err
			}
			store, err := installed.Load(storePath)
			if err != nil {
				return err
			}

			if len(args) > 0 {
				if err := refreshNamedPackage(storePath, store, cfg, args[0]); err != nil {
					return err
				}
				return upgradeNamed(cmd, cfg, storePath, store, args[0])
			}
			if err := refreshInstalledStore(storePath, store, cfg); err != nil {
				return err
			}

			upgradable, pinned := upgradeCandidates(store)
			if f.all {
				return upgradeAll(cmd, cfg, storePath, upgradable, pinned, !f.noColor)
			}
			printUpgradeReport(cmd, upgradable, pinned, !f.noColor)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&f.all, "all", "a", false, "upgrade every package with an available upgrade")
	cmd.Flags().BoolVar(&f.noColor, "no-color", false, "disable colored output")
	cmd.Flags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
	return cmd
}

func refreshNamedPackage(storePath string, store *installed.Store, cfg *config.Config, target string) error {
	pkg, ok := findInstalledPackage(installed.SortedPackages(store), target)
	if !ok {
		return fmt.Errorf("%s is not installed", target)
	}
	opts, err := resolveInstalledOptions(cfg, pkg)
	if err != nil {
		return err
	}
	refreshed, err := refreshPackage(pkg, opts)
	if err != nil {
		return err
	}
	if refreshed.RefreshedAt.Equal(pkg.RefreshedAt) && refreshed.CurrentTag == pkg.CurrentTag {
		return nil
	}
	store.Packages[pkg.Key()] = refreshed
	return installed.Save(storePath, store)
}

// upgradeCandidates splits packages with an available upgrade into freely
// upgradable ones and ones pinned to a tag. Packages from sources with no
// queryable release list are skipped, since no upgrade can be determined.
func upgradeCandidates(store *installed.Store) (upgradable, pinned []upgradeCandidate) {
	for _, pkg := range installed.SortedPackages(store) {
		if !strings.EqualFold(pkg.Source, "GitHub") {
			continue
		}
		if !semver.IsUpgrade(pkg.InstalledTag, pkg.CurrentTag) {
			continue
		}
		candidate := upgradeCandidate{pkg: pkg, pinned: pkg.Options.Tag != ""}
		if candidate.pinned {
			pinned = append(pinned, candidate)
			continue
		}
		upgradable = append(upgradable, candidate)
	}
	return upgradable, pinned
}

func printUpgradeReport(cmd *cobra.Command, upgradable, pinned []upgradeCandidate, colorUpgrades bool) {
	out := cmd.OutOrStdout()

	if len(upgradable) == 0 && len(pinned) == 0 {
		_, _ = fmt.Fprintln(out, "No available upgrades.")
		return
	}

	if len(upgradable) > 0 {
		printUpgradeTable(out, upgradable, colorUpgrades)
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintf(out, "%s available.\n", pluralUpgrades(len(upgradable)))
	}

	if len(pinned) > 0 {
		if len(upgradable) > 0 {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintln(out, "The following packages have an upgrade available, but require explicit targeting for upgrade:")
		printUpgradeTable(out, pinned, colorUpgrades)
	}
}

func pluralUpgrades(count int) string {
	if count == 1 {
		return "1 upgrade"
	}
	return fmt.Sprintf("%d upgrades", count)
}

func printUpgradeTable(out io.Writer, candidates []upgradeCandidate, colorUpgrades bool) {
	rows := make([][]string, 0, len(candidates))
	coloredRows := make([]bool, 0, len(candidates))
	for _, candidate := range candidates {
		rows = append(rows, []string{
			candidate.pkg.Name,
			candidate.pkg.InstalledTag,
			candidate.pkg.CurrentTag,
			candidate.pkg.Source,
		})
		coloredRows = append(coloredRows, colorUpgrades)
	}
	printTableWithRowColors(out, []string{"Name", "Version", "Available", "Source"}, rows, coloredRows)
}

func upgradeAll(cmd *cobra.Command, cfg *config.Config, storePath string, upgradable, pinned []upgradeCandidate, colorUpgrades bool) error {
	out := cmd.OutOrStdout()

	if len(upgradable) == 0 {
		if len(pinned) == 0 {
			_, _ = fmt.Fprintln(out, "No available upgrades.")
			return nil
		}
		_, _ = fmt.Fprintln(out, "No packages can be upgraded automatically.")
	}

	failures := []string{}
	for _, candidate := range upgradable {
		_, _ = fmt.Fprintf(out, "Upgrading %s from %s to %s\n", candidate.pkg.Name, candidate.pkg.InstalledTag, candidate.pkg.CurrentTag)
		if err := applyUpgrade(cfg, storePath, candidate.pkg); err != nil {
			_, _ = fmt.Fprintf(out, "failed to upgrade %s: %v\n", candidate.pkg.Name, err)
			failures = append(failures, candidate.pkg.Name)
		}
	}

	if len(pinned) > 0 {
		if len(upgradable) > 0 {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintln(out, "The following packages have an upgrade available, but require explicit targeting for upgrade:")
		printUpgradeTable(out, pinned, colorUpgrades)
	}

	if len(failures) > 0 {
		return fmt.Errorf("failed to upgrade: %s", strings.Join(failures, ", "))
	}
	return nil
}

func upgradeNamed(cmd *cobra.Command, cfg *config.Config, storePath string, store *installed.Store, target string) error {
	out := cmd.OutOrStdout()
	pkg, ok := findInstalledPackage(installed.SortedPackages(store), target)
	if !ok {
		return fmt.Errorf("%s is not installed", target)
	}
	if !strings.EqualFold(pkg.Source, "GitHub") {
		return fmt.Errorf("%s was installed from %s, so no upgrade can be determined", pkg.Name, pkg.Source)
	}
	if !semver.IsUpgrade(pkg.InstalledTag, pkg.CurrentTag) {
		_, _ = fmt.Fprintf(out, "%s is already up to date (%s).\n", pkg.Name, pkg.InstalledTag)
		return nil
	}
	_, _ = fmt.Fprintf(out, "Upgrading %s from %s to %s\n", pkg.Name, pkg.InstalledTag, pkg.CurrentTag)
	return applyUpgrade(cfg, storePath, pkg)
}

// resolveInstalledOptions layers the global config section, then the matching
// repository section, then the options the package was installed with.
//
// The stored tag is never applied because it would pin the download to the
// installed version, and upgrade_only is dropped because the upgrade decision
// has already been made from release metadata.
func resolveInstalledOptions(cfg *config.Config, pkg installed.Package) (options.Flags, error) {
	opts, err := configOptionsForTarget(cfg, pkg.Name)
	if err != nil {
		return options.Flags{}, err
	}
	opts.Tag = ""
	opts.UpgradeOnly = false

	// Absent values are indistinguishable from zero values in the store, so only
	// non-zero stored options override the config.
	stored := pkg.Options
	if stored.Prerelease {
		opts.Prerelease = true
	}
	if stored.DownloadSource {
		opts.Source = true
	}
	if stored.Output != "" {
		opts.Output = stored.Output
	}
	if stored.System != "" {
		opts.System = stored.System
	}
	if stored.ExtractFile != "" {
		opts.ExtractFile = stored.ExtractFile
	}
	if stored.All {
		opts.All = true
	}
	if stored.DownloadOnly {
		opts.DLOnly = true
	}
	if len(stored.Asset) > 0 {
		opts.Asset = stored.Asset
	}
	if len(stored.Ignore) > 0 {
		opts.Ignore = stored.Ignore
	}
	if stored.Verify != "" {
		opts.Verify = stored.Verify
	}
	if opts.SourceType == "" {
		opts.SourceType = pkg.Source
	}
	if opts.Output == "" {
		opts.Output = pkg.InstallLocation
	}

	opts.Asset = config.SubstituteTemplateVarsSlice(opts.Asset, opts.System)
	opts.Ignore = config.SubstituteTemplateVarsSlice(opts.Ignore, opts.System)
	return opts, nil
}

func applyUpgrade(cfg *config.Config, storePath string, pkg installed.Package) error {
	opts, err := resolveInstalledOptions(cfg, pkg)
	if err != nil {
		return err
	}
	opts.Tag = pkg.CurrentTag
	if err := runEngine(pkg.Name, opts); err != nil {
		return err
	}
	return restoreStoredOptions(storePath, pkg)
}

// restoreStoredOptions puts back the options the package was installed with,
// which the upgrade run would otherwise overwrite with its own synthesized
// flags. A pinned package stays pinned, now to the newer tag.
func restoreStoredOptions(storePath string, pkg installed.Package) error {
	store, err := installed.Load(storePath)
	if err != nil {
		return err
	}
	key := pkg.Key()
	current, ok := store.Packages[key]
	if !ok {
		return nil
	}
	restored := pkg.Options
	if restored.Tag != "" {
		restored.Tag = current.InstalledTag
	}
	current.Options = restored
	store.Packages[key] = current
	return installed.Save(storePath, store)
}
