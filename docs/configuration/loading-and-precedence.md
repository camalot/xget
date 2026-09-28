---
title: 📄 Loading and Precedence
nav_order: 1
layout: default
parent: ⚙️ Configuration
has_children: false
---

<!-- markdownlint-disable-next-line MD025 MD022 -->
# Configuration Loading and Precedence
{: .no_toc }

---

xget supports both `TOML` and `YAML` configuration formats while maintaining
backward compatibility with original [zyedidia/eget](https://github.com/zyedidia/eget) `TOML` files.

## Supported filenames

- `.xget.toml`
- `.xget.yaml`
- `.xget.yml`
- `.eget.toml` (backward compatibility)
- `.eget.yaml` / `.eget.yml` (accepted if present)

## Search order

The configuration loader checks the following locations in descending priority
(TOML, then YML, then YAML for `.xget` before `.eget` at each location):

1. Path from `--config` if set.
2. Path from `XGET_CONFIG` if set.
   - Path from `EGET_CONFIG` if set (compatibility with original [eget](https://github.com/zyedidia/eget)).
3. Current directory: `./.xget.<ext>`.
   - `.eget.<ext>` is also checked for backward compatibility.
4. User home: `~/.xget.<ext>` (Windows: `%USERPROFILE%/.xget.<ext>`).
   - `.eget.<ext>` is also checked for backward compatibility.
5. OS config path: `$XDG_CONFIG_HOME/xget/.xget.<ext>` or `~/.config/xget/.xget.<ext>`.
   - `.eget.<ext>` is also checked for backward compatibility.
6. Windows: `%LOCALAPPDATA%/xget/.xget.<ext>`.
   - `.eget.<ext>` is also checked for backward compatibility.

All existing files are merged from lowest to highest priority. An explicitly
configured path (`--config`, `XGET_CONFIG`, or `EGET_CONFIG`) is the highest
layer; it does not discard files found in the standard locations. Scalar values
and lists in a higher-priority file replace lower-priority values, while
unspecified keys, repository sections, and source profile fields are retained.
The path reported by xget is the highest-priority file.

To use only the highest-priority file, set `global.config_merge` to `false` in
that file (default: `true`). A lower-priority file cannot disable merging for a
higher-priority one. For example, if `~/.config/xget/.xget.yml` contains
`global.target: ~/.local/bin` and `./.xget.yml` contains
`global.target: /usr/local/bin`, the effective target is `/usr/local/bin`.

```yaml
global:
   config_merge: false
```

For TOML, use `config_merge = false` under `[global]`.

## Resolution precedence

When multiple values are available, xget resolves them in this order:

1. CLI flags.
2. Repository section values (`"owner/repo"`).
3. Global section values (`global`).
4. Built-in defaults.

This means an explicit command-line flag always wins, followed by repo-specific settings, followed by global defaults.

## Environment files

Before running a command, xget loads environment files from the current
directory. This lets a local token be supplied without exporting it manually.
Files are loaded in descending priority:

1. `.secrets`
2. Alphabetically sorted `*.secrets`
3. `.env`
4. Alphabetically sorted `*.env`

The first file that defines a variable wins. Variables already present in the
parent shell are not replaced and therefore have the highest priority.

## Inheritance

A repository section inherits any setting it does not define from the `global` section, so `global` acts as the default for every repository.

The exceptions are `asset_filters`, `binary`, `pre_release`, `tag`, and `verify_sha256`, which are repository-only settings and are never inherited. `github_token` and `config_merge` are global-only. `xget_update_check` is global-only and defaults to `true`. `source` selects a named profile from the `sources` section and defaults to the built-in `github` profile.

Within a source profile, token environment variables are checked in the order
listed by `token_env`, followed by the profile's `token`. GitHub profiles use
legacy `global.github_token` only as a final fallback.

```yaml
global:
  system: linux/amd64
  target: "~/bin"

"zyedidia/micro":
  target: "~/.local/bin/micro"
```

Here `zyedidia/micro` overrides `target` but still uses the global `system` value.

## Backward compatibility

xget retains compatibility with earlier [zyedidia/eget](https://github.com/zyedidia/eget) behavior and filenames. In addition to `.eget.toml`, xget will also read `.eget.yml` and `.eget.yaml` if they are present. The keys remain the same across TOML and YAML; for example, `target`, `asset_filters`, `download_only`, and `verify_sha256` map to the same runtime flags.

## Example

```toml
[global]
target = "~/bin"

["zyedidia/micro"]
target = "~/.local/bin/micro"
```

This makes all repositories install to `~/bin` by default, while `zyedidia/micro` overrides the target to its own installation directory.
