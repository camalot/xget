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

If a GitHub repository is provided, xget will search the latest release for assets that look like a binary for your system. If a direct URL is provided, xget skips detection and downloads the file directly. If a local file is provided, xget extracts its contents without any network call.

## Examples

```bash
xget zyedidia/micro --tag nightly
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
```

## Basic command syntax

```text
Download pre-built binaries from GitHub releases

Usage:
  xget [TARGET] [flags]
  xget [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the xget version

Flags:
      --all                      extract all candidate files
  -a, --asset strings            filter assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not:, escapes: ~~ and ^^, explicit literal: text: (for example ^musl, not:~.*\.sbom\.json$, text:~literal)
  -c, --config string            path to the config file to use
  -k, --disable-ssl              disable SSL verification for download requests
  -D, --download-all             download all projects defined in the config file
  -d, --download-only            stop after downloading the asset (no extraction)
  -f, --file string              glob to select files for extraction
  -h, --help                     help for xget
      --ignore strings           exclude assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not: (inverts ignore), escapes: ~~ and ^^, explicit literal: text:; can be specified multiple times
      --pre-release              include pre-releases when fetching the latest version
  -q, --quiet                    only print essential output
      --rate                     show GitHub API rate limiting information
  -r, --remove                   remove the given file from $XGET_BIN or the current directory
      --sha256                   show the SHA-256 hash of the downloaded asset
      --source                   download the source code for the target repo instead of a release
  -s, --system string            target system to download for (use all for all choices)
  -t, --tag string               tagged release to use instead of latest
      --to string                move to given location after extracting
      --upgrade-only             only download if release is more recent than current version
      --verify string[="auto"]   verify the downloaded asset checksum; pass a hash or use --verify with no value to use GitHub's published SHA256 when available
      --verify-sha256 string     verify the downloaded asset checksum against the one provided

Use "xget [command] --help" for more information about a command.
```

## Extract behavior

When installing an executable, xget will place it in the current directory by default. If the environment variable `XGET_BIN` is non-empty, xget will place the executable there instead.

Directories may also be specified as files to extract. When that happens, xget extracts everything inside the directory. For example:

```bash
xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
```

If xget downloads an asset called `xxx`, and there is also a matching checksum asset such as `xxx.sha256` or `xxx.sha256sum`, xget will automatically verify the hash before extracting it.

## GitHub rate limits

GitHub limits unauthenticated API requests to 60 per hour. If you want to make more requests, use a personal access token and set it in either `GITHUB_TOKEN`, `EGET_GITHUB_TOKEN`, or `XGET_GITHUB_TOKEN` before running xget. If more than one is set, `XGET_GITHUB_TOKEN` takes precedence.

You can also provide the token value by reading it from a file via `@/path/to/file`.

For more details, see the [Configuration](../configuration) section and the [FAQ](faq) page.
