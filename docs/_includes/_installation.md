
<!-- markdownlint-disable MD041 -->

## How to get xget

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget upgrade camalot/xget`.

### Quick-install script

### Bash

``` shell
curl -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh
shasum -a 256 xget.sh # verify with hash below
curl -o xget.sh.sha256 https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh.sha256
shasum -a 256 --check xget.sh.sha256
bash xget.sh
```

``` shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option:

``` shell
curl -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash -s -- -d /usr/local/bin
```

### PowerShell

The script to install will automatically download the sha256 checksum and verify the script before executing it. If you want to manually verify the checksum, you can download the script and the checksum file separately and run the following command:

```powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 -OutFile xget.ps1
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1.sha256 -OutFile xget.ps1.sha256
$FileHash = Get-FileHash xget.ps1 -Algorithm SHA256
$ExpectedHash = Get-Content xget.ps1.sha256 | ForEach-Object { $_.Split(' ')[0] }
if ($FileHash.Hash -ne $ExpectedHash) {
    Write-Error "Checksum verification failed. Expected: $ExpectedHash, Actual: $($FileHash.Hash)"
} else {
  Write-Output "Checksum verification passed."
}
```

``` powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

<!-- ### Homebrew

``` shell
brew install xget
```
-->

<!-- ### Chocolatey

``` shell
choco install xget
``` -->

<!-- ### Scoop

``` shell
scoop bucket add main
scoop install xget
``` -->

<!-- ### Winget

``` shell
winget install camalot.xget
``` -->

### Pre-built binaries

Pre-built binaries are available on the [releases](https://github.com/camalot/xget/releases) page.

### From source

Install the latest released version:

``` shell
go install github.com/camalot/xget@latest
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

See the [xget-action README](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.
