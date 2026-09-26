---
title: 📝 Examples
nav_order: 5
layout: default
parent: 🧭 Usage
---

<!-- markdownlint-disable MD022 MD025 -->
# Examples
{: .no_toc }

---

## Install

### Install Example

```bash
xget install mikefarah/yq
```

### Advanced Install Example

```bash
xget install mikefarah/yq --tag v4.30.8 --to /usr/local/bin/yq --ignore .sbom.json --asset 
```

## List

### List Example

```bash
xget list camalot/xget
```

``` text
NAME                                TAG      DATE
-------------------------------------------------------
v4.53.6                             v4.53.6  2026-08-20
v4.53.4                             v4.53.4  2026-08-19
v4.53.3                             v4.53.3  2026-06-06
v4.53.2                             v4.53.2  2026-04-17
v4.52.5                             v4.52.5  2026-03-25
v4.52.4                             v4.52.4  2026-02-14
v4.52.2                             v4.52.2  2026-01-31
v4.52.1 - TOML roundtrip and more!  v4.52.1  2026-01-31
v4.50.1 - HCL!                      v4.50.1  2025-12-14
v4.49.2                             v4.49.2  2025-11-25
```

### Installed Example

```bash
xget list --installed
```

``` text
PACKAGE                         TAG/VERSION        LATEST             LOCATION      INSTALLED/UPDATED
-----------------------------------------------------------------------------------------------------
github:BurntSushi/ripgrep       15.2.0             15.2.0             ~\.local\bin  2026-09-20
github:GBerghoff/envdiff        v0.2.0             v0.2.0             ~\.local\bin  2026-09-01
github:akinoiro/ssh-list        v1.5.1             v1.5.1             ~\.local\bin  2026-09-07
github:ankddev/envfetch         v2.1.2             v2.1.2             ~\.local\bin  2026-09-01
github:anomalyco/opencode       v1.18.32           v1.18.32           ~\.local\bin  2026-09-24
github:antonmedv/fx             39.2.0             39.2.0             ~\.local\bin  2026-09-01
github:byawitz/ggh              v0.1.5             v0.1.5             ~\.local\bin  2026-09-07
github:caarlos0/svu             v3.4.1             v3.4.1             ~\.local\bin  2026-09-07
github:mikefarah/yq             v4.53.6            v4.53.6            ~\.local\bin  2026-09-07
```
