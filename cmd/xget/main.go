package main

import (
	"fmt"
	"os"

	"github.com/camalot/xget/internal/cli"
	"github.com/camalot/xget/internal/config"
)

func main() {
	if err := config.LoadDotenvFiles(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := cli.Execute(); err != nil {
		if !cli.IsSilent(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(cli.ExitCodeFor(err))
	}
}
