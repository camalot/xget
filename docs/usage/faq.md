---
title: ❓ FAQ
nav_order: 3
layout: default
parent: 🧭 Usage
---

<!-- markdownlint-disable MD022 MD025 -->
# Frequently Asked Questions
{: .no_toc }

## Why use xget instead of [zyedidia/eget](https://github.com/zyedidia/eget)

- `xget` is a fork of `eget` with additional features and improvements. It aims to provide better compatibility with various release patterns, enhanced
asset filtering, and more robust handling of pre-built binaries. While `eget` is no longer actively maintained, `xget` continues to evolve to meet the
needs of users who rely on pre-built binaries from GitHub releases.
- `xget` uses a up to date version of Go, as well as modern libraries and practices to ensure better performance, security, and maintainability compared to the original `eget`.
- While introducing new features and improvements that enhance usability and performance,`xget` v2.0.0 is backwards compatible with the v1.x releases of `eget`. this means you can take your existing `eget` workflows and configurations and continue using them with `xget` without major changes. All you need to do is replace `eget` with `xget` in your commands, or create a symbolic link to `xget` named `eget`.

## Is xget compatible with all operating systems?

xget is designed to work on Unix-like systems such as Linux, macOS, and BSD variants, as well as Windows. However, the availability of pre-built binaries for your specific OS and architecture depends on the software you are trying to install. Always check the release assets to ensure compatibility.

## How is this different from a package manager?

xget only downloads pre-built binaries uploaded to GitHub by the developers of the repository. It does not maintain a central package list or a registry, and it does not manage dependencies. xget does not install software into system-wide directories unless you tell it to do so.

xget works best for installing standalone command-line tools that ship as a single binary. Tools built in Go, Rust, or Haskell often fit this model well.

## Does xget keep track of installed binaries?

xget records successful installs in `~/.config/xget/.xget.installed.yml`, including the files it extracted. Use `xget uninstall owner/repo` or `xget remove owner/repo` to remove a tracked package's files and record. The older `xget owner/repo --remove` form remains supported.

When no installed package matches, xget removes the target basename from `$XGET_BIN`, the current directory, or a directory passed with `--from`.

## Is this secure?

xget does not execute downloaded code. It only finds and extracts binaries from GitHub releases. If you trust the code you are downloading, then using xget is
reasonable. If xget finds a checksum file such as `xxx.sha256` or `xxx.sha256sum`, it will automatically verify the download checksum. You can also use
`--sha256` or `--verify / --verify-sha256` to validate checksums manually. If downloading a package directly from GitHub, if the asset in the release
has a sha256 checksum associated with the asset, xget will automatically verify it.

## Does this work only for GitHub repositories?

No. xget supports:

- GitHub repositories,
- direct URLs to archive files or executables, and
- local files that should be extracted in-place.

When you provide a direct URL or local file, xget skips repo detection and downloads or extracts the artifact directly.

## How can I make my software compatible with xget?

xget should work with many common release patterns, but these rules make compatibility more reliable:

- Provide pre-built binaries as GitHub release assets.
- Name the binary using the pattern `OS_Arch` and include it in each asset name. Supported values include `darwin`, `linux`, `windows`, `netbsd`, `openbsd`, `freebsd`, `android`, `illumos`, `solaris`, and `plan9`, along with architectures like `amd64`, `i386`, `arm`, `arm64`, and `riscv64`.
- Include `.sha256` or `.sha256sum` checksum files for each asset when possible.
- Keep each release archive to a single executable or app image per system.
- Use `.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may also upload a single executable directly or a compressed executable ending in `.gz`, `.bz2`, or `.xz`.

## Does this work with monorepos?

Yes. Use `owner/repo@TAG` or `--tag TAG` to select a release tag or tag fragment. If an exact tag match is not found, xget will look for the latest release whose tag contains the given value. This is useful when a repo contains multiple projects and each one has its own release tags.

Use `@latest` or `--tag latest` to explicitly select the latest stable release. Add `--pre-release` to select the newest release, whether it is stable or a prerelease.

For more examples of matching and filtering, see [Asset Filtering](asset-filtering).
