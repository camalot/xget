---
title: xget
section: 1
header: xget Manual
---

# NAME
  xget - easily install prebuilt binaries from GitHub

# SYNOPSIS
  xget `[--help] [OPTIONS] TARGET`

  xget `install TARGET [OPTIONS]`

  xget `COMMAND [ARGS] [OPTIONS]`

# DESCRIPTION
  xget is a tool for downloading and extracting prebuilt binaries from releases
  on GitHub. To use it, provide a repository and xget will search through the
  assets from the latest release in an attempt to find a suitable prebuilt
  binary for your system. If one is found, the asset will be downloaded and
  xget will extract the binary to the current directory. xget should only be
  used for installing simple, static prebuilt binaries, where the extracted
  binary is all that is needed for installation. For more complex installation,
  you may use the `--download-only` option, and perform extraction manually.

  The **`PROJECT`** argument passed to xget should either be a GitHub
  repository, formatted as **`user/repo`** or **`user/repo@TAG`**, in which
  case xget will search the release assets, a direct URL, in which case xget will directly download and
  extract from the given URL, or a local file, in which case xget will extract
  directly from the local file.

  If xget downloads an asset called `xxx` and there also exists an asset called
  `xxx.sha256` or `xxx.sha256sum`, xget will automatically verify that the
  SHA-256 checksum of the downloaded asset matches the one contained in that
  file, and abort installation if a mismatch occurs.

  When installing an executable, xget will place it in the current directory by
  default. If the environment variable **`XGET_BIN`** is non-empty, xget will
  place the executable in that directory. The `--to` flag may also be used to
  customize the install location.

  Directories can also be specified as files to extract, and all files within
  them will be extracted. For example:

      xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5

  GitHub limits API requests to 60 per hour for unauthenticated users. Use
  `xget rate` or `xget --rate` to inspect the active limit. Set a personal
  access token in **`GITHUB_TOKEN`**, **`EGET_GITHUB_TOKEN`**, or
  **`XGET_GITHUB_TOKEN`** to make more requests; **`XGET_GITHUB_TOKEN`** takes
  precedence. xget prints this setup guidance when no token is configured for a
  rate check and after a GitHub API rate-limit response.

  xget loads environment files from the current directory in descending
  priority: `.secrets`, alphabetically sorted `*.secrets`, `.env`, then
  alphabetically sorted `*.env`. Earlier file values win; variables already set
  in the parent shell have the highest priority.

  Successful installs are recorded in `~/.config/xget/.xget.installed.yml`. Each
  record includes the repository or URL, source type, install location,
  timestamps, selected asset, download URL, extracted files, effective install
  options, SHA-256, and the installed and latest known tags. Records are keyed as
  `<source>:<repo-or-url>`, such as `github:nektos/act`. Reinstalling a package
  refreshes its existing record. This store is what `xget list --installed` and
  `xget upgrade` read from.

  The behavior of xget is configurable in a number of ways via options.
  Documentation for these options is provided below.

# COMMANDS
  `xget install TARGET`

:    Download and install a pre-built binary from GitHub releases. This is equivalent to the backwards-compatible `xget TARGET` form and accepts the same options.

  `xget uninstall PACKAGE`, `xget remove PACKAGE`

:    Remove a tracked package's extracted files and its installed metadata record. When no installed package matches, remove the target basename from `$XGET_BIN`, the current directory, or the directory selected with `--from`. `xget TARGET --remove` is retained as a backwards-compatible equivalent.

  `xget rate`

:    Show GitHub API rate limiting information. This is equivalent to the backwards-compatible `xget --rate` form. When no token is configured, setup guidance is written to standard error while rate information remains on standard output. xget also provides this guidance after a GitHub API response with status 429.

  `xget upgrade [PACKAGE]`

:    List and apply available upgrades for installed packages. With no arguments, the newest release of every installed package is looked up, the installed metadata store is refreshed, and packages with a newer release are listed in yellow. Pass `--no-color` for plain output. An upgrade is available only when the newest release is newer than the installed one; tags are compared as semantic versions, including prerelease ordering, falling back to a plain difference check for tags that are not semver-shaped. Packages installed from a direct URL or local file are skipped, since there is no release list to query. Pass a package to upgrade it, or `-a`/`--all` to upgrade everything that is not pinned. `PACKAGE` accepts the full name (`owner/repo`), the store key (`github:owner/repo`), or the bare repository name (`repo`). A package installed with a `tag` is pinned; it is listed separately, is never upgraded by `--all`, and must be named explicitly. Upgrading a pinned package re-pins it to the newer tag. The upgrade re-runs the download using options resolved from the `global` config section, then the matching `"owner/repo"` section, then the options stored at install time; `tag` and `upgrade_only` are never applied because either would prevent the newer release from being downloaded, and both are left untouched in the installed metadata store.

  `xget list [TARGET]`

:    Show up to ten recent releases for `TARGET`, including their name, tag, and publication date. Add `--pre-release` to include prereleases. With no `TARGET`, list the repositories defined in the configuration file. With `--installed`, show installed package metadata instead, including the last install or upgrade date; rows with a newer available version are yellow unless `--no-color` is set.

  `xget config <SUBCOMMAND>`

:    Get, set, and edit configuration values, in the same spirit as `git config`. See the CONFIGURATION section below.

  `xget version`

:    Print the xget version, commit, and build date.

  `xget completion <SHELL>`

:    Generate a shell completion script for `bash`, `zsh`, `fish`, or `powershell`.

# OPTIONS
  `-t, --tag=`

:    Use the given tagged release instead of the latest release. `user/repo@TAG` is equivalent. Use `latest` to explicitly select the latest stable release; with `--pre-release`, it selects the newest release whether stable or prerelease. If the project does not have a tag that matches exactly, xget will look for a tag that contains the given string, and use the latest one. Example: **`xget -t nightly zyedidia/micro`**.

  `--pre-release`

:    Include pre-releases when fetching the latest version. This will get the latest overall release, even if it is a pre-release.

  `--source`

:    Download the source code for the repository (only works for GitHub repositories) rather than a release. Downloads from the "master" branch by default. Use `--tag` to download a different tag or branch.

  `--to=`

:    Move the executable to the given name after extraction. If the name is `-`, it the data will be written to stdout. Example: **`xget zyedidia/micro --to /usr/local/bin`**. Example: **`xget --asset nvim.appimage --to nvim neovim/neovim`**.

  `-s, --system=`

:    Use the given system as the target instead of the host. Systems follow the notation 'OS/Arch', where OS is a valid OS (darwin, windows, linux, netbsd, openbsd, freebsd, android, illumos, solaris, plan9), and Arch is a valid architecture (amd64, 386, arm, arm64, riscv64). If the special value **all** is used, all possibilities are given and the user must select manually. Example: **`xget -s darwin/amd64 zyedidia/micro`**.

  `-f, --file=`

:    Extract the file that matches the given glob. You may want use this option to extract non-binary files. Example: **`xget -f LICENSE zyedidia/micro`**.

  `--all`

:    Extract all candidate files.

  `-q, --quiet`

:    Only print essential output.

  `--download-only`

:    Stop after downloading the asset. This prevents xget from performing extraction, allowing you to perform manual installation after the asset is downloaded.

  `--download-all`

:   Download all projects defined in the configuration file.

  `--upgrade-only`

:    Only download the asset if the release is more recent than an existing asset with the same name in `$XGET_BIN`, or the current directory if `$XGET_BIN` is not defined.

  `-a, --asset=`

:    Filter assets by matcher. Literal values prefer exact basename match first, then contains matching. Regex prefixes: `~`, `=~`, and `re:`. Negative prefixes: `^` and `not:`. Use `~~` or `^^` to escape a leading `~` or `^`, or use `text:` to force literal mode. This option can be specified multiple times for additional filtering. Example: **`xget --asset nvim.appimage neovim/neovim`**. Example: **`xget --asset '~^tool_.*\.zip$' --asset 'not:re:.*\.sbom\.json$' owner/repo`**. If the assets are filterable using the `--system` detector (i.e., if applying the detector does not remove all candidates), the system detector is applied. Use `--system all` to always consider all assets.

  `--ignore=`

:    Exclude assets by matcher. Accepts literal contains matching and regex matching via `~`, `=~`, or `re:`. Negative prefixes `^` and `not:` are also supported and invert the ignore behavior. Use `~~` or `text:` when a literal begins with `~`. This option can be specified multiple times. Example: **`xget --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$' tacocontent/ironstate`**.

  Patterns beginning with `~` are parsed as regex, and patterns beginning with `^` are parsed as negative by default. Use `~~`, `^^`, or `text:` for literal prefixes.

  Always quote matchers starting with `~` (for example `--ignore '~\.sha512$'`). Unquoted, shells such as PowerShell, bash, and zsh expand a leading `~` to the home directory before xget receives the argument, silently disabling the filter.

  `--sha256`

:    Show the SHA-256 hash of the downloaded asset. This can be used to verify that the asset is not corrupted.

  `--verify, --verify-sha256=`

:    Verify the SHA-256 hash of the downloaded asset against the one provided as an argument. Similar to `--sha256`, but xget will do the verification for you.

  `--rate`

:    Show GitHub API rate limiting information.

  `--remove`

:    Uninstall the target package. This is equivalent to `xget uninstall TARGET`.

  `--from=DIR`

:    When no installed package matches, remove the target basename from `DIR` instead of `$XGET_BIN` or the current directory.

  `-k, --disable-ssl`

:    Disable SSL certificate verification for GET requests. Cannot be used in combination with a `GITHUB_TOKEN`.

  `--non-interactive`

:    Never prompt for input. Available for every command. If a run would require the user to choose an asset or file, xget writes `Interactive user request while execution is in non-interactive mode` to stderr and exits with code 16.

  `-c, --config=`

:    Use the given configuration file instead of searching the default locations. Example: **`xget -c ./project.xget.toml owner/repo`**.

  `-h, --help`

:    Show a help message.

  `-v, --version`

:    Print the xget version. Equivalent to **`xget version`**.

# EXIT STATUS
  `0`

:    Success.

  `1`

:    General error.

  `16`

:    An interactive action was required while running with `--non-interactive`.

# ENVIRONMENT
  `XGET_BIN`

:    Directory to place extracted executables in when `--to` is not given. `EGET_BIN` is also honored for compatibility.

  `XGET_GITHUB_TOKEN`, `EGET_GITHUB_TOKEN`, `GITHUB_TOKEN`

:    GitHub API token used to raise the request rate limit. `XGET_GITHUB_TOKEN` takes precedence. A value of `@/path/to/file` reads the token from that file.

  `XGET_CONFIG`, `EGET_CONFIG`

:    Path to the configuration file, overriding the default search locations.

  `XGET_EDITOR`, `VISUAL`, `EDITOR`

:    Editor used by `xget config edit`, checked in that order. Falls back to `nano`, and to `notepad` on Windows when `nano` is not on `PATH`. The value may include arguments, which are passed before the file path.

  `XDG_CONFIG_HOME`

:    Base directory for the OS configuration path. Defaults to `~/.config` when unset.

# CONFIGURATION
  xget checks for configuration in this order:

  1. `--config` if specified.
  2. `XGET_CONFIG` if set.
  3. `EGET_CONFIG` if set for compatibility with original eget.
  4. `./.xget.toml`, `./.xget.yml`, and `./.xget.yaml`.
  5. `~/.xget.toml`, `~/.xget.yml`, and `~/.xget.yaml`.
  6. `~/.config/xget/.xget.toml`, `~/.config/xget/.xget.yml`, and `~/.config/xget/.xget.yaml`.
  7. On Windows, `%LOCALAPPDATA%/xget/.xget.toml`, `%LOCALAPPDATA%/xget/.xget.yml`, and `%LOCALAPPDATA%/xget/.xget.yaml`.

  xget also supports the legacy eget-compatible filename `.eget.toml` in those same locations, and accepts `.eget.yml` / `.eget.yaml` if present.

  Both global settings can be configured, as well as setting on a per-repository basis.

  Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
  are the owner and repository name of the target repository (note that the `owner/repo` 
  format is quoted).

  A repository section inherits any setting it does not define from the `global`
  section. The only exceptions are `asset_filters`, `pre_release`, `tag`, and
  `verify_sha256`, which are repository-only and are never inherited.

  Values are resolved in this order, with each layer overriding the one before it:

  1. Built-in defaults.
  2. The `global` section.
  3. The matching `"owner/repo"` section.
  4. Command-line flags.

  For example, the following configuration file will set the `--to` flag to `~/bin` for 
  all repositories, and will set the `--to` flag to `~/.local/bin` for the `zyedidia/micro` 
  repository.

```toml
  [global]
  target = "~/bin"

  ["zyedidia/micro"]
  target = "~/.local/bin"
```

  More complete example configuration:

```toml
[global]
  quiet = false
  show_hash = false
  upgrade_only = true
  target = "./test"

["zyedidia/micro"]
  upgrade_only = false
  show_hash = true
  asset_filters = [ "static", ".tar.gz" ]
  target = "~/.local/bin/micro"
```

  By using the configuration above, you could run the following command to download
  the latest release of `micro`:
  **`xget zyedidia/micro`**

  Without the configuration, you would need to run the following command instead:
  **`xget zyedidia/micro --to ~/.local/bin/micro --sha256 --asset static --asset .tar.gz`**

## Available settings

  Unless noted otherwise, a setting may appear in the `global` section, in a
  `"owner/repo"` section, or both.

  `all`

:    Whether to extract all candidate files.

  `asset_filters`

:    An array of asset matchers. Literal values prefer exact basename match then contains matching. Regex prefixes: `~`, `=~`, and `re:`. Negative prefixes: `^` and `not:`. Use `~~`/`^^` or `text:` to force literal matching. Repository sections only.

  `disable_ssl`

:    Whether to disable SSL certificate verification for download requests.

  `download_only`

:    Whether to stop after downloading the asset (no extraction).

  `download_source`

:    Whether to download the repository source code instead of a release asset.

  `file`

:    The glob to select files for extraction.

  `github_token`
  
:    GitHub API token to use for requests. Global section only. Prefer `XGET_GITHUB_TOKEN` or `GITHUB_TOKEN`; xget warns when this is set in a config file, since it stores the token in plaintext where it may be committed to source control.

  `ignore`

:    An array of asset matchers to exclude. Supports the same matcher syntax as `asset_filters`.

  `pre_release`

:    Whether to include pre-releases when fetching the latest version. Repository sections only.

  `quiet`

:    Whether to only print essential output.

  `show_hash`

:    Whether to show the SHA-256 hash of the downloaded asset.

  `source`

:    The source type for the target, such as `GitHub` or `URL`.

  `system`

:    The target system to download for, in `OS/Arch` notation, or `all`.

  `tag`

:    Pin the repository to a specific tagged release. Repository sections only. A package installed with a `tag` is not upgraded by `xget upgrade --all`.

  `target`

:    The directory to move the downloaded file to after extraction.

  `upgrade_only`

:    Whether to only download if release is more recent than current version.

  `verify_sha256`

:    Verify the downloaded asset checksum against the given hash. Repository sections only.

## Template variables

  `asset_filters` and `ignore` entries sourced from the configuration file
  support the template variables `{{.OS}}` and `{{.Arch}}`, which are replaced
  with the effective target OS and architecture. Values passed on the command
  line with `--asset` or `--ignore` are used verbatim.

## The config command

  `xget config` reads and writes configuration values, in the same spirit as
  `git config`. Every subcommand accepts `-c`/`--config` to operate on a
  specific file.

  When `--config` is not given, the file is resolved using the search order
  above. If no configuration file exists anywhere, a new one is created at
  `$XDG_CONFIG_HOME/xget/.xget.yml`, or `~/.config/xget/.xget.yml` when
  `XDG_CONFIG_HOME` is empty or unset. This applies on all platforms.

  The format is determined by the file extension: `.toml` files are written as
  TOML, `.yml` and `.yaml` files as YAML. Any other extension is rejected.
  Writing rewrites the file, so comments and the original key ordering are not
  preserved.

  `xget config get <global|owner/repo> <key>`

:    Print the value of a key. List values print one entry per line. If the section or key is not set, nothing is printed and xget exits with status 1.

  `xget config set <global|owner/repo> <key>=<value>`

:    Set a value. Scalar keys are replaced. List keys (`ignore`, `asset_filters`) append the value; repeat the command to add more entries, and appending a value that is already present is a no-op. Booleans accept `true`, `false`, `t`, `f`, `1`, and `0`.

  `xget config clear <global|owner/repo> <key>`

:    Remove a key entirely, including every entry of a list key. Exits with status 1 and prints nothing when the key was not set. Empty repository sections are kept, because a section with no keys is still meaningful for `--download-all`.

  `xget config pop <global|owner/repo> <key>=<value>`

:    Remove a single entry from a list key; the key itself is removed when its last entry is removed. For scalar keys this behaves like `clear` when the current value matches, and otherwise exits with status 1.

  `xget config list`

:    Print every value in the resolved file as `section.key=value`, with `global` first and remaining sections sorted alphabetically. List keys produce one line per entry.

  `xget config path`

:    Print the configuration file path that would be read from or written to, whether or not it exists.

  `xget config edit`

:    Open the configuration file in an editor, creating it and its parent directories first if needed. After the editor exits, xget re-reads the file and reports a parse error if the edit made it invalid. See the ENVIRONMENT section for editor selection.

# FOR MAINTAINERS

To guarantee compatibility of your software's pre-built binaries with xget, you
can follow these rules.

* Provide your pre-built binaries as GitHub release assets.
* Format the system name as `OS_Arch` and include it in every pre-built binary
  name. Supported OSes are `darwin`/`macos`, `windows`, `linux`, `netbsd`, `openbsd`,
  `freebsd`, `android`, `illumos`, `solaris`, `plan9`. Supported architectures
  are `amd64`, `i386`, `arm`, `arm64`, `riscv64`.
* If desired, include `*.sha256` files for each asset, containing the SHA-256
  checksum of each asset. These checksums will be automatically verified by
  xget.
* Include only a single executable or appimage per system in each release archive.
* Use `.tar.gz`, `.tar.bz2`, `.tar.xz`, `.tar`, or `.zip` for archives. You may
  also directly upload the executable without an archive, or a compressed
  executable ending in `.gz`, `.bz2`, or `.xz`.

If you don't follow these rules, xget may still work well with your software.
xget's auto-detection is much more relaxed than what is required by these
rules, but if you follow these rules your software is guaranteed to be
compatible with xget.

# BUGS

See GitHub Issues: <https://github.com/camalot/xget/issues>

# AUTHOR

Ryan Conrad <camalot@gmail.com> xget fork
Zachary Yedidia <zyedidia@gmail.com> [eget](https://github.com/zyedidia/eget)
