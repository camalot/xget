package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/camalot/xget/internal/installed"
	"github.com/spf13/cobra"
)

func TestPrintInstalledPackagesUsesTableFormat(t *testing.T) {
	cmd := &cobra.Command{}
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)

	printInstalledPackages(cmd, []installed.Package{
		{
			Name:            "nektos/act",
			InstallLocation: "~/.local/bin",
			InstalledAt:     time.Date(2026, 9, 1, 17, 10, 28, 0, time.UTC),
			RefreshedAt:     time.Date(2026, 9, 1, 17, 10, 28, 0, time.UTC),
			InstalledTag:    "v0.2.89",
			CurrentTag:      "v0.2.89",
			Source:          "GitHub",
		},
	})

	got := buf.String()
	if !strings.Contains(got, "PACKAGE") || !strings.Contains(got, "TAG/VERSION") || !strings.Contains(got, "LATEST") || !strings.Contains(got, "LOCATION") {
		t.Fatalf("expected table header, got:\n%s", got)
	}
	if !strings.Contains(got, "github:nektos/act") || !strings.Contains(got, "v0.2.89") || !strings.Contains(got, "~/.local/bin") {
		t.Fatalf("expected installed package row, got:\n%s", got)
	}
	if strings.Contains(got, "Installed at:") || strings.Contains(got, "Download URL:") {
		t.Fatalf("expected summary table, got detail output:\n%s", got)
	}
}
