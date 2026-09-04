package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/camalot/xget/internal/home"
	"github.com/camalot/xget/internal/installed"
	"github.com/spf13/cobra"
)

func newUninstallCommand(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "uninstall PACKAGE",
		Aliases:       []string{"remove"},
		Short:         "Remove an installed package",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallPackage(cmd, args[0], f.from)
		},
	}
	cmd.Flags().StringVar(&f.from, "from", "", "directory to remove an untracked target from")
	return cmd
}

func uninstallPackage(cmd *cobra.Command, target, from string) error {
	storePath, err := installed.DefaultPath()
	if err != nil {
		return err
	}
	store, err := installed.Load(storePath)
	if err != nil {
		return err
	}

	packageTarget, _, _ := splitTargetTag(target)
	if pkg, ok := findInstalledPackage(installed.SortedPackages(store), packageTarget); ok {
		if err := removeInstalledFiles(pkg); err != nil {
			return err
		}
		delete(store.Packages, pkg.Key())
		if err := installed.Save(storePath, store); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled `%s`\n", pkg.Name)
		return nil
	}

	return removeUntrackedTarget(cmd, target, from)
}

func removeInstalledFiles(pkg installed.Package) error {
	for _, extractedFile := range pkg.ExtractedFiles {
		path := extractedFile
		if !filepath.IsAbs(path) && pkg.InstallLocation != "" {
			path = filepath.Join(pkg.InstallLocation, path)
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func removeUntrackedTarget(cmd *cobra.Command, target, from string) error {
	directory := from
	if directory == "" {
		directory = os.Getenv("XGET_BIN")
	}
	if directory == "" {
		directory = "."
	}
	expanded, err := home.Expand(directory)
	if err != nil {
		return err
	}
	name, err := safeUninstallBaseName(target)
	if err != nil {
		return err
	}
	path := filepath.Join(expanded, name)
	// #nosec G703 -- name is a validated basename joined to the user-selected directory.
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s is not installed and %s was not found", target, path)
		}
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed `%s`\n", path)
	return nil
}

func safeUninstallBaseName(target string) (string, error) {
	if target == "" || filepath.IsAbs(target) {
		return "", fmt.Errorf("invalid target %q", target)
	}
	clean := filepath.Clean(target)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid target %q", target)
	}
	name := filepath.Base(clean)
	if name == "." || name == ".." || strings.ContainsRune(name, os.PathSeparator) {
		return "", fmt.Errorf("invalid target %q", target)
	}
	return name, nil
}
