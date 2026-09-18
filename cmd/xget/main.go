package main

import (
	"fmt"
	"os"

	"github.com/camalot/xget/internal/cli"
	"github.com/camalot/xget/internal/config"
	"github.com/camalot/xget/internal/engine"
)

func main() {
	if err := engine.RemovePreviousExecutable(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
