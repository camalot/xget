package main

import (
	"fmt"
	"os"
	"os/exec"
)

func fileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	xget := os.Getenv("TEST_XGET")

	must(run(xget, "--system", "linux/amd64", "jgm/pandoc"))
	must(fileExists("pandoc"))

	must(run(xget, "zyedidia/micro", "--tag", "nightly", "--asset", "osx"))
	must(fileExists("micro"))

	must(run(xget, "--asset", "nvim.appimage", "--to", "nvim", "neovim/neovim"))
	must(fileExists("nvim"))

	must(run(xget, "--system", "darwin/amd64", "sharkdp/fd"))
	must(fileExists("fd"))

	must(run(xget, "--system", "windows/amd64", "--asset", "windows-gnu", "BurntSushi/ripgrep"))
	must(fileExists("rg.exe"))

	must(run(xget, "-f", "xget.1", "camalot/xget"))
	must(fileExists("xget.1"))

	fmt.Println("ALL TESTS PASS")
}
