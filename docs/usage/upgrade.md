---
title: "Upgrading Packages"
nav_order: 2
layout: default
parent: 🧭 Usage
---

<!-- markdownlint-disable-next-line MD025 MD022 -->
# Upgrading Packages
{: .no_toc }

`xget upgrade` reports and applies newer releases for the packages recorded in
`~/.config/xget/.xget.installed.yml`.

<!-- markdownlint-disable-next-line MD025 MD022 -->
## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Listing available upgrades

```bash
xget upgrade
```

Every installed package is looked up against its source, the installed metadata
store is refreshed with the newest release tag, and packages with a newer release
are listed:

Available upgrades are shown in yellow. Pass `--no-color` for plain output.

```text
Name                 Version  Available  Source
-----------------------------------------------
bschaatsbergen/cidr  v2.2.0   v2.3.0     GitHub

1 upgrade available.
```

Only packages with an available upgrade are listed. When there are none, xget
prints `No available upgrades.`

An upgrade is available when the newest release is *newer* than the installed
one. Tags are compared as semantic versions, including prerelease ordering, so
`v2.0.0-beta` → `v2.0.0` is an upgrade but `v2.0.0` → `v2.0.0-beta` is not. Tags
that are not semver-shaped fall back to a plain difference check.

Packages installed from a direct URL or a local file are skipped: there is no
release list to query, so no upgrade can be determined.

## Pinned packages

A package whose stored options include a `tag` was installed pinned to that exact
release. These are never upgraded by `--all` and are listed separately:

```text
The following packages have an upgrade available, but require explicit targeting for upgrade:
Name                 Version  Available  Source
-----------------------------------------------
bschaatsbergen/cidr  v2.2.0   v2.3.0     GitHub
```

Upgrading one requires naming it:

```bash
xget upgrade bschaatsbergen/cidr
```

The package stays pinned afterward; the stored `tag` is updated to the newer
release.

## Upgrading

```bash
xget upgrade <package>   # upgrade one package
xget upgrade --all       # upgrade everything that is not pinned
```

`<package>` accepts the full name (`bschaatsbergen/cidr`), the store key
(`github:bschaatsbergen/cidr`), or the bare repository name (`cidr`).

If a package is already current, xget reports it and exits successfully. During
`--all`, a package that fails to upgrade is reported and the remaining packages
are still attempted; xget exits with status `1` at the end.

## Options used for the upgrade

The upgrade re-runs the download with the options resolved in this order, each
layer overriding the one before it:

1. The `global` section of the config file.
2. The matching `"owner/repo"` section of the config file.
3. The options stored for the package at install time.

Two options are deliberately never applied:

- **`tag`** — applying the stored tag would pin the download to the version that
  is already installed, so the newer release tag is used instead.
- **`upgrade_only`** — the upgrade decision has already been made from release
  metadata, so the additional file-timestamp check would only cause the download
  to be skipped.

Both remain untouched in the installed metadata store after the upgrade.

If neither the config nor the stored options set an output location, the
package's recorded `install_location` is used.

Use `--config <file>` to resolve the config layers from a specific file.

## Refreshing without upgrading

Running `xget upgrade` with no arguments only refreshes metadata and prints the
report; nothing is downloaded. `xget list --installed` performs the same refresh
and prints the full installed package table.
