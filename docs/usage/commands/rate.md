---
layout: default
title: 💹 Rate
nav_order: 5
parent: 🖥️ Commands
grand_parent: 🧭 Usage
---

# Rate command

Show the active GitHub API rate limit, including the remaining request quota
and reset time. This is useful when GitHub API requests are being rate limited.

```text
xget rate [flags]
```

## Flags

| Flag | Description |
| --- | --- |
| `-c, --config FILE` | Use the specified configuration file. |
| `-k, --disable-ssl` | Disable SSL verification for download requests. |
| `-h, --help` | Show rate help. |
| `--non-interactive` | Fail instead of prompting when input is required. Implied when stdin is not a terminal. |

## Examples

```bash
# Show GitHub API rate-limit information.
xget rate

# Use a specific configuration file.
xget rate --config ./xget.yml
```

The backward-compatible `xget --rate` option is also available; see the
[root command flags](index#root-command-flags).