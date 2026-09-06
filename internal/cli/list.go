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
	"github.com/camalot/xget/internal/home"
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
		matches := findInstalledPackages(packages, args[0])
		if len(matches) == 0 {
			return fmt.Errorf("%s is not installed", args[0])
		}
		printInstalledPackagesWithColor(cmd, matches, !noColor)
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
	for key, records := range store.Packages {
		for index, pkg := range records {
			opts, err := resolveInstalledOptions(cfg, pkg)
			if err != nil {
				return err
			}
			refreshed, err := refreshPackage(pkg, opts)
			if err != nil {
				return err
			}
			if !refreshed.RefreshedAt.Equal(pkg.RefreshedAt) || refreshed.CurrentTag != pkg.CurrentTag {
				store.Packages[key][index] = refreshed
				changed = true
			}
		}
	}
	if !changed {
		return nil
	}
	return installed.Save(storePath, store)
}

// findInstalledPackages returns every tracked install location matching target,
// which may be a store key, a full repo name, or a bare package name.
func findInstalledPackages(packages []installed.Package, target string) []installed.Package {
	matches := []installed.Package{}
	for _, pkg := range packages {
		if strings.EqualFold(pkg.Key(), target) || pkg.Name == target || strings.HasSuffix(pkg.Name, "/"+target) {
			matches = append(matches, pkg)
		}
	}
	return matches
}

// selectInstalledLocation narrows matches to the single record installed at
// location. An empty location leaves matches untouched.
func selectInstalledLocation(matches []installed.Package, target, location string) ([]installed.Package, error) {
	if location == "" {
		return matches, nil
	}
	expanded, err := home.Expand(location)
	if err != nil {
		return nil, err
	}
	for _, pkg := range matches {
		if installed.SamePath(installedLocation(pkg), expanded) {
			return []installed.Package{pkg}, nil
		}
	}
	return nil, fmt.Errorf("package %s is not installed to %s", installedTargetName(matches, target), location)
}

// errAmbiguousLocation asks the user to pick one of several tracked locations.
func errAmbiguousLocation(matches []installed.Package, target, flag string) error {
	locations := make([]string, 0, len(matches))
	for _, pkg := range matches {
		locations = append(locations, "  "+displayLocation(installedLocation(pkg)))
	}
	return fmt.Errorf("%s is installed to multiple locations; select one with %s:\n%s",
		installedTargetName(matches, target), flag, strings.Join(locations, "\n"))
}

func installedTargetName(matches []installed.Package, target string) string {
	if len(matches) > 0 && matches[0].Name != "" {
		return matches[0].Name
	}
	return target
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
	return installed.SamePath(first, second)
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
