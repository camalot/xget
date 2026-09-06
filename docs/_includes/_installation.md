
<!-- markdownlint-disable MD041 -->

## How to get xget

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget upgrade camalot/xget`.

### Quick-install script

> [!NOTE]
> The quick-install scripts will automatically download the sha256 checksum of the install script and verify the script before executing it.

### Bash

``` shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option:

``` shell
curl -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash -s -- -d /usr/local/bin
```

### PowerShell

``` powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

> [!NOTE]
> The powershell script can also specify the installation directory with the `-InstallDir` parameter.

### Homebrew

``` shell
brew install camalot/scoop/xget
```

<!-- ### Chocolatey

``` shell
choco install xget
``` -->

### Scoop

``` shell
scoop bucket add https://github.com/camalot/scoop
scoop install xget
```

<!-- ### Winget

``` shell
winget install camalot.xget
``` -->

### eget

``` shell
eget camalot/xget --asset '^.sbom.json'
```

### Pre-built binaries

Pre-built binaries are available on the [releases](https://github.com/camalot/xget/releases) page.

### From source

Install the latest released version:

``` shell
go install github.com/camalot/xget/cmd/xget@latest
```

You can run directly via `go run`

``` shell
go run github.com/camalot/xget/cmd/xget@latest
```

or install from HEAD:

``` shell
git clone https://github.com/camalot/xget
cd xget
go build ./...
```

A man page can be generated from the source tree with `pandoc`:

``` shell
pandoc docs/_man/xget.md -s -t man -o xget.1
```

You can also use `xget` to download the man page: `xget -f xget.1 camalot/xget`.

### GitHub Action

For GitHub Actions workflows, use [xget-action](https://github.com/camalot/xget-action)
to install `xget` (with binary caching) and run it in a single step:

``` yaml
- name: Install a tool with xget
  uses: camalot/xget-action@v1
  with:
    package: junegunn/fzf
```

You can also use the action to just install `xget` on the GitHub Actions runner without installing any packages. You can then use `xget` in subsequent steps to install other tools as needed.

``` yaml
- uses: camalot/xget-action@v1
- shell: bash
  run: |
    xget install eza-community/eza --to ~/.local/bin
    xget install junegunn/fzf --to ~/.local/bin
    xget install bschaatsbergen/cidr --to ~/.local/bin
```

The `xget-action` uses the `--non-interactive` flag by default to ensure that installations do not prompt for user input, which is suitable for automated CI/CD environments.

See the [xget-action](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.
