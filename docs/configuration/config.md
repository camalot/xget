---
title: 📄 The config Command
nav_order: 2
layout: default
parent: ⚙️ Configuration
has_children: false
---

<!-- markdownlint-disable-next-line MD025 MD022 -->
# The `config` Command
{: .no_toc }

`xget config` reads and writes configuration values from the command line, in the
same spirit as `git config`.

<!-- markdownlint-disable-next-line MD025 MD022 -->
## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Which file is used

Every `config` subcommand accepts `--config <file>` (`-c`). When it is omitted,
xget resolves the file using the normal
[search order](loading-and-precedence):

1. `XGET_CONFIG` (or `EGET_CONFIG`) if set.
2. The first existing file in the standard candidate locations.

If no configuration file exists anywhere, a new one is created at
`$XDG_CONFIG_HOME/xget/.xget.yml`. When `XDG_CONFIG_HOME` is empty or unset,
`~/.config/xget/.xget.yml` is used. This applies on Linux, macOS, and Windows.

Use `xget config path` to print the file that will be used:

```bash
xget config path
xget config path --config ./.xget.toml
```

## Format

The on-disk format is determined by the extension of the resolved file:

| Extension | Format |
| --------- | ------ |
| `.toml` | TOML |
| `.yml`, `.yaml` | YAML |

Writing to a `.toml` file produces TOML; writing to a `.yml`/`.yaml` file produces
YAML. Any other extension is rejected.

> **Note:** writing rewrites the file from its parsed contents. Comments and the
> original key ordering are not preserved.

## Sections

The first argument to `get`, `set`, `clear`, and `pop` is the section:

- `global` — the `[global]` / `global:` section.
- `<owner>/<repo>` — a repository section, for example `zyedidia/micro`.

See [Configuration](index) for the keys valid in each section. Unknown
keys are rejected with the list of valid keys for that section.

## `xget config set`

```bash
xget config set global <key>=<value> [--config <file>]
xget config set <org>/<repo> <key>=<value> [--config <file>]
```

Scalar keys (strings and booleans) are replaced. List keys — `ignore` and
`asset_filters` — **append**; run the command again to add more entries.
Appending a value that is already present is a no-op.

```bash
xget config set global target=~/bin
xget config set global upgrade_only=true
xget config set zyedidia/micro asset_filters=static
xget config set zyedidia/micro asset_filters=.tar.gz
```

> [!IMPORTANT]
> **Shell quoting caveat:** always quote matchers that start with `~` (for example
> `--ignore '~\.sha512$'`). Unquoted, most shells — including PowerShell, bash, and
> zsh — expand a leading `~` to your home directory *before* xget ever sees the
> argument, silently turning your regex into an unrelated path and disabling the
> filter. Single quotes are safest since they also prevent `$` from being treated
> as variable interpolation.

Boolean values accept `true`, `false`, `t`, `f`, `1`, and `0` (and their
capitalized spellings). Anything else is rejected.

To replace a whole list, clear it first:

```bash
xget config clear zyedidia/micro asset_filters
xget config set zyedidia/micro asset_filters=static
```

## `xget config get`

```bash
xget config get global <key> [--config <file>]
xget config get <org>/<repo> <key> [--config <file>]
```

Prints the raw value from the configuration file. List values are printed one
entry per line. If the section or key is not set, nothing is printed and xget
exits with status `1`.

```bash
xget config get global target
xget config get zyedidia/micro asset_filters
```

## `xget config clear`

```bash
xget config clear global <key> [--config <file>]
xget config clear <org>/<repo> <key> [--config <file>]
```

Removes the key entirely, including all entries of a list key. If the key was not
set, nothing is printed and xget exits with status `1`. Empty repository sections
are left in place, because a section with no keys is still meaningful for
`xget --download-all`.

## `xget config pop`

```bash
xget config pop global <key>=<value> [--config <file>]
xget config pop <org>/<repo> <key>=<value> [--config <file>]
```

Removes a single entry from a list key. When the last entry is removed, the key
itself is removed. For scalar keys, `pop` behaves like `clear` when the current
value matches `<value>`; otherwise it exits with status `1`.

```bash
xget config pop zyedidia/micro asset_filters=.tar.gz
xget config pop global ignore='~\.sbom\.json$'
```

## `xget config list`

```bash
xget config list [--config <file>]
```

Prints every value in the resolved configuration file as `section.key=value`,
with `global` first and the remaining sections sorted alphabetically. List keys
produce one line per entry.

```text
global.target=~/bin
global.upgrade_only=true
zyedidia/micro.asset_filters=static
zyedidia/micro.asset_filters=.tar.gz
```

## `xget config path`

```bash
xget config path [--config <file>]
```

Prints the configuration file path that xget would read from or write to, whether
or not it exists yet.

## `xget config edit`

```bash
xget config edit [--config <file>]
```

Opens the resolved configuration file in an editor, creating it (and its parent
directories) first if it does not exist. After the editor exits, xget re-reads the
file and reports a parse error if the edit made it invalid.

The editor is resolved from the first non-empty value of:

1. `XGET_EDITOR`
2. `VISUAL`
3. `EDITOR`

If none are set, xget falls back to `nano` on Linux and macOS. On Windows it uses
`nano` when it is on `PATH`, otherwise `notepad`.

The value may include arguments, which are passed before the file path (split on
whitespace; quoting is not interpreted).

```bash
XGET_EDITOR=nano xget config edit
EDITOR="code --wait" xget config edit --config ./.xget.toml
```

```powershell
$Env:XGET_EDITOR = "code --wait"
xget config edit
```

## Exit codes

| Situation | Exit code | Output |
| --------- | --------- | ------ |
| Success | `0` | value or confirmation |
| `get`/`clear`/`pop` on an unset key | `1` | none |
| Invalid section, key, value, or file | `1` | error on stderr |
