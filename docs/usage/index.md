---
title: 🧭 Usage
nav_order: 2
layout: default
---

<!-- markdownlint-disable MD022 MD025 -->
# xget Usage
{: .no_toc }

`xget` downloads and extracts pre-built binaries from GitHub releases. The target can be:

- a GitHub repository such as `owner/repo`,
- a direct URL to a release asset, or
- a local file to extract from.

If a GitHub repository is provided, xget will search the latest release for assets that look like a binary for your system. Append `@TAG` to select a release directly, such as `eza-community/eza@v0.23.5`; this is equivalent to `--tag TAG`. If a direct URL is provided, xget skips detection and downloads the file directly. If a local file is provided, xget extracts its contents without any network call.

Use `@latest` or `--tag latest` to explicitly select the latest stable release. With `--pre-release`, `latest` selects the newest release regardless of whether it is stable or a prerelease.

Use `xget list owner/repo` to show up to ten recent releases with their name, tag, and publication date. Add `--pre-release` to include prereleases. `xget list --installed` also shows when each package was last installed or upgraded.

Rows with a newer available version are yellow in `xget list --installed` and `xget upgrade`. Pass `--no-color` to either command for plain output.

## Examples

```bash
xget install eza-community/eza --tag v0.23.5 --to ~/.local/bin --asset .zip --ignore .sbom.json
xget zyedidia/micro --tag nightly
xget slavaGanzin/await@2.1.0
xget eza-community/eza --tag v0.23.5
xget eza-community/eza@latest --pre-release
xget jgm/pandoc --to /usr/local/bin
xget junegunn/fzf
xget neovim/neovim
xget ogham/exa --asset ^musl
xget tacocontent/ironstate --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$'
xget tacocontent/ironstate --asset 're:\.zip$' --ignore 'not:re:\.zip\.sbom\.json$'
xget --system darwin/amd64 sharkdp/fd
xget BurntSushi/ripgrep
xget -f xget.1 camalot/xget
xget zachjs/sv2v
xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
xget --all --file '*' ActivityWatch/activitywatch
xget list camalot/xget
xget list --installed
xget list camalot/xget --installed
xget upgrade
xget upgrade camalot/xget
xget upgrade --all
```

## Basic command syntax

Use `xget install TARGET` to install a package. The original `xget TARGET`
form remains supported and runs the same installation command for backwards
compatibility.

```text
Download pre-built binaries from GitHub releases

Usage:
  xget [TARGET] [flags]
  xget install TARGET [flags]
  xget [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Get, set, and edit xget configuration values
  help        Help about any command
  install     Download and install a pre-built binary from GitHub releases
  list        List available or installed packages
  rate        Show GitHub API rate limiting information
  uninstall   Remove an installed package
  upgrade     List and apply available upgrades for installed packages
  version     Print the xget version

Flags:
      --all                      extract all candidate files
  -a, --asset strings            filter assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not:, escapes: ~~ and ^^, explicit literal: text: (for example ^musl, not:~.*\.sbom\.json$, text:~literal); quote patterns starting with ~ so your shell doesn't expand it to a home directory path
  -c, --config string            path to the config file to use
  -k, --disable-ssl              disable SSL verification for download requests
  -D, --download-all             download all projects defined in the config file
  -d, --download-only            stop after downloading the asset (no extraction)
  -f, --file string              glob to select files for extraction
  -h, --help                     help for xget
      --ignore strings           exclude assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not: (inverts ignore), escapes: ~~ and ^^, explicit literal: text:; can be specified multiple times; quote patterns starting with ~ so your shell doesn't expand it to a home directory path
      --non-interactive          fail instead of prompting when user input is required
      --pre-release              include pre-releases when fetching the latest version
  -q, --quiet                    only print essential output
      --rate                     show GitHub API rate limiting information
      --from string              directory to remove an untracked target from
    -r, --remove                   uninstall the target package
      --sha256                   show the SHA-256 hash of the downloaded asset
      --source                   download the source code for the target repo instead of a release
  -s, --system string            target system to download for (use all for all choices)
  -t, --tag string               tagged release to use instead of latest
      --to string                move to given location after extracting
      --upgrade-only             only download if release is more recent than current version
      --verify string[="auto"]   verify the downloaded asset checksum; pass a hash or use --verify with no value to use GitHub's published SHA256 when available
      --verify-sha256 string     verify the downloaded asset checksum against the one provided
  -v, --version                  version for xget

Use "xget [command] --help" for more information about a command.
```

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `1` | General error. |
| `16` | An interactive action was required while running with `--non-interactive`. |

## Extract behavior

When installing an executable, xget will place it in the current directory by default. If the environment variable `XGET_BIN` is non-empty, xget will place the executable there instead.

Directories may also be specified as files to extract. When that happens, xget extracts everything inside the directory. For example:

```bash
xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
```

If xget downloads an asset called `xxx`, and there is also a matching checksum asset such as `xxx.sha256` or `xxx.sha256sum`, xget will automatically verify the hash before extracting it.

## Installed package tracking

Successful installs are recorded in `~/.config/xget/.xget.installed.yml`. Records include source type, repo or URL, install location, install and refresh timestamps, selected asset, download URL, extracted files, effective options, SHA-256, and installed/current tag when release metadata is available. Records are keyed as `<source>:<repo-or-url>`, such as `github:nektos/act`.

Use `xget list TARGET` to list assets that can be installed for a target. Use `xget list --installed` to show all installed records, or `xget list TARGET --installed` to show one installed package.

Use `xget uninstall TARGET` or `xget remove TARGET` to remove a tracked package's extracted files and record. The backwards-compatible `xget TARGET --remove` form does the same. When no installed record matches, xget removes the target basename from `$XGET_BIN`, the current directory, or a directory selected with `--from`.

## GitHub rate limits

GitHub limits unauthenticated API requests to 60 per hour. Use `xget rate` or the backwards-compatible `xget --rate` to inspect the active limit. When no token is configured, xget writes setup guidance to stderr; it does the same after GitHub responds with a rate-limit status.

To make more requests, set a personal access token in `GITHUB_TOKEN`, `EGET_GITHUB_TOKEN`, or `XGET_GITHUB_TOKEN`; `XGET_GITHUB_TOKEN` takes precedence. xget loads environment files from the current directory in this order: `.secrets`, alphabetically sorted `*.secrets`, `.env`, then alphabetically sorted `*.env`. Earlier files win, and values already set by the shell always take precedence.

You can also provide the token value by reading it from a file via `@/path/to/file`.

For more details, see the [Configuration](../configuration) section and the [FAQ](faq) page.
