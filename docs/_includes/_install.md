
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
