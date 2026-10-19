---
title: 🧩 Asset Filtering
nav_order: 3
layout: default
parent: 🧭 Usage
---

<!-- markdownlint-disable MD022 MD025 -->
# Asset Filtering
{: .no_toc }

xget can narrow the assets it considers by using `--asset` and `--ignore` matchers. These filters operate before system detection and can be repeated multiple times.

## Matching styles

`--asset` supports these forms:

- Literal matcher: `--asset .zip`
- Literal anti-match: `--asset '^musl'` or `--asset 'not:musl'`
- Regex matcher: `--asset '~^tool_.*\.zip$'`, `--asset '=~^tool_.*\.zip$'`, or `--asset 're:^tool_.*\.zip$'`
- Regex anti-match: `--asset '^~.*\.sbom\.json$'`, `--asset 'not:~.*\.sbom\.json$'`, or `--asset 'not:re:.*\.sbom\.json$'`

`--ignore` always excludes matches and accepts the same general matcher styles:

- `--ignore '.sbom.json'`
- `--ignore '~\.zip\.sbom\.json$'`
- `--ignore '=~\.zip\.sbom\.json$'`
- `--ignore 're:\.zip\.sbom\.json$'`
- `--ignore 'not:arm64'` (inverted ignore; keeps only matches for `arm64`)

> Patterns beginning with `~` are treated as regex. Patterns beginning with `^` are treated as negative by default. If you need a literal value that starts with either symbol, escape it or use `text:`.
<!-- -->

> [!IMPORTANT]
> **Shell quoting caveat:** always quote matchers that start with `~` (for example
> `--ignore '~\.sha512$'`). Unquoted, most shells — including PowerShell, bash, and
> zsh — expand a leading `~` to your home directory *before* xget ever sees the
> argument, silently turning your regex into an unrelated path and disabling the
> filter. Single quotes are safest since they also prevent `$` from being treated
> as variable interpolation.

Examples:

```bash
xget tacocontent/ironstate --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$'
```

```bash
xget --asset '~~literal-starts-with-tilde'
xget --asset '^^literal-starts-with-caret'
xget --asset 'text:~literal-starts-with-tilde'
```

## Detection and matching behavior

- `--asset` matchers are applied before system detection.
- Literal matches prefer exact basename match, then substring match.
- Negative forms use `^` or `not:`.
- Regex forms use `~`, `=~`, or `re:` prefixes.
- Escaping rules allow literal matching for leading `~` and `^` characters.

This gives you a flexible way to choose assets like `.zip` variants, skip provenance files, or prefer only architecture-specific binaries.

## Template Variables

{% raw %}

The `--file`, `--asset`, and `--binary` (alias `--name`) arguments support
template variables, as do their `file`, `asset_filters`, and `binary`
configuration settings and configured `ignore` lists:

- `{{.OS}}` - Replaced with the target operating system (e.g., `linux`, `darwin`, `windows`)
- `{{.Arch}}` - Replaced with the target architecture (e.g., `amd64`, `arm64`, `386`)

The values are determined by:

1. The `--system` flag if provided (e.g., `--system linux/amd64`)
2. The effective `system` setting from configuration
3. The runtime system (detected automatically from your current OS/architecture)

With `--system all`, templates use the runtime OS and architecture. Architecture
names are Go architecture names (`amd64`, not `x64`; `386`, not `x86`), so archive
paths must use those names for a template to match.

Select files inside the archive for the target architecture:

```shell
xget install gerardog/gsudo --file="{{.Arch}}/*.exe"
```

Use `--all` to extract every matching file without prompting if the glob matches
more than one file. To filter the release asset and rename a single extracted file:

```shell
xget install owner/tool --system windows/amd64 --asset="{{.OS}}_{{.Arch}}.zip" --file="{{.Arch}}/tool.exe" --binary="tool-{{.OS}}-{{.Arch}}"
```

Example:

```toml
["junegunn/fzf"]
asset_filters = ["{{.OS}}_{{.Arch}}.tar.gz"]
```

``` yaml
"junegunn/fzf":
  asset_filters:
    - "{{.OS}}_{{.Arch}}.tar.gz"
```

When running `xget junegunn/fzf` on a Linux AMD64 system, this will match assets containing `linux_amd64.tar.gz`.
When running `xget junegunn/fzf --system darwin/arm64`, it will match assets containing `darwin_arm64.tar.gz`.

Substitution happens after CLI/config precedence is resolved and before asset
matching, archive glob matching, and binary-name validation. Upgrades also expand
these variables in stored options. `--ignore` values passed on the command line
remain literal with respect to templates; configured `ignore` values are expanded.

{% endraw %}

For more examples, see [Usage](index) and [Configuration](../configuration).
