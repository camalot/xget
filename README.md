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
xget zyedidia/micro --tag nightly
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
```

## How to get xget

Before you can get anything, you have to get xget. If you already have xget and want to upgrade, use `xget camalot/xget`.

### Quick-install script

``` shell
curl -o xget.sh https://camalot.github.io/xget.sh
shasum -a 256 xget.sh # verify with hash below
bash xget.sh
```

Or alternatively (less secure):

``` shell
curl https://camalot.github.io/xget.sh | sh
```

You can then place the downloaded binary in a location on your `$PATH` such as `/usr/local/bin`.

To verify the script, the sha256 checksum is `TODO: GENERATE CHECKSUM FOR SHELL SCRIPT` (use `shasum -a 256 xget.sh` after downloading the script).

One of the reasons to use xget is to avoid running curl into bash, but unfortunately you can't xget xget until you have xget.

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

GitHub limits API requests to 60 per hour for unauthenticated users. If you
would like to perform more requests (up to 5,000 per hour), you can set up a
personal access token and assign it to an environment variable named either
`GITHUB_TOKEN`, `EGET_GITHUB_TOKEN` or `XGET_GITHUB_TOKEN` when running xget.
If both are set, `XGET_GITHUB_TOKEN` will take precedence. xget will read this
variable and send the token as authorization with requests to GitHub. It is
also possible to read the token from a file by using `@/path/to/file` as the
token value.

``` text
Usage:
  xget [OPTIONS] TARGET

Application Options:
  -t, --tag=           tagged release to use instead of latest
      --pre-release    include pre-releases when fetching the latest version
      --source         download the source code for the target repo instead of a release
      --to=            move to given location after extracting
  -s, --system=        target system to download for (use "all" for all choices)
  -f, --file=          glob to select files for extraction
      --all            extract all candidate files
  -q, --quiet          only print essential output
  -d, --download-only  stop after downloading the asset (no extraction)
      --upgrade-only   only download if release is more recent than current version
    -a, --asset=         filter assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not:, escapes: ~~ and ^^, explicit literal: text:
      --ignore=        exclude assets by matcher; regex prefixes: ~, =~, re:, negative prefixes: ^ or not: (inverts ignore), escapes: ~~ and ^^, explicit literal: text:; can be specified multiple times
      --sha256         show the SHA-256 hash of the downloaded asset
      --verify-sha256= verify the downloaded asset checksum against the one provided
      --rate           show GitHub API rate limiting information
  -r, --remove         remove the given file from $XGET_BIN or the current directory
  -v, --version        show version information
  -h, --help           show this help message
  -D, --download-all   download all projects defined in the config file
  -k, --disable-ssl    disable SSL verification for download
  -c, --config=        path to config file (TOML or YAML)
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
the repository. It does not maintain a central list of packages, nor does it do
any dependency management. xget does not "install" executables by placing them
in system-wide directories (such as `/usr/local/bin`) unless instructed, and it
does not maintain a registry for uninstallation. xget works best for installing
software that comes as a single binary with no additional files needed (CLI
tools made in Go, Rust, or Haskell tend to fit this description).

### Does xget keep track of installed binaries?

xget does not maintain any sort of manifest containing information about
installed binaries. In general, xget does not maintain any state across
invocations. However, xget does support the `--upgrade-only` option, which
will first check `XGET_BIN` to determine if you have already downloaded the
tool you are trying to install -- if so it will only download a new version if
the GitHub release is newer than the binary on your file system.

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

Yes, you can pass a tag or tag identifier with the `--tag TAG` option. If no
tag exactly matches, xget will look for the latest release with a tag that
contains `TAG`. So if your repository contains releases for multiple different
just pass the appropriate tag (for the project you want) to xget, and
it will find the latest release for that particular project (as long as
releases for that project are given tags that contain the project name).

## Contributing

If you find a bug, have a suggestion, or something else, please open an issue
for discussion. I am sometimes prone to leaving pull requests unmerged, so
please double check with me before investing lots of time into implementing a
pull request. See [DOCS.md](DOCS.md) for more in-depth documentation.
