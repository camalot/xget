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

xget supports both TOML and YAML configuration formats while maintaining
backward compatibility with original [zyedidia/eget](https://github.com/zyedidia/eget) TOML files.

Supported config filenames:

- `.xget.toml`
- `.xget.yaml` / `.xget.yml`
- `.eget.toml` (backward compatibility)
- `.eget.yaml` / `.eget.yml` (accepted if present)

Search order:

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

Resolution precedence:

1. CLI flags.
2. Repository section values (`"owner/repo"`).
3. Global section values (`global`).
4. Built-in defaults.

Config keys remain unchanged across TOML and YAML. For example, `target`,
`asset_filters`, `download_only`, and `verify_sha256` map to the same flags as
before.
