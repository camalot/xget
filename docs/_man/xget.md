---
title: xget
section: 1
header: xget Manual
---

# NAME
  xget - easily install prebuilt binaries from GitHub

# SYNOPSIS
  xget `[version] [--help] [OPTIONS] TARGET`

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
  repository, formatted as **`user/repo`**, in which case xget will search the
  release assets, a direct URL, in which case xget will directly download and
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

  GitHub limits API requests to 60 per hour for unauthenticated users. If you
  would like to perform more requests (up to 5,000 per hour), you can set up a
  personal access token and assign it to an environment variable named either
  **`GITHUB_TOKEN`** or **`XGET_GITHUB_TOKEN`** when running xget. If both are set,
  **`XGET_GITHUB_TOKEN`** will take precedence. xget will read this variable and
  send the token as authorization with requests to GitHub. It is also possible to
  read the token from a file by using `@/path/to/file` as the token value.

  The behavior of xget is configurable in a number of ways via options.
  Documentation for these options is provided below.

# OPTIONS
  `-t, --tag=`

:    Use the given tagged release instead of the latest release. If the project does not have a tag that matches exactly, xget will look for a tag that contains the given string, and use the latest one. Example: **`xget -t nightly zyedidia/micro`**.

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

   --upgrade-only

:    Only download the asset if the release is more recent than an existing asset with the same name in `$XGET_BIN`, or the current directory if `$XGET_BIN` is not defined.

  `-a, --asset=`

:    Filter assets by matcher. Literal values prefer exact basename match first, then contains matching. Regex prefixes: `~`, `=~`, and `re:`. Negative prefixes: `^` and `not:`. Use `~~` or `^^` to escape a leading `~` or `^`, or use `text:` to force literal mode. This option can be specified multiple times for additional filtering. Example: **`xget --asset nvim.appimage neovim/neovim`**. Example: **`xget --asset '~^tool_.*\.zip$' --asset 'not:re:.*\.sbom\.json$' owner/repo`**. If the assets are filterable using the `--system` detector (i.e., if applying the detector does not remove all candidates), the system detector is applied. Use `--system all` to always consider all assets.

  `--ignore=`

:    Exclude assets by matcher. Accepts literal contains matching and regex matching via `~`, `=~`, or `re:`. Negative prefixes `^` and `not:` are also supported and invert the ignore behavior. Use `~~` or `text:` when a literal begins with `~`. This option can be specified multiple times. Example: **`xget --asset '~\.zip$' --ignore '~\.zip\.sbom\.json$' tacocontent/ironstate`**.

  Patterns beginning with `~` are parsed as regex, and patterns beginning with `^` are parsed as negative by default. Use `~~`, `^^`, or `text:` for literal prefixes.

  `--sha256`

:    Show the SHA-256 hash of the downloaded asset. This can be used to verify that the asset is not corrupted.

  `--verify, --verify-sha256=`

:    Verify the SHA-256 hash of the downloaded asset against the one provided as an argument. Similar to `--sha256`, but xget will do the verification for you.

  `--rate`

:    Show GitHub API rate limiting information.

  `--remove`

:    Remove the target file from `$XGET_BIN` (or the current directory if unset). Note that this flag is boolean, and means xget will treat `TARGET` as a file to be removed.

  `-k, --disable-ssl`

:    Disable SSL certificate verification for GET requests. Cannot be used in combination with a `GITHUB_TOKEN`.

  `-h, --help`

:    Show a help message.

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

  The previous documentation used `./xget.<ext>` and `xget/xget.<ext>` instead of the dot-prefixed `.xget.<ext>` paths, which were incorrect.

  Both global settings can be configured, as well as setting on a per-repository basis.

  Sections can be named either `global` or `"owner/repo"`, where `owner` and `repo`
  are the owner and repository name of the target repository (note that the `owner/repo` 
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

  `all`

:    Whether to extract all candidate files.

  `asset_filters`

:    An array of asset matchers. Literal values prefer exact basename match then contains matching. Regex prefixes: `~`, `=~`, and `re:`. Negative prefixes: `^` and `not:`. Use `~~`/`^^` or `text:` to force literal matching.

  `ignore`

:    An array of asset matchers to exclude. Supports the same matcher syntax as `asset_filters`.

  `download_only`

:    Whether to stop after downloading the asset (no extraction).

  `file`

:    The glob to select files for extraction.

  `github_token`
  
:    GitHub API token to use for requests.

  `quiet`

:    Whether to only print essential output.

  `show_hash`

:    Whether to show the SHA-256 hash of the downloaded asset.

  `system`

:    The target system to download for.

  `target`

:    The directory to move the downloaded file to after extraction.

  `upgrade_only`

:    Whether to only download if release is more recent than current version.

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
