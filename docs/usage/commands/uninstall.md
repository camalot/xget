---
layout: default
title: 🗑️ Uninstall
nav_order: 7
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Uninstall command

Remove an installed package's extracted files and its tracking record. The
command accepts `uninstall` and `remove` as aliases.

```text
xget uninstall PACKAGE [flags]
xget remove PACKAGE [flags]
```

When a package is tracked in several locations, specify `--from` to remove one
copy or `--all` to remove it from every tracked location. If the target is not
tracked, xget removes its basename from the selected directory.

## Flags

| Flag | Description |
| --- | --- |
| `-a, --all` | Remove the package from every tracked install location. |
| `--from PATH` | Select the install location to remove from, or the directory containing an untracked target. |
| `-h, --help` | Show uninstall help. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |

## Examples

```bash
# Remove the tracked installation of a package.
xget uninstall jgm/pandoc

# Remove only the copy installed in a particular directory.
xget uninstall jgm/pandoc --from ~/.local/bin

# Remove every tracked copy.
xget uninstall jgm/pandoc --all

# Use the remove alias.
xget remove jgm/pandoc
```