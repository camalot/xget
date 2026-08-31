
<!-- markdownlint-disable MD041 -->

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget camalot/xget`.

### Quick-install script

``` shell
curl -o xget.sh https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh
shasum -a 256 xget.sh # verify with hash below
echo "TODO: GENERATE CHECKSUM FOR SHELL SCRIPT"
bash xget.sh
```

Or alternatively (less secure):

``` shell
curl https://github.com/camalot/xget/raw/refs/heads/main/install/xget.sh | bash
```

You can then place the downloaded binary in a location on your `$PATH` such as `/usr/local/bin`.

To verify the script, the sha256 checksum is `TODO: GENERATE CHECKSUM FOR SHELL SCRIPT` (use `shasum -a 256 xget.sh` after downloading the script).

One of the reasons to use xget is to avoid running curl into bash, but unfortunately you can't xget xget until you have xget.

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
