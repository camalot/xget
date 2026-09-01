
<!-- markdownlint-disable MD041 -->

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget camalot/xget`.

### Quick-install script

### Bash

``` shell
curl -o xget.sh https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh
shasum -a 256 xget.sh # verify with hash below
curl -o xget.sh.sha256 https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh.sha256
shasum -a 256 --check xget.sh.sha256
bash xget.sh
```

``` shell
curl https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh | bash
```

The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option:

``` shell
curl -o xget.sh https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh | bash -s -- -d /usr/local/bin
```

### PowerShell

``` powershell
iwr https://github.com/camalot/xget/raw/refs/heads/main/install/xget.ps1 | iex
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
scoop bucket add xget https://github.com/camalot/xget
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
