package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

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
		printInstalledPackage(cmd, pkg)
		return nil
	}
	if len(packages) == 0 {
		cmd.Println("no installed packages")
		return nil
	}
	for i, pkg := range packages {
		if i > 0 {
			cmd.Println("")
		}
		printInstalledPackage(cmd, pkg)
	}
	return nil
}

func refreshInstalledStore(storePath string, store *installed.Store) error {
	changed := false
	for name, pkg := range store.Packages {
		refreshed, err := engine.RefreshInstalledPackage(pkg)
		if err != nil {
			return err
		}
		if !refreshed.RefreshedAt.Equal(pkg.RefreshedAt) || refreshed.CurrentTag != pkg.CurrentTag || refreshed.CurrentVersion != pkg.CurrentVersion {
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
		if pkg.Name == target || pkg.Repo == target || strings.HasSuffix(pkg.Repo, "/"+target) {
			return pkg, true
		}
	}
	return installed.Package{}, false
}

func printInstalledPackage(cmd *cobra.Command, pkg installed.Package) {
	cmd.Printf("Name: %s\n", pkg.Name)
	cmd.Printf("Source: %s\n", pkg.Source)
	cmd.Printf("Repo: %s\n", pkg.Repo)
	cmd.Printf("Installed tag: %s\n", pkg.InstalledTag)
	cmd.Printf("Current tag: %s\n", pkg.CurrentTag)
	cmd.Printf("Installed version: %s\n", pkg.InstalledVersion)
	cmd.Printf("Current version: %s\n", pkg.CurrentVersion)
	cmd.Printf("Installed at: %s\n", pkg.InstalledAt.Format(timeFormat))
	cmd.Printf("Refreshed at: %s\n", pkg.RefreshedAt.Format(timeFormat))
	cmd.Printf("Install location: %s\n", pkg.InstallLocation)
	cmd.Printf("Download URL: %s\n", pkg.DownloadURL)
	cmd.Printf("Asset: %s\n", pkg.Asset)
	cmd.Printf("SHA256: %s\n", pkg.SHA256)
	if len(pkg.ExtractedFiles) > 0 {
		cmd.Printf("Extracted files: %s\n", strings.Join(pkg.ExtractedFiles, ", "))
	}
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
