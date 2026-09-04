# xget: easy pre-built binary installation

> [!NOTE]
> xget is a forked codebase of [zyedidia/eget](https://github.com/zyedidia/eget) focusing on some additional features and improvements. The original project does not seem to be actively maintained.
> The version of xget is starting at v2.0.0 to avoid confusion with the original project.

[![Release](https://img.shields.io/github/release/camalot/xget.svg?label=Release)](https://github.com/camalot/xget/releases)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/camalot/xget/blob/main/LICENSE)

**xget** is the best way to easily get pre-built binaries for your favorite
tools. It downloads and extracts pre-built binaries from releases on GitHub. To
use it, provide a repository and xget will search through the assets from the
latest release in an attempt to find a suitable prebuilt binary for your
system. If one is found, the asset will be downloaded and xget will extract the
binary to the current directory. xget should only be used for installing
simple, static prebuilt binaries, where the extracted binary is all that is
needed for installation. For more complex installation, you may use the
`--download-only` option, and perform extraction manually.

![xget Demo](https://github.com/camalot/xget/raw/refs/heads/main/docs/assets/images/xget-demo.gif)

For software maintainers, if you provide prebuilt binaries on GitHub, you can
list `xget` as a one-line method for users to install your software.

xget has a number of detection mechanisms and should work out-of-the-box with
most software that is distributed via single binaries on GitHub releases. First
try using xget on your software, it may already just work. Otherwise, see the
FAQ for a clear set of rules to make your software compatible with xget.

For more in-depth documentation, see [DOCS.md](DOCS.md).

## Project structure

xget now uses a structured Cobra/Viper architecture to make future command and
configuration extensions easier:

- `cmd/xget/main.go`: application entrypoint.
- `internal/cli/root.go`: root Cobra command and flag wiring.
- `internal/config`: configuration loading and normalization (TOML + YAML).
- `internal/options`: runtime option model used by the engine.
- `internal/engine`: find/detect/verify/extract/download runtime flow.
- `internal/version`: version string used by the CLI and release builds.

This preserves compatibility with existing eget/xget behavior while modernizing
the internal layout.

## Examples

``` shell
xget install eza-community/eza --tag v0.23.5 --to ~/.local/bin --asset .zip --ignore .sbom.json
xget zyedidia/micro --tag nightly
xget slavaGanzin/await@2.1.0
xget eza-community/eza --tag v0.23.5
xget eza-community/eza@latest --pre-release
xget jgm/pandoc --to /usr/local/bin
xget junegunn/fzf
xget neovim/neovim
xget ogham/exa --asset ^musl
xget tacocontent/ironstate --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$'
xget tacocontent/ironstate --asset 're:\.zip$' --ignore 'not:re:\.zip\.sbom\.json$'
xget --system darwin/amd64 sharkdp/fd
xget BurntSushi/ripgrep
xget -f xget.1 camalot/xget
xget zachjs/sv2v
xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
xget --all --file '*' ActivityWatch/activitywatch
xget list camalot/xget
xget list --installed
xget list camalot/xget --installed
```

## How to get xget

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget camalot/xget`.

### Quick-install script

### Bash

``` shell
curl -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh
shasum -a 256 xget.sh # verify with hash below
curl -o xget.sh.sha256 https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh.sha256
shasum -a 256 --check xget.sh.sha256
bash xget.sh
```

``` shell
curl -fsSL https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash
```

The default install location is `$HOME/.local/bin`. You can change the install location with the `-d` or `--dir` option:

``` shell
curl -o xget.sh https://raw.githubusercontent.com/camalot/xget/main/install/xget.sh | bash -s -- -d /usr/local/bin
```

### PowerShell

The script to install will automatically download the sha256 checksum and verify the script before executing it. If you want to manually verify the checksum, you can download the script and the checksum file separately and run the following command:

```powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 -OutFile xget.ps1
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1.sha256 -OutFile xget.ps1.sha256
$FileHash = Get-FileHash xget.ps1 -Algorithm SHA256
$ExpectedHash = Get-Content xget.ps1.sha256 | ForEach-Object { $_.Split(' ')[0] }
if ($FileHash.Hash -ne $ExpectedHash) {
    Write-Error "Checksum verification failed. Expected: $ExpectedHash, Actual: $($FileHash.Hash)"
} else {
  Write-Output "Checksum verification passed."
}
```

``` powershell
iwr https://raw.githubusercontent.com/camalot/xget/main/install/xget.ps1 | iex
```

<!-- ### Homebrew

``` shell
brew install xget
```
-->

<!-- ### Chocolatey

``` shell
choco install xget
``` -->

<!-- ### Scoop

``` shell
scoop bucket add main
scoop install xget
``` -->

<!-- ### Winget

``` shell
winget install camalot.xget
``` -->

### Pre-built binaries

Pre-built binaries are available on the [releases](https://github.com/camalot/xget/releases) page.

### From source

Install the latest released version:

``` shell
go install github.com/camalot/xget@latest
```

or install from HEAD:

``` shell
git clone https://github.com/camalot/xget
cd xget
go build ./...
```

A man page can be generated from the source tree with `pandoc`:

``` shell
pandoc docs/_man/xget.md -s -t man -o xget.1
```

You can also use `xget` to download the man page: `xget -f xget.1 camalot/xget`.

### GitHub Action

For GitHub Actions workflows, use [xget-action](https://github.com/camalot/xget-action)
to install `xget` (with binary caching) and run it in a single step:

``` yaml
- name: Install a tool with xget
  uses: camalot/xget-action@v1
  with:
    package: junegunn/fzf
```

See the [xget-action README](https://github.com/camalot/xget-action#readme)
for the full list of inputs/outputs and more examples.

## Usage

The `TARGET` argument passed to xget should either be a GitHub repository,
formatted as `user/repo`, in which case xget will search the release assets, a
direct URL, in which case xget will directly download and extract from the
given URL, or a local file, in which case xget will extract directly from the
local file.

If xget downloads an asset called `xxx` and there also exists an asset called
`xxx.sha256` or `xxx.sha256sum`, xget will automatically verify that the
SHA-256 checksum of the downloaded asset matches the one contained in that
file, and abort installation if a mismatch occurs.

When installing an executable, xget will place it in the current directory by
default. If the environment variable `XGET_BIN` is non-empty, xget will
place the executable in that directory.

Directories can also be specified as files to extract, and all files within
them will be extracted. For example:

```shell
xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
```

Successful installs are tracked in `~/.config/xget/.xget.installed.yml`. Each
record includes the repo or URL, source type, install location, timestamps,
selected asset, download URL, extracted files, effective install options,
SHA-256, and installed/current tag when release metadata is available. Records
are keyed as `<source>:<repo-or-url>`, such as `github:nektos/act`, and
reinstalling a package refreshes the existing record.

Use `xget list TARGET` to show up to ten recent releases with their name, tag,
and publication date. Add `--pre-release` to include prereleases. Use
`xget list --installed` to show installed package metadata, including the last
install or upgrade date, or `xget list TARGET --installed` to show one installed
package.

Rows with a newer available version are yellow in `xget list --installed` and
`xget upgrade`. Pass `--no-color` to either command for plain output.

### Upgrading installed packages

`xget upgrade` refreshes the installed metadata store and reports which packages
have a newer release, similar to `winget upgrade`. Full details are in the
[upgrade documentation](https://camalot.github.io/xget/usage/upgrade).

```bash
xget upgrade                      # list available upgrades
xget upgrade bschaatsbergen/cidr  # upgrade one package
xget upgrade --all                # upgrade everything that is not pinned
```

```text
Name                 Version  Available  Source
-----------------------------------------------
bschaatsbergen/cidr  v2.2.0   v2.3.0     GitHub

1 upgrade available.
```

Only packages with a newer release are listed. Tags are compared as semantic
versions, including prerelease ordering, falling back to a plain difference
check for tags that are not semver-shaped. Packages installed from a direct URL
or a local file are skipped, since there is no release list to query.

A package installed with a pinned `tag` is never upgraded by `--all`. Those are
listed separately and must be named explicitly:

```text
The following packages have an upgrade available, but require explicit targeting for upgrade:
Name                 Version  Available  Source
-----------------------------------------------
bschaatsbergen/cidr  v2.2.0   v2.3.0     GitHub
```

The package stays pinned afterward, with its stored `tag` updated to the newer
release.

The upgrade re-runs the download using options resolved from the `global` config
section, then the matching `"owner/repo"` section, then the options stored at
install time. The stored `tag` and `upgrade_only` are never applied, because
either would prevent the newer release from being downloaded; both are left
untouched in the installed metadata store. When no output location is
configured, the recorded `install_location` is used.

GitHub limits API requests to 60 per hour for unauthenticated users. Use
`xget rate` (or the backwards-compatible `xget --rate`) to inspect the
current limit. When no token is configured, xget prints setup guidance to
standard error. It also prints this guidance after a GitHub API response with
status 429.

To make more requests, use a personal access token and set it in either
`GITHUB_TOKEN`, `EGET_GITHUB_TOKEN`, or `XGET_GITHUB_TOKEN` before running
xget. `XGET_GITHUB_TOKEN` takes precedence. A token can also be read from a
file by using `@/path/to/file` as its value.

xget loads environment files from the current directory before running a
command. File values are considered in this order: `.secrets`, alphabetically
sorted `*.secrets`, `.env`, then alphabetically sorted `*.env`. A value from an
earlier file wins; environment variables already set by the shell always win.

``` text
Download pre-built binaries from GitHub releases

Usage:
  xget [TARGET] [flags]
  xget install TARGET [flags]
  xget [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Get, set, and edit xget configuration values
  help        Help about any command
  install     Download and install a pre-built binary from GitHub releases
  list        List available or installed packages
  rate        Show GitHub API rate limiting information
  uninstall   Remove an installed package
  upgrade     List and apply available upgrades for installed packages
  version     Print the xget version

Flags:
      --all                      extract all candidate files
  -a, --asset strings            filter assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not:, escapes: ~~ and ^^, explicit literal: text: (for example ^musl, not:~.*\.sbom\.json$, text:~literal)
  -c, --config string            path to the config file to use
  -k, --disable-ssl              disable SSL verification for download requests
  -D, --download-all             download all projects defined in the config file
  -d, --download-only            stop after downloading the asset (no extraction)
  -f, --file string              glob to select files for extraction
  -h, --help                     help for xget
      --ignore strings           exclude assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not: (inverts ignore), escapes: ~~ and ^^, explicit literal: text:; can be specified multiple times
      --pre-release              include pre-releases when fetching the latest version
  -q, --quiet                    only print essential output
      --rate                     show GitHub API rate limiting information
      --from string              directory to remove an untracked target from
    -r, --remove                   uninstall the target package
      --sha256                   show the SHA-256 hash of the downloaded asset
      --source                   download the source code for the target repo instead of a release
  -s, --system string            target system to download for (use all for all choices)
  -t, --tag string               tagged release to use instead of latest
      --to string                move to given location after extracting
      --upgrade-only             only download if release is more recent than current version
      --verify string[="auto"]   verify the downloaded asset checksum; pass a hash or use --verify with no value to use GitHub's published SHA256 when available
      --verify-sha256 string     verify the downloaded asset checksum against the one provided

Use "xget [command] --help" for more information about a command.
```

```text
Generate the autocompletion script for xget for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  xget completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion

Use "xget completion [command] --help" for more information about a command.
```

## Configuration

xget can be configured with either TOML or YAML files. Existing TOML-based
configuration remains fully supported.

Configuration search order:

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

Backward compatibility:

- xget also checks the original eget-compatible file name `.eget.toml` in the same locations.
- `.eget.yml` and `.eget.yaml` are accepted if present, even though original eget only shipped TOML.

OS config path defaults to:

- Linux/macOS: `$XDG_CONFIG_HOME/xget/.xget.<ext>` or `~/.config/xget/.xget.<ext>`.
- Windows: `%LOCALAPPDATA%/xget/.xget.<ext>` and `%USERPROFILE%/.xget.<ext>` are both checked.

Option precedence:

1. CLI flags.
2. Repository section (`"owner/repo"`) in config.
3. Global section (`global`) in config.
4. Built-in defaults.

A repository section inherits any setting it does not define from the `global`
section. The exceptions are `asset_filters`, `pre_release`, `tag`, and
`verify_sha256`, which are repository-only and never inherited; `github_token`
is global-only.

Both global settings can be configured, as well as setting on a per-repository basis.

Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
are the owner and repository name of the target repository (not that the `owner/repo`
format is quoted).

For example, the following configuration file will set the `--to` flag to `~/bin` for
all repositories, and will set the `--to` flag to `~/.local/bin` for the `zyedidia/micro`
repository.

```toml
[global]
target = "~/bin"

["zyedidia/micro"]
target = "~/.local/bin"
```

Equivalent YAML:

```yaml
global:
  target: "~/bin"

"zyedidia/micro":
  target: "~/.local/bin"
```

### Managing configuration with `xget config`

`xget config` reads and writes configuration values from the command line, similar
to `git config`. Full details are in the
[config command documentation](https://camalot.github.io/xget/configuration/config).

```bash
# scalar keys are replaced
xget config set global target=~/bin
xget config set global upgrade_only=true

# list keys (ignore, asset_filters) append; repeat to add more
xget config set zyedidia/micro asset_filters=static
xget config set zyedidia/micro asset_filters=.tar.gz

# read values (list values print one per line)
xget config get global target
xget config get zyedidia/micro asset_filters

# remove one list entry, or the whole key
xget config pop zyedidia/micro asset_filters=.tar.gz
xget config clear global target

# inspect
xget config list
xget config path

# open in an editor
xget config edit
```

Every subcommand accepts `--config <file>` (`-c`) to operate on a specific file.

Which file is written:

1. `--config <file>` if given.
2. `XGET_CONFIG` / `EGET_CONFIG` if set.
3. The first existing file in the search order above.
4. Otherwise a new `$XDG_CONFIG_HOME/xget/.xget.yml` is created, falling back to
   `~/.config/xget/.xget.yml` when `XDG_CONFIG_HOME` is empty or unset. This
   applies on Linux, macOS, and Windows.

The file is written in the format matching its extension: `.toml` files stay TOML,
`.yml`/`.yaml` files stay YAML. Writing rewrites the file, so comments and the
original key ordering are not preserved.

`xget config get`, `clear`, and `pop` exit with status `1` and print nothing when
the requested key is not set.

`xget config edit` opens the file with the first editor found in `XGET_EDITOR`,
`VISUAL`, or `EDITOR`, falling back to `nano` (and to `notepad` on Windows when
`nano` is not on `PATH`).

```bash
XGET_EDITOR=nano xget config edit
```

```powershell
$Env:XGET_EDITOR = "code --wait"
xget config edit
```

## Template Variables in Asset Filters / Ignore Lists

The `asset_filters` and `ignore` settings support template variables that are automatically replaced with the appropriate values based on your target system:

- `{{.OS}}` - Replaced with the target operating system (e.g., `linux`, `darwin`, `windows`)
- `{{.Arch}}` - Replaced with the target architecture (e.g., `amd64`, `arm64`, `386`)

The values are determined by:

1. The `--system` flag if provided (e.g., `--system linux/amd64`)
2. The `system` setting in the repository configuration
3. The runtime system (detected automatically from your current OS/architecture)

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

> Template substitution only applies to `asset_filters`/`ignore` values read from the config file. Passing `--asset` or `--ignore` on the command line uses those values verbatim (no `{{.OS}}`/`{{.Arch}}` substitution).

## Available settings - global section

> [!IMPORTANT]
> ⚠️ `github_token` is supported for backwards compatibility with the original `eget` project, but it is recommended to use `XGET_GITHUB_TOKEN` or `GITHUB_TOKEN` environment variable instead. Storing your GitHub token in a config file is not recommended, as it may be accidentally committed to source control and stored in plaintext. Use environment variables instead. `xget` will output a warning if `github_token` is set in the config file.

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| ⚠️ `github_token` | `N/A` | GitHub API token to use for requests | `""` |
| `all` | `--all` | Whether to extract all candidate files. | `false` |
| `download_only` | `--download-only` | Whether to stop after downloading the asset (no extraction). | `false` |
| `download_source` | `--source` | Whether to download the source code for the target repo instead of a release. | `false` |
| `file` | `--file` | The glob to select files for extraction. | `*` |
| `ignore` | `--ignore` | An array of asset matchers to exclude. Supports the same matcher syntax as `asset_filters` (`~`, `=~`, `re:`, `^`, `not:`, `~~`, `^^`, `text:`). Negative forms (`^`/`not:`) invert the ignore behavior. | `[]` |
| `pre_release` | `--pre-release` | Whether to include pre-releases when fetching the latest version. | `false` |
| `quiet` | `--quiet` | Whether to only print essential output. | `false` |
| `show_hash` | `--sha256` | Whether to show the SHA-256 hash of the downloaded asset. | `false` |
| `system` | `--system` | The target system to download for. | `all` |
| `target` | `--to` | The directory to move the downloaded file to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Whether to only download if release is more recent than current version. | `false` |
| `disable_ssl` | `--disable-ssl` | Disable SSL certificate verification for downloads. | `false` |

## Available settings - repository sections

| Setting | Related Flag | Description | Default |
| --- | --- | --- | --- |
| `all` | `--all` | Whether to extract all candidate files. | `false` |
| `asset_filters` | `--asset` | An array of asset matchers. Literal values use exact basename match first, then contains matching. Regex prefixes: `~`, `=~`, `re:`. Negative prefixes: `^`, `not:`. Use `~~`/`^^` or `text:` to force literal matching when needed. | `[]` |
| `download_only` | `--download-only` | Whether to stop after downloading the asset (no extraction). | `false` |
| `download_source` | `--source` | Whether to download the source code for the target repo instead of a release. | `false` |
| `file` | `--file` | The glob to select files for extraction. | `*` |
| `ignore` | `--ignore` | An array of asset matchers to exclude. Supports the same matcher syntax as `asset_filters`. Negative forms (`^`/`not:`) invert the ignore behavior. | `[]` |
| `pre_release` | `--pre-release` | Whether to include pre-releases when fetching the latest version. | global value |
| `quiet` | `--quiet` | Whether to only print essential output. | `false` |
| `show_hash` | `--sha256` | Whether to show the SHA-256 hash of the downloaded asset. | `false` |
| `system` | `--system` | The target system to download for. | `all` |
| `target` | `--to` | The directory to move the downloaded file to after extraction. | `.` |
| `upgrade_only` | `--upgrade-only` | Whether to only download if release is more recent than current version. | `false` |
| `verify_sha256` | `--verify-sha256` / `--verify` | Verify the sha256 hash of the asset against a provided hash. | `""` |
| `disable_ssl` | `--disable-ssl` | Disable SSL certificate verification for downloads. | `false` |

## Example configuration

```toml
[global]
  quiet = false
  show_hash = false
  upgrade_only = true
  ignore = ["~\\.sbom\\.json$"]
  target = "./test"

["zyedidia/micro"]
  upgrade_only = false
  show_hash = true
  asset_filters = [ "static", ".tar.gz" ]
  ignore = ["not:arm64"]
  target = "~/.local/bin/micro"
```

```yaml
global:
  quiet: false
  show_hash: false
  upgrade_only: true
  ignore:
    - "~\\.sbom\\.json$"
  target: "./test"

"zyedidia/micro":
  upgrade_only: false
  show_hash: true
  asset_filters:
    - "static"
    - ".tar.gz"
  ignore:
    - "not:arm64"
  target: "~/.local/bin/micro"
```

By using the configuration above, you could run the following command to download the latest release of `micro`:

```bash
xget zyedidia/micro
```

Without the configuration, you would need to run the following command instead:

```bash
export XGET_GITHUB_TOKEN=ghp_1234567890 &&\
xget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz
```

## Asset filtering

`--asset` supports short and long forms and can be repeated:

- Literal matcher: `--asset .zip`
- Literal anti-match: `--asset '^musl'` or `--asset 'not:musl'`
- Regex matcher: `--asset '~^tool_.*\.zip$'`, `--asset '=~^tool_.*\.zip$'`, or `--asset 're:^tool_.*\.zip$'`
- Regex anti-match: `--asset '^~.*\.sbom\.json$'`, `--asset 'not:~.*\.sbom\.json$'`, or `--asset 'not:re:.*\.sbom\.json$'`

`--ignore` always excludes matches and can also be repeated. It accepts literal and
regex matcher styles:

- `--ignore '.sbom.json'`
- `--ignore '~\.zip\.sbom\.json$'`
- `--ignore '=~\.zip\.sbom\.json$'`
- `--ignore 're:\.zip\.sbom\.json$'`
- `--ignore 'not:arm64'` (inverted ignore, keeps only matches for `arm64`)

Note: patterns that start with `~` are parsed as regex and patterns that start
with `^` are parsed as negative by default.

If a literal pattern starts with `~` or `^`, use escaping or explicit text mode:

- `--asset '~~literal-starts-with-tilde'`
- `--asset '^^literal-starts-with-caret'`
- `--asset 'text:~literal-starts-with-tilde'`

This allows combinations such as "use `.zip` but ignore `.zip.sbom.json`":

```bash
xget tacocontent/ironstate --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$'
```

## FAQ

### How is this different from a package manager?

xget only downloads pre-built binaries uploaded to GitHub by the developers of
the repository. It does not maintain a central list of packages or manage
dependencies. xget does not "install" executables into system-wide directories
(such as `/usr/local/bin`) unless instructed. It works best for software that
comes as a single binary with no additional files needed.

### Does xget keep track of installed binaries?

xget records successful installs in `~/.config/xget/.xget.installed.yml`,
including their extracted files. Use `xget uninstall owner/repo` (or
`xget remove owner/repo`) to remove those files and the matching record. The
legacy `xget owner/repo --remove` form remains supported.

When no installed record matches, xget tries to remove the target basename from
`$XGET_BIN`, the current directory, or a directory passed with `--from`.

### Is this secure?

xget does not run any downloaded code -- it just finds executables from GitHub
releases and downloads/extracts them. If you trust the code you are downloading
(i.e. if you trust downloading pre-built binaries from GitHub) then using xget
is perfectly safe. If xget finds a matching asset ending in `.sha256` or
`.sha256sum`, the SHA-256 checksum of your download will be automatically
verified. You can also use the `--sha256` or `--verify-sha256` options to
manually verify the SHA-256 checksums of your downloads (checksums are provided
in an alternative manner by your download source).

### Does this work only for GitHub repositories?

At the moment xget supports searching GitHub releases, direct URLs, and local
files. If you provide a direct URL instead of a GitHub repository, xget will
skip the detection phase and download directly from the given URL. If you
provide a local file, xget will skip detection and download and just perform
extraction from the local file.

### How can I make my software compatible with xget?

xget should work out-of-the-box with many methods for releasing software, and
does not require that you build your release process for xget in particular.
However, here are some rules that will guarantee compatibility with xget.

- Provide your pre-built binaries as GitHub release assets.
- Format the system name as `OS_Arch` and include it in every pre-built binary
  name. Supported OSes are `darwin`/`macos`, `windows`, `linux`, `netbsd`,
  `openbsd`, `freebsd`, `android`, `illumos`, `solaris`, `plan9`. Supported
  architectures are `amd64`, `i386`, `arm`, `arm64`, `riscv64`.
- If desired, include `*.sha256` files for each asset, containing the SHA-256
  checksum of each asset. These checksums will be automatically verified by
  xget.
- Include only a single executable or appimage per system in each release archive.
- Use `.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may
  also directly upload the executable without an archive, or a compressed
  executable ending in `.gz`, `.bz2`, or `.xz`.

### Does this work with monorepos?

Yes. Select a tag or tag fragment with either `xget owner/repo@TAG` or
`xget owner/repo --tag TAG`. If no tag exactly matches, xget looks for the
latest release whose tag contains `TAG`. This is useful when a repository has
releases for multiple projects.

Use `@latest` or `--tag latest` to explicitly request the latest stable
release. Adding `--pre-release` instead selects the newest release, whether it
is a stable release or a prerelease.

## Contributing

If you find a bug, have a suggestion, or something else, please open an issue
for discussion. I am sometimes prone to leaving pull requests unmerged, so
please double check with me before investing lots of time into implementing a
pull request. See [DOCS.md](DOCS.md) for more in-depth documentation.
