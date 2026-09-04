package main

import (
	"fmt"
	"os"

	"github.com/camalot/xget/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if !cli.IsSilent(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cli.ExitCodeFor(err))
	}
}
