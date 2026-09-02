package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set via -ldflags at build time (see .goreleaser.yaml); "dev" locally.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the xget version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.Println(fmt.Sprintf("xget %s (commit %s, built %s) - https://github.com/camalot/xget", version, commit, date))
			return nil
		},
	}
}
