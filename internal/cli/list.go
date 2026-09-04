package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
	"github.com/camalot/xget/internal/installed"
	"github.com/camalot/xget/internal/semver"
	"github.com/spf13/cobra"
)

type listFlags struct {
	installed  bool
	prerelease bool
	noColor    bool
	config     string
}

var listReleases = engine.ListReleases

func newListCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list [TARGET]",
		Short: "List available or installed packages",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(f.config)
			if err != nil {
				return err
			}
			if cfg.Global.GithubToken != "" && os.Getenv("XGET_GITHUB_TOKEN") == "" {
				if err := os.Setenv("XGET_GITHUB_TOKEN", cfg.Global.GithubToken); err != nil {
					return err
				}
			}
			if f.installed {
				return listInstalled(cmd, cfg, args, f.noColor)
			}
			if len(args) == 0 {
				return listConfigured(cmd, cfg)
			}

			target := args[0]
			opts, err := optionsForTarget(cfg, cmd, &rootFlags{}, target)
			if err != nil {
				return err
			}
			if f.prerelease {
				opts.Prerelease = true
			}
			releases, err := listReleases(target, opts.Prerelease)
			if err != nil {
				return err
			}
			printAvailableReleases(cmd, releases)
			return nil
		},
	}
	cmd.Flags().BoolVar(&f.installed, "installed", false, "show installed package metadata")
	cmd.Flags().BoolVar(&f.prerelease, "pre-release", false, "include pre-releases")
	cmd.Flags().BoolVar(&f.noColor, "no-color", false, "disable colored output")
	cmd.Flags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
	return cmd
}

func listConfigured(cmd *cobra.Command, cfg *config.Config) error {
	out := cmd.OutOrStdout()
	if len(cfg.Repositories) == 0 {
		if cfg.Path == "" {
			_, _ = fmt.Fprintln(out, "no config file found; run `xget list --installed` to show installed packages")
		} else {
			_, _ = fmt.Fprintf(out, "no packages configured in %s; run `xget list --installed` to show installed packages\n", cfg.Path)
		}
		return nil
	}
	names := make([]string, 0, len(cfg.Repositories))
	for name := range cfg.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		_, _ = fmt.Fprintln(out, name)
	}
	return nil
}

func listInstalled(cmd *cobra.Command, cfg *config.Config, args []string, noColor bool) error {
	storePath, err := installed.DefaultPath()
	if err != nil {
		return err
	}
	store, err := installed.Load(storePath)
	if err != nil {
		return err
	}
	if err := refreshInstalledStore(storePath, store, cfg); err != nil {
		return err
	}
	packages := installed.SortedPackages(store)
	if len(args) > 0 {
		pkg, ok := findInstalledPackage(packages, args[0])
		if !ok {
			return fmt.Errorf("%s is not installed", args[0])
		}
		printInstalledPackagesWithColor(cmd, []installed.Package{pkg}, !noColor)
		return nil
	}
	if len(packages) == 0 {
		cmd.Println("no installed packages")
		return nil
	}
	printInstalledPackagesWithColor(cmd, packages, !noColor)
	return nil
}

func refreshInstalledStore(storePath string, store *installed.Store, cfg *config.Config) error {
	changed := false
	for name, pkg := range store.Packages {
		opts, err := resolveInstalledOptions(cfg, pkg)
		if err != nil {
			return err
		}
		refreshed, err := refreshPackage(pkg, opts)
		if err != nil {
			return err
		}
		if !refreshed.RefreshedAt.Equal(pkg.RefreshedAt) || refreshed.CurrentTag != pkg.CurrentTag {
			store.Packages[name] = refreshed
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return installed.Save(storePath, store)
}

func findInstalledPackage(packages []installed.Package, target string) (installed.Package, bool) {
	for _, pkg := range packages {
		if strings.EqualFold(pkg.Key(), target) || pkg.Name == target || strings.HasSuffix(pkg.Name, "/"+target) {
			return pkg, true
		}
	}
	return installed.Package{}, false
}

func printInstalledPackages(cmd *cobra.Command, packages []installed.Package) {
	printInstalledPackagesWithColor(cmd, packages, false)
}

func printInstalledPackagesWithColor(cmd *cobra.Command, packages []installed.Package, colorUpgrades bool) {
	rows := make([][]string, 0, len(packages))
	coloredRows := make([]bool, 0, len(packages))
	for _, pkg := range packages {
		rows = append(rows, []string{
			pkg.Key(),
			pkg.InstalledTag,
			pkg.CurrentTag,
			displayLocation(installedLocation(pkg)),
			formatDate(pkg.InstalledAt),
		})
		coloredRows = append(coloredRows, colorUpgrades && semver.IsUpgrade(pkg.InstalledTag, pkg.CurrentTag))
	}
	printTableWithRowColors(cmd.OutOrStdout(), []string{"PACKAGE", "TAG/VERSION", "LATEST", "LOCATION", "INSTALLED/UPDATED"}, rows, coloredRows)
}

func printAvailableReleases(cmd *cobra.Command, releases []engine.Release) {
	rows := make([][]string, 0, len(releases))
	for _, release := range releases {
		name := release.Name
		if name == "" {
			name = release.Tag
		}
		rows = append(rows, []string{name, release.Tag, formatDate(release.PublishedAt)})
	}
	printTable(cmd.OutOrStdout(), []string{"NAME", "TAG", "DATE"}, rows)
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}

func installedLocation(pkg installed.Package) string {
	for _, extractedFile := range pkg.ExtractedFiles {
		if samePath(pkg.InstallLocation, extractedFile) {
			return filepath.Dir(pkg.InstallLocation)
		}
	}
	return pkg.InstallLocation
}

func samePath(first, second string) bool {
	if filepath.Clean(first) == filepath.Clean(second) {
		return true
	}
	firstAbsolute, firstErr := filepath.Abs(first)
	secondAbsolute, secondErr := filepath.Abs(second)
	return firstErr == nil && secondErr == nil && strings.EqualFold(firstAbsolute, secondAbsolute)
}

func displayLocation(location string) string {
	if location == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return location
	}
	cleanLocation := filepath.Clean(location)
	cleanHome := filepath.Clean(home)
	if strings.EqualFold(cleanLocation, cleanHome) {
		return "~"
	}
	withSeparator := cleanHome + string(os.PathSeparator)
	if strings.HasPrefix(strings.ToLower(cleanLocation), strings.ToLower(withSeparator)) {
		return "~" + string(os.PathSeparator) + cleanLocation[len(withSeparator):]
	}
	return location
}
