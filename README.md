# xget: easy pre-built binary installation

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
tools. It downloads and extracts pre-built binaries from releases on GitHub or GitLab. To
use it, provide a repository and xget will search through the assets from the
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
most software that is distributed via single binaries on GitHub or GitLab
releases. First try using xget on your software, it may already just work.
Otherwise, see the FAQ for a clear set of rules to make your software
compatible with xget.

## Documentation

- [Documentation](https://camalot.github.io/xget)
- [Features](https://camalot.github.io/xget/features)
- [Examples](https://camalot.github.io/xget/usage/examples)
- [Installation](https://camalot.github.io/xget/installation)
- [GitHub Action](https://camalot.github.io/xget/usage/action)
- [Usage](https://camalot.github.io/xget/usage)
  - [Shell Completion](https://camalot.github.io/xget/usage/completion)
  - [Asset Filtering](https://camalot.github.io/xget/usage/asset-filtering)
  - [Updating Packages](https://camalot.github.io/xget/usage/upgrade)
- [FAQ](https://camalot.github.io/xget/faq)
- [Configuration](https://camalot.github.io/xget/configuration)
- [Code of Conduct](https://camalot.github.io/xget/code-of-conduct)
- [Contributing](https://camalot.github.io/xget/contributing)
- [Release Notes](https://camalot.github.io/xget/changelog)
- [Support](https://camalot.github.io/xget/support)
- [License](https://camalot.github.io/xget/license)

## Installation

> [!NOTE]
> See [full documentation](https://camalot.github.io/xget/installation) for latest installation instructions.

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget self-update`.

<!--- START INSTALL SECTION --->

<!-- markdownlint-disable MD033 -->
## Windows

<details markdown="block">
<summary markdown="span"><b>Windows install options</b></summary>

### Bash (Git Bash / WSL)

```shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

> [!NOTE]
> The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option

``` shell
curl -fsSL -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh && bash xget.sh -d /usr/local/bin
```

### PowerShell

> [!NOTE]
> When run from a downloaded file, the quick-install scripts will download and verify the sha256 checksum of the install script before continuing. When piped directly to `bash` or `iex`, script checksum verification is skipped because there is no local script file to hash. Release asset checksums are still verified.

```powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

### Scoop

```shell
scoop bucket add camalot https://github.com/camalot/scoop
scoop install camalot/xget
```

<!-- NOT YET AVAILABLE 
### WinGet

```shell
winget install camalot.xget
```

### Chocolatey

```shell
choco install xget
```
-->

### Archive

| Asset | Architecture |
| --- | --- |
| `xget_<VERSION>_windows_amd64.zip` | x64 |
| `xget_<VERSION>_windows_arm64.zip` | ARM64 |

</details>

## Linux

<details markdown="block">
<summary markdown="span"><b>Linux install options</b></summary>

### Bash

```shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

> [!NOTE]
> The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option

``` shell
curl -fsSL -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh && bash xget.sh -d /usr/local/bin
```

### PowerShell

```powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

### Debian / Ubuntu (`.deb`)

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/xget_<VERSION>_amd64.deb
sudo dpkg --install xget_<VERSION>_amd64.deb
```

Also available: `xget_<VERSION>_arm64.deb`.

### Fedora / RHEL / AlmaLinux (`.rpm`)

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/xget-<VERSION>-1.el9.x86_64.rpm
sudo dnf install ./xget-<VERSION>-1.el9.x86_64.rpm
```

Also available: `xget-<VERSION>-1.el9.aarch64.rpm`.

### Alpine (`.apk`)

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/xget-<VERSION>-r0-x86_64.apk
sudo apk add --allow-untrusted xget-<VERSION>-r0-x86_64.apk
```

Also available: `xget-<VERSION>-r0-aarch64.apk`.

### Arch Linux (`.pkg.tar.zst`)

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/xget-bin-<VERSION>-1-x86_64.pkg.tar.zst
sudo pacman -U xget-bin-<VERSION>-1-x86_64.pkg.tar.zst
```

Also available: `xget-bin-<VERSION>-1-aarch64.pkg.tar.zst`.

### Homebrew

```shell
brew install camalot/scoop/xget
```

### Archive

| Asset | Architecture |
| --- | --- |
| `xget_<VERSION>_linux_amd64.tar.gz` | x86_64 |
| `xget_<VERSION>_linux_arm64.tar.gz` | aarch64 |

</details>

## macOS

<details markdown="block">
<summary markdown="span"><b>macOS install options</b></summary>

### Bash

```shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

> [!NOTE]
> The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option

``` shell
curl -fsSL -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh && bash xget.sh -d /usr/local/bin
```

### PowerShell

```powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

### Homebrew

```shell
brew install camalot/scoop/xget
```

### Archive

| Asset | Architecture |
| --- | --- |
| `xget_<VERSION>_darwin_amd64.tar.gz` | Intel |
| `xget_<VERSION>_darwin_arm64.tar.gz` | Apple Silicon |

</details>

## Android (Termux)

<details markdown="block">
<summary markdown="span"><b>Android install options</b></summary>

### Bash

```shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

> [!NOTE]
> The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option

``` shell
curl -fsSL -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh && bash xget.sh -d /usr/local/bin
```

### Archive

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/xget_<VERSION>_android_arm64.tar.gz
tar -xzf xget_<VERSION>_android_arm64.tar.gz xget
install -Dm755 xget "$PREFIX/bin/xget"
```

| Asset | Architecture |
| --- | --- |
| `xget_<VERSION>_android_arm64.tar.gz` | ARM64 |

</details>

## Verify a download

<details markdown="block">
<summary markdown="span"><b>Checksum verification</b></summary>

```shell
curl -fsSLO https://github.com/camalot/xget/releases/download/v<VERSION>/checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

</details>

## zyedidia/eget

<details markdown="block">
<summary markdown="span"><b>Install with eget</b></summary>

If you already have eget, you can use it to install xget by running:

```shell
eget camalot/xget --asset "^.sbom.json"
```

</details>

Already have `xget`? Upgrade in place with `xget self-update`.
Full instructions: <https://camalot.github.io/xget/installation.html>

<!-- markdownlint-enable MD033 -->
<!--- END INSTALL SECTION --->

### GitHub Action

For GitHub Actions workflows, use [xget-action](https://github.com/camalot/xget-action)
to install `xget` (with binary caching) and run it in a single step:

``` yaml
- name: Install a tool with xget
  uses: camalot/xget-action@v1
  with:
    package: junegunn/fzf
```

You can also use the action to just install `xget` on the GitHub Actions runner without installing any packages. You can then use `xget` in subsequent steps to install other tools as needed.

``` yaml
- uses: camalot/xget-action@v1
- shell: bash
  run: |
    xget install eza-community/eza --to ~/.local/bin
    xget install junegunn/fzf --to ~/.local/bin
    xget install bschaatsbergen/cidr --to ~/.local/bin
```

The `xget-action` uses the `--non-interactive` flag by default to ensure that installations do not prompt for user input, which is suitable for automated CI/CD environments.

See the [xget-action](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.

## Contributing

If you find a bug, have a suggestion, or something else, please open an issue
for discussion. See [CONTRIBUTING](CONTRIBUTING.md) for more information.

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
