---
title: 🧩 Asset Filtering
nav_order: 2
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

For more examples, see [Usage](index) and [Configuration](../configuration).
