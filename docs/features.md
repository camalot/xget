---
layout: default
title: 🏠 Features
nav_order: 2
---

<!-- markdownlint-disable MD022 MD025 -->
# Features
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

xget is a fork of [zyedidia/eget](https://github.com/zyedidia/eget) that keeps the
original workflow — point it at a repository, get a binary — while adding package
tracking, upgrades, more release sources, and richer asset matching.

xget v2 is backwards compatible with eget v1: existing commands and `.eget.toml`
configuration files keep working, so switching is usually a matter of replacing
`eget` with `xget` (or symlinking `xget` to `eget`).

## Feature list

### Release sources

- GitHub and GitLab releases, direct URLs, and local archives.
- Named [source profiles](configuration) for self-hosted or alternate hosts, with
  per-profile `host`, `api_url`, `token_env`, and `token` settings.
- `PROFILE:owner/repo` shorthand, for example `xget install gitlab:gitlab-org/cli`.
- Nested GitLab namespaces (`group/subgroup/project`).

### Installed package management

- Successful installs are recorded in `~/.config/xget/.xget.installed.yml` with the
  asset, download URL, extracted files, options, checksum, and tag.
- The same package can be tracked in multiple locations and managed independently.
- `xget list --installed` shows installed packages, versions, locations, and dates.
- `xget upgrade` reports and applies available upgrades, including `--all`, pinned
  tags, and per-location upgrades.
- `xget self-update` updates xget itself.
- `xget uninstall` removes tracked files and the matching record.
- `--untracked` opts a single install out of the store.

### Asset selection

- `--asset` matchers with literal, regex (`~`, `=~`, `re:`), and negative
  (`^`, `not:`) forms, plus escapes (`~~`, `^^`) and explicit literals (`text:`).
- `--ignore` exclusion matchers using the same syntax, including inverted ignores.
- `{% raw %}{{.OS}}{% endraw %}` and `{% raw %}{{.Arch}}{% endraw %}` template
  variables in configured `asset_filters` and `ignore` values.
- Tag selection with `owner/repo@TAG`, `--tag`, `@latest`, and monorepo tag
  fragment matching.

### Verification and safety

- Automatic verification of `*.sha256` / `*.sha256sum` sibling assets.
- `--verify` uses GitHub's published SHA-256 when available; `--verify-sha256`
  checks against a value you provide; `--sha256` prints the hash.
- Warnings for plaintext tokens in configuration, with `@/path/to/token` file
  references and `disable_token_warning` opt-out.
- `--non-interactive` fails (exit code `16`) instead of prompting, for CI use.

### Configuration

- TOML **or** YAML configuration, with a documented search order and `.eget.*`
  fallbacks for compatibility.
- `xget config get/set/pop/clear/list/path/edit` for command-line management.
- Global and per-repository sections with defined precedence.
- Environment files (`.secrets`, `*.secrets`, `.env`, `*.env`) loaded per run.
- Tokens from `XGET_GITHUB_TOKEN`, `EGET_GITHUB_TOKEN`, or `GITHUB_TOKEN`.

### Distribution and tooling

- Subcommands (`install`, `list`, `upgrade`, `uninstall`, `config`, `rate`,
  `version`, `completion`) alongside the original flag-only invocation.
- Shell completions for bash, zsh, fish, and PowerShell.
- Packages for Scoop, Homebrew, `.deb`, `.rpm`, `.apk`, and Arch, plus install
  scripts for bash and PowerShell.
- [GitHub Action](usage/action) with binary caching.
- Release checksums and SBOM assets.

## Comparison matrix

| Feature | xget | eget |
| --- | :---: | :---: |
| GitHub releases | ✅ | ✅ |
| GitLab releases | ✅ | ❌ |
| Named source profiles / self-hosted hosts | ✅ | ❌ |
| Direct URL and local file targets | ✅ | ✅ |
| Installed package tracking | ✅ | ❌ |
| Tracking one package in several locations | ✅ | ❌ |
| `upgrade` / `update` command | ✅ | ⚠️ `--upgrade-only` only |
| `self-update` | ✅ | ⚠️ reinstall `zyedidia/eget` |
| `uninstall` with tracked file removal | ✅ | ⚠️ `--remove` by file name |
| List releases for a repository | ✅ | ❌ |
| Subcommands | ✅ | ❌ flags only |
| Shell completion | ✅ | ❌ |
| Literal asset filters | ✅ | ✅ |
| Regex asset filters | ✅ | ❌ |
| Negative / anti-match filters | ✅ | ⚠️ literal `^` prefix |
| `--ignore` exclusion list | ✅ | ❌ |
| OS/Arch templates in configured filters | ✅ | ❌ |
| Tag selection (`--tag`, tag fragments, monorepos) | ✅ | ✅ |
| `owner/repo@TAG` shorthand | ✅ | ❌ |
| Pre-release selection | ✅ | ✅ |
| Automatic `.sha256` sibling verification | ✅ | ✅ |
| `--verify-sha256` against a provided hash | ✅ | ✅ |
| `--verify` using the published GitHub checksum | ✅ | ❌ |
| Plaintext token warnings and token files | ✅ | ⚠️ token files only |
| TOML configuration | ✅ | ✅ |
| YAML configuration | ✅ | ❌ |
| `.eget.toml` compatibility | ✅ | ✅ |
| Config management from the CLI (`xget config`) | ✅ | ❌ |
| `.env` / `.secrets` loading | ✅ | ❌ |
| Non-interactive mode with dedicated exit code | ✅ | ❌ |
| Download all configured projects (`-D`) | ✅ | ✅ |
| Source archive download (`--source`) | ✅ | ✅ |
| `--disable-ssl` | ✅ | ✅ |
| API rate limit reporting | ✅ | ✅ |
| Linux packages (`.deb`, `.rpm`, `.apk`, Arch) | ✅ | ❌ |
| Scoop / Homebrew | ✅ | ⚠️ Homebrew, Chocolatey |
| GitHub Action | ✅ | ❌ |
| SBOM and release checksums | ✅ | ❌ |

Legend: ✅ supported, ⚠️ partial or different approach, ❌ not available.

## Migrating from eget

- Replace `eget` with `xget` in commands, or symlink `xget` as `eget`.
- Existing `.eget.toml` files are still discovered and honored.
- `EGET_GITHUB_TOKEN` and `EGET_CONFIG` are still read; the `XGET_` variants take
  precedence.
- `eget owner/repo --remove` maps to `xget uninstall owner/repo`.

See the [installation guide](installation) to get started, or the [FAQ](faq) for
more background on the fork.
