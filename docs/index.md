---
layout: home
title: 🏠 Home
nav_order: 1
---

<!-- markdownlint-disable MD022 MD025 -->
# XGET
{: .no_toc }

> [!NOTE]
> xget is a forked codebase of [zyedidia/eget](https://github.com/zyedidia/eget) focusing on some additional features and improvements. The original project does not seem to be actively maintained.
> The version of xget is starting at v2.0.0 to avoid confusion with the original project.

[![Release](https://img.shields.io/github/release/camalot/xget.svg?label=Release&style=for-the-badge)](https://github.com/camalot/xget/releases)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge&label=license&labelColor=%23555555)](https://github.com/camalot/xget/blob/main/LICENSE)
[![Codecov](https://img.shields.io/codecov/c/github/camalot/xget?style=for-the-badge&label=COVERAGE&logo=codecov&logoColor=white)](https://app.codecov.io/gh/camalot/xget/tree/develop)
[![GitHub Build](https://img.shields.io/github/actions/workflow/status/camalot/xget/.github%2Fworkflows%2Fci.yml?style=for-the-badge&logo=github&label=BUILD)](https://github.com/camalot/xget/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/camalot/xget/.github%2Fworkflows%2Frelease.yml?style=for-the-badge&logo=github&label=RELEASE)](https://github.com/camalot/xget/actions/workflows/release.yml)
[![Issues](https://img.shields.io/github/issues/camalot/xget?style=for-the-badge&logo=github&color=blue)](https://github.com/camalot/xget/issues)

<!-- markdownlint-disable MD033 -->
<p align="center">
  <img src="https://github.com/camalot/xget/raw/refs/heads/main/docs/assets/images/xget-logo-800x484.png" alt="xget logo" width="800">
</p>
<!-- markdownlint-enable MD033 -->

**xget** is the best way to easily get pre-built binaries for your favorite
tools. It downloads and extracts pre-built binaries from releases on GitHub or
GitLab. To use it, provide a repository and xget will search through the assets from the
latest release in an attempt to find a suitable prebuilt binary for your
system. If one is found, the asset will be downloaded and xget will extract the
binary to the current directory. xget should only be used for installing
simple, static prebuilt binaries, where the extracted binary is all that is
needed for installation. For more complex installation, you may use the
`--download-only` option, and perform extraction manually.

![xget Demo](https://github.com/camalot/xget/raw/refs/heads/main/docs/assets/images/xget-demo.gif)

For software maintainers, if you provide prebuilt binaries on GitHub or
GitLab, you can list `xget` as a one-line method for users to install your
software.

xget has a number of detection mechanisms and should work out-of-the-box with
most software that is distributed via single binaries on GitHub or GitLab releases. First
try using xget on your software, it may already just work. Otherwise, see the
[FAQ](usage/faq) for a clear set of rules to make your software compatible with xget.

## Documentation sections

- [Installation](installation)
- [Usage](usage)
- [Asset Filtering](usage/asset-filtering)
- [Configuration](configuration)
- [FAQ](usage/faq)
