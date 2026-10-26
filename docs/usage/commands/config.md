---
layout: default
title: ⚙️ Config
nav_order: 2
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Configuration commands

The `config` command reads and writes the configuration file selected by
xget's normal [configuration search order](../../configuration/loading-and-precedence).
Use `--config` to choose a specific file. If no file exists, a new one is
created at `$XDG_CONFIG_HOME/xget/.xget.yml` or `~/.config/xget/.xget.yml`.

```text
xget config [flags]
xget config [command]
```

## Subcommands

| Command | Description |
| --- | --- |
| `xget config clear <scope> <key>` | Remove a configuration key. |
| `xget config edit` | Open the configuration file in an editor. |
| `xget config get <scope> <key>` | Print a configuration value. |
| `xget config list` | Print every configured value. |
| `xget config path` | Print the path to the configuration file that would be used. |
| `xget config pop <scope> <key>=<value>` | Remove one value from a list-valued key. |
| `xget config set <scope> <key>=<value>` | Set a configuration value. |

`<scope>` is `global`, a repository such as `owner/repo`, or a source profile
such as `sources.NAME`.

## Flags

| Flag | Description |
| --- | --- |
| `-c, --config FILE` | Use the specified configuration file. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |
| `-h, --help` | Show help for the selected config command. |

The `--config` and `--non-interactive` flags are available on `config` and its
subcommands.

## Examples

```bash
# Find the configuration file xget would use.
xget config path

# Print a configured value or all configured values.
xget config get global system
xget config list

# Set a scalar value or append to a list-valued key.
xget config set global system=linux/amd64
xget config set global 'ignore=*.sha256'

# Remove a key or one matching value from a list.
xget config clear global system
xget config pop global ignore=*.sha256

# Open the configuration file in the configured editor.
xget config edit
```

Scalar keys are replaced by `set`. List-valued keys such as `ignore` and
`asset_filters` append values; repeat `set` to add entries. Use `pop` to remove
one entry or `clear` to remove the entire key. The editor is selected from
`XGET_EDITOR`, `VISUAL`, or `EDITOR`, with `nano` as the default (or Notepad on
Windows when nano is unavailable).