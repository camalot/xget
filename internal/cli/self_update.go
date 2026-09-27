package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/installed"
	"github.com/camalot/xget/internal/lib/constants"
	"github.com/camalot/xget/internal/semver"
	"github.com/spf13/cobra"
)

const (
	xgetRepo           = "camalot/xget"
	selfCheckInterval  = 24 * time.Hour
	selfCheckTimeout   = 5 * time.Second
	developmentVersion = "dev"
)

// Indirected for testing.
var (
	runningExecutable = func() (string, error) {
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.EvalSymlinks(executable)
	}
	now = time.Now
)

func xgetPackage() installed.Package {
	return installed.Package{Name: xgetRepo, Source: constants.ProviderGithub}
}

func latestXgetTag(cfg *config.Config) (string, error) {
	pkg := xgetPackage()
	opts, err := resolveInstalledOptions(cfg, pkg)
	if err != nil {
		return "", err
	}
	refreshed, err := refreshPackage(pkg, opts)
	if err != nil {
		return "", err
	}
	return refreshed.CurrentTag, nil
}

// selfUpdateRunningExecutable updates the running binary in place when xget is
// not tracked in the installed store, leaving it untracked.
func selfUpdateRunningExecutable(cmd *cobra.Command, f *upgradeFlags, cfg *config.Config) error {
	out := cmd.OutOrStdout()
	executable, err := runningExecutable()
	if err != nil {
		return fmt.Errorf("locate running xget executable: %w", err)
	}

	tag := f.tag
	if tag == "" {
		latest, err := latestXgetTag(cfg)
		if err != nil {
			return err
		}
		if !semver.IsUpgrade(version, latest) {
			_, _ = fmt.Fprintf(out, "xget is already up to date (%s).\n", version)
			return nil
		}
		tag = latest
		_, _ = fmt.Fprintf(out, "Upgrading xget from %s to %s in %s\n", version, tag, displayLocation(executable))
	} else {
		_, _ = fmt.Fprintf(out, "Installing xget %s in %s\n", tag, displayLocation(executable))
	}

	opts, err := resolveInstalledOptions(cfg, xgetPackage())
	if err != nil {
		return err
	}
	opts.Tag = tag
	opts.Output = executable
	opts.Untracked = true
	return runEngine(xgetRepo, opts)
}

// checkForSelfUpdate prints a notice when a newer xget release exists. Unless
// force is set, it queries at most once per selfCheckInterval and records the
// attempt in the installed store.
func checkForSelfUpdate(cmd *cobra.Command, force bool) {
	if version == developmentVersion {
		return
	}
	if !force {
		storePath, err := installed.DefaultPath()
		if err != nil {
			return
		}
		store, err := installed.Load(storePath)
		if err != nil {
			return
		}
		current := now()
		if current.Sub(store.SelfUpdate.LastChecked) < selfCheckInterval {
			return
		}
		// Record the attempt up front so an offline or failing check is not retried on every run.
		store.SelfUpdate.LastChecked = current.UTC()
		if err := installed.Save(storePath, store); err != nil {
			return
		}
	}

	result := make(chan string, 1)
	go func() {
		latest, err := latestXgetTag(config.Default())
		if err != nil {
			latest = ""
		}
		result <- latest
	}()

	var latest string
	select {
	case latest = <-result:
	case <-time.After(selfCheckTimeout):
		return
	}
	if semver.IsUpgrade(version, latest) {
		printUpdateNotice(cmd.ErrOrStderr(), version, latest)
	}
}

func printUpdateNotice(w io.Writer, current, latest string) {
	lines := []string{
		"",
		"A new version of xget is available!",
		fmt.Sprintf("%s → %s", current, latest),
		"",
		"Run `xget self-update` to update.",
		"",
	}
	width := 0
	for _, line := range lines {
		width = max(width, utf8.RuneCountInString(line))
	}
	border := strings.Repeat("─", width+6)
	var b strings.Builder
	b.WriteString("\n╭" + border + "╮\n")
	for _, line := range lines {
		pad := strings.Repeat(" ", width-utf8.RuneCountInString(line))
		b.WriteString("│   " + line + pad + "   │\n")
	}
	b.WriteString("╰" + border + "╯\n")
	_, _ = io.WriteString(w, b.String())
}

// selfCheckMode reports whether cmd should run the self-update check and
// whether it should bypass the daily interval.
func selfCheckMode(cmd *cobra.Command) (run, force bool) {
	for c := cmd; c != nil; c = c.Parent() {
		name := c.Name()
		if strings.HasPrefix(name, "__") || name == "completion" || name == "help" || name == "self-update" {
			return false, false
		}
	}
	switch cmd.Name() {
	case "upgrade":
		return true, true
	case "list":
		listInstalled, _ := cmd.Flags().GetBool("installed")
		return true, listInstalled
	}
	return true, false
}
