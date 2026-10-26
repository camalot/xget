---
layout: default
title: 📝 List
nav_order: 4
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# List command

List recent releases for a target, or show metadata for packages recorded in
the installed package store.

```text
xget list [TARGET] [flags]
```

Without `--installed`, provide a repository or other target to list its
available releases. With `--installed`, the command lists tracked packages;
combine it with `--check` to look up their latest versions.

## Flags

| Flag | Description |
| --- | --- |
| `--check` | Check installed packages for updates. |
| `-c, --config FILE` | Use the specified configuration file. |
| `-h, --help` | Show list help. |
| `--installed` | Show installed package metadata. |
| `--no-color` | Disable colored output. |
| `--pre-release` | Include prereleases when listing available releases. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |

## Examples

```bash
# List recent releases for a repository.
xget list camalot/xget

# List all tracked installations.
xget list --installed

# Check tracked packages for newer releases.
xget list --installed --check --no-color

# List a repository's releases including prereleases.
xget list camalot/xget --pre-release
```