---
layout: default
title: 🔄 Self-Update
nav_order: 6
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Self-update command

Update xget to the latest release. By default, the command updates the active
xget installation.

```text
xget self-update [flags]
```

## Flags

| Flag | Description |
| --- | --- |
| `-c, --config FILE` | Use the specified configuration file. |
| `-h, --help` | Show self-update help. |
| `-t, --tag TAG` | Install the specified tag instead of the latest release; allows downgrades. |
| `--to PATH` | Update only the copy installed to this tracked location. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |

## Examples

```bash
# Update xget to the latest release.
xget self-update

# Install a specific xget release (including downgrades).
xget self-update --tag v1.2.3

# Update only the copy tracked at a particular location.
xget self-update --to ~/.local/bin
```