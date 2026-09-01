package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
	"github.com/camalot/xget/internal/installed"
	"github.com/spf13/cobra"
)

type listFlags struct {
	installed bool
	config    string
}

func newListCommand() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:   "list [TARGET]",
		Short: "List available or installed packages",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.installed {
				return listInstalled(cmd, args)
			}

			cfg, err := config.Load(f.config)
			if err != nil {
				return err
			}
			if cfg.Global.GithubToken != "" && os.Getenv("XGET_GITHUB_TOKEN") == "" {
				if err := os.Setenv("XGET_GITHUB_TOKEN", cfg.Global.GithubToken); err != nil {
					return err
				}
			}
			if len(args) == 0 {
				return listConfigured(cmd, cfg)
			}

			target := args[0]
			opts, err := optionsForTarget(cfg, cmd, &rootFlags{}, target)
			if err != nil {
				return err
			}
			assets, err := engine.ListAvailable(target, opts)
			if err != nil {
				return err
			}
			for _, asset := range assets {
				cmd.Println(asset)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&f.installed, "installed", false, "show installed package metadata")
	cmd.Flags().StringVarP(&f.config, "config", "c", "", "path to the config file to use")
	return cmd
}

func listConfigured(cmd *cobra.Command, cfg *config.Config) error {
	if len(cfg.Repositories) == 0 {
		cmd.Println("no configured packages")
		return nil
	}
	names := make([]string, 0, len(cfg.Repositories))
	for name := range cfg.Repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		cmd.Println(name)
	}
	return nil
}

func listInstalled(cmd *cobra.Command, args []string) error {
	storePath, err := installed.DefaultPath()
	if err != nil {
		return err
	}
	store, err := installed.Load(storePath)
	if err != nil {
		return err
	}
	if err := refreshInstalledStore(storePath, store); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not refresh installed metadata: %v\n", err)
	}
	packages := installed.SortedPackages(store)
	if len(args) > 0 {
		pkg, ok := findInstalledPackage(packages, args[0])
		if !ok {
			return fmt.Errorf("%s is not installed", args[0])
		}
		printInstalledPackages(cmd, []installed.Package{pkg})
		return nil
	}
	if len(packages) == 0 {
		cmd.Println("no installed packages")
		return nil
	}
	printInstalledPackages(cmd, packages)
	return nil
}

func refreshInstalledStore(storePath string, store *installed.Store) error {
	changed := false
	for name, pkg := range store.Packages {
		refreshed, err := engine.RefreshInstalledPackage(pkg)
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
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "PACKAGE\tTAG/VERSION\tLATEST\tLOCATION")
	_, _ = fmt.Fprintln(w, "-------------------------------------------------------------------")
	for _, pkg := range packages {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", pkg.Key(), pkg.InstalledTag, pkg.CurrentTag, displayLocation(pkg.InstallLocation))
	}
	_ = w.Flush()
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
