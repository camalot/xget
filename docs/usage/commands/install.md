---
layout: default
title: 🚀 Install
nav_order: 3
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Install command

Download and install a pre-built binary from a GitHub or GitLab release, a
direct URL, or a local archive. A repository target can include a release tag
as `owner/repo@TAG`; alternatively, use `--tag`.

```text
xget install TARGET [flags]
```

The legacy form `xget TARGET [flags]` is also supported. See [usage](../index)
for details about targets, extraction, and tracked installs.

## Flags

| Flag | Description |
| --- | --- |
| `--all` | Extract all candidate files without prompting. |
| `-a, --asset MATCHER` | Filter release assets by matcher. See [asset filtering](../asset-filtering). |
| `--binary NAME`, `--name NAME` | Rename one extracted file, keeping its extension; ignored when multiple files are extracted. |
| `--clobber_global.asset_filters BOOL` | Replace inherited `asset_filters` values instead of merging. |
| `--clobber_global.ignore BOOL` | Replace inherited `ignore` values instead of merging. |
| `-c, --config FILE` | Use the specified configuration file. |
| `-k, --disable-ssl` | Disable SSL verification for download requests. |
| `-D, --download-all` | Download all projects defined in the configuration file. |
| `-d, --download-only` | Stop after downloading the asset; do not extract it. |
| `-f, --file GLOB` | Select files within the downloaded archive with a glob. |
| `--from PATH` | Install location to remove a package from, or directory for removing an untracked target. |
| `-h, --help` | Show install help. |
| `--ignore MATCHER` | Exclude matching assets; may be specified more than once. |
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

`--clobber_global.asset_filters` and `--clobber_global.ignore` control whether
the corresponding array flags replace or merge configured values. See
[configuration array merging](../../configuration/index#merging-arrays-with-clobber_global).

## Examples

```bash
# Install the latest matching release.
xget install eza-community/eza

# Install a specific release to a directory.
xget install eza-community/eza --tag v0.23.5 --to ~/.local/bin

# Select a system and filter the release asset.
xget install sharkdp/fd --system linux/amd64 --asset musl

# Extract a named file from a direct URL archive.
xget install https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5

# Keep the installation out of xget's installed package store.
xget install jgm/pandoc --untracked
```