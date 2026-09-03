---
title: 📄 Configuration
nav_order: 3
layout: default
---

xget supports a variety of configuration options to customize its behavior. You can configure xget using a configuration file, command-line arguments, or environment variables.

For backwards compatibility with eget, xget supports both TOML and YAML configuration files. The same options are available in either format and can be layered by CLI flags, per-repository settings, and global settings.

For information on where xget looks for configuration files, see the [Configuration Loading and Precedence](configuration/loading-and-precedence) page.

## Example configuration

### TOML

```toml
[global]
  quiet = false
  show_hash = false
  upgrade_only = true
  source = "GitHub"
  ignore = ["~\\.sbom\\.json$"]
  target = "./test"

["zyedidia/micro"]
  upgrade_only = false
  show_hash = true
  asset_filters = ["static", ".tar.gz"]
  ignore = ["not:arm64"]
  target = "~/.local/bin/micro"
```

### YAML

```yaml
global:
  quiet: false
  show_hash: false
  upgrade_only: true
  source: GitHub
  ignore:
    - "~\\.sbom\\.json$"
  target: "./test"

"zyedidia/micro":
  upgrade_only: false
  show_hash: true
  asset_filters:
    - "static"
    - ".tar.gz"
  ignore:
    - "not:arm64"
  target: "~/.local/bin/micro"
```

With the configuration above, the following command downloads the latest release of `micro`:

```bash
xget zyedidia/micro
```

Without the config, you would typically need a command like:

```bash
export XGET_GITHUB_TOKEN=ghp_1234567890 && \
xget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz
```

## Available settings - global section

> [!IMPORTANT]
> `github_token` is supported for backwards compatibility with `eget`, but storing a token in a config file is not recommended. Prefer `XGET_GITHUB_TOKEN`, `GITHUB_TOKEN`, or `@/path/to/file`. xget will warn if `github_token` is set in a config file.

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| `github_token` | N/A | GitHub API token to use for requests | `""` |
| `all` | `--all` | Whether to extract all candidate files. | `false` |
| `download_only` | `--download-only` | Stop after downloading the asset without extraction. | `false` |
| `download_source` | `--source` | Download the source code for the repo instead of a release. | `false` |
| `file` | `--file` | Glob to select files for extraction. | `*` |
| `ignore` | `--ignore` | Asset matchers to exclude. | `[]` |
| `pre_release` | `--pre-release` | Include pre-releases when fetching the latest version. | `false` |
| `quiet` | `--quiet` | Print only essential output. | `false` |
| `show_hash` | `--sha256` | Show the SHA-256 hash of the downloaded asset. | `false` |
| `source` | N/A | Source provider metadata to record for installs. Defaults to `GitHub` for GitHub targets and `URL` for direct URLs/local files. | `""` |
| `system` | `--system` | Target system to download for. | `all` |
| `target` | `--to` | Directory to move downloaded files to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Only download if the release is newer than the current installed version. | `false` |
| `disable_ssl` | `--disable-ssl` | Disable SSL certificate verification for downloads. | `false` |

## Available settings - repository sections

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| `all` | `--all` | Extract all candidate files. | `false` |
| `asset_filters` | `--asset` | Array of asset matchers. | `[]` |
| `download_only` | `--download-only` | Stop after downloading the asset without extraction. | `false` |
| `download_source` | `--source` | Download the source code for the repo instead of a release. | `false` |
| `file` | `--file` | Glob to select files for extraction. | `*` |
| `ignore` | `--ignore` | Array of asset matchers to exclude. | `[]` |
| `pre_release` | `--pre-release` | Include pre-releases when fetching the latest version. | global value |
| `quiet` | `--quiet` | Print only essential output. | `false` |
| `show_hash` | `--sha256` | Show the SHA-256 hash of the downloaded asset. | `false` |
| `source` | N/A | Source provider metadata to record for installs. Defaults to `GitHub` for GitHub targets and `URL` for direct URLs/local files. | global value |
| `system` | `--system` | Target system to download for. | `all` |
| `target` | `--to` | Directory to move downloaded files to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Only download if the release is newer than the current installed version. | `false` |
| `verify_sha256` | `--verify-sha256` / `--verify` | Verify the asset hash against a provided hash. | `""` |
| `disable_ssl` | `--disable-ssl` | Disable SSL certificate verification for downloads. | `false` |

## Repository and global precedence

The precedence order is:

1. CLI flags.
2. Repository section values (`"owner/repo"`).
3. Global section values (`global`).
4. Built-in defaults.

This means the same config file can set a default target for all repositories and then override it for a single repo in a more specific section.

```toml
[global]
target = "~/bin"

["zyedidia/micro"]
target = "~/.local/bin"
```

The equivalent YAML form is:

```yaml
global:
  target: "~/bin"

"zyedidia/micro":
  target: "~/.local/bin"
```

For the file lookup order and environment-variable behavior, see [Configuration Loading and Precedence](configuration/loading-and-precedence).
