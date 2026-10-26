---
layout: default
title: ⬆️ Upgrade
nav_order: 8
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Upgrade command

Look up the latest versions of tracked packages and list available upgrades.
With no arguments, xget refreshes installed package metadata and lists
available upgrades. Add a package name to work with a specific installation.

```text
xget upgrade [PACKAGE] [flags]
```

The command also accepts the `update` alias. A package installed to several
locations is upgraded in every out-of-date location unless `--to` selects one.
Packages pinned to a tag are listed separately and are not upgraded by `--all`;
name a pinned package explicitly to upgrade it.

## Flags

| Flag | Description |
| --- | --- |
| `-a, --all` | Upgrade every package with an available upgrade. |
| `-c, --config FILE` | Use the specified configuration file. |
| `-h, --help` | Show upgrade help. |
| `--no-color` | Disable colored output. |
| `-t, --tag TAG` | Install the specified tag instead of the latest release; allows downgrades. |
| `--to PATH` | Upgrade only the copy installed to this tracked location. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |

## Examples

```bash
# Check tracked packages for available upgrades.
xget upgrade

# Upgrade all packages that have an available update.
xget upgrade --all

# Check or upgrade a specific package.
xget upgrade jgm/pandoc

# Upgrade only the copy installed in a particular directory.
xget upgrade jgm/pandoc --to ~/.local/bin

# Install a specific tag, including a downgrade.
xget upgrade jgm/pandoc --tag 3.10
```