---
layout: default
title: 🖥️ Commands
nav_order: 1
has_children: true
parent: 🧭 Usage
---

# Commands

Use `xget install TARGET` to download and install a binary. For backward
compatibility, `xget TARGET` is equivalent. Other commands manage configuration,
list installed packages and releases, update packages, and generate shell
completions.

## Available commands

| Command | Description |
| --- | --- |
| [`xget completion`](completion) | Generate shell completion scripts. |
| [`xget config`](config) | Get, set, and edit configuration values. |
| [`xget install`](install) | Download and install a binary. |
| [`xget list`](list) | List available or installed packages. |
| [`xget rate`](rate) | Show GitHub API rate-limit information. |
| [`xget self-update`](self-update) | Update xget itself. |
| [`xget uninstall`](uninstall) | Remove an installed package. |
| [`xget upgrade`](upgrade) | List and apply available package upgrades. |

Run `xget help` to see this command list, or `xget COMMAND --help` for details
about a specific command.

## Root command flags

These flags are available with the backward-compatible `xget TARGET` install
form. The flags used with `xget install TARGET` are documented on the
[install page](install). Flags that apply to a particular operation are also
available on their corresponding command.

| Flag | Description |
| --- | --- |
| `--all` | Extract all candidate files. |
| `-a, --asset MATCHER` | Filter release assets by matcher. See [asset filtering](../asset-filtering). |
| `--binary NAME` | Rename one extracted file, keeping its extension; ignored if multiple files are extracted. |
| `--clobber_global.asset_filters` | Replace inherited `asset_filters` values instead of merging. |
| `--clobber_global.ignore` | Replace inherited `ignore` values instead of merging. |
| `-c, --config FILE` | Use the specified configuration file. |
| `-k, --disable-ssl` | Disable SSL verification for download requests. |
| `-D, --download-all` | Download all projects defined in the configuration file. |
| `-d, --download-only` | Stop after downloading the asset; do not extract it. |
| `-f, --file GLOB` | Select files for extraction with a glob. |
| `--from PATH` | Install location to remove a package from, or directory for removing an untracked target. |
| `-h, --help` | Show help for xget. |
| `--ignore MATCHER` | Exclude matching assets; may be specified more than once. |
| `--name NAME` | Alias for `--binary`. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |
| `--pre-release` | Include prereleases when selecting the latest version. |
| `--provider NAME` | Select the release source profile (default: `github`). |
| `-q, --quiet` | Print only essential output. |
| `--rate` | Show GitHub API rate-limit information. |
| `-r, --remove` | Uninstall the target package. |
| `--sha256` | Show the downloaded asset's SHA-256 hash. |
| `--source` | Download the repository source code instead of a release asset. |
| `-s, --system OS/ARCH` | Select the target system; use `all` to show all choices. |
| `-t, --tag TAG` | Select a release tag instead of the latest release. |
| `--to PATH` | Move extracted files to the specified location. |
| `--untracked` | Do not record this install; existing records are left untouched. |
| `--upgrade-only` | Download only if the release is newer than the installed version. |
| `--verify[=HASH]` | Verify the asset checksum against a hash, or use GitHub's published SHA-256 when available if no value is given. |
| `--verify-sha256 HASH` | Verify the asset checksum against the given hash. |
| `-v, --version` | Print the xget version. |

Array flags replace inherited `asset_filters` and `ignore` values by default.
Set the corresponding `--clobber_global.*` flag to `false` to merge values
instead. See [configuration array merging](../../configuration/index#merging-arrays-with-clobber_global).

## Examples

```bash
# Show command help and the installed xget version.
xget help
xget --version

# Install with the traditional root-command syntax.
xget sharkdp/fd --system darwin/amd64

# Show GitHub API rate-limit information.
xget --rate
```