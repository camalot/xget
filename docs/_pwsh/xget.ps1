function Invoke-XGet {
<# 
.SYNOPSIS
    xget - easily install prebuilt binaries from GitHub or GitLab.

.DESCRIPTION
    xget is a tool for downloading and extracting prebuilt binaries from releases
    on GitHub or GitLab. To use it, provide a repository and xget will search through the
    assets from the latest release in an attempt to find a suitable prebuilt
    binary for your system. If one is found, the asset will be downloaded and
    xget will extract the binary to the current directory. xget should only be
    used for installing simple, static prebuilt binaries, where the extracted
    binary is all that is needed for installation. For more complex installation,
    you may use the `--download-only` option, and perform extraction manually.

    The **`PROJECT`** argument passed to xget should either be a GitHub or GitLab
    repository, formatted as **`user/repo`** or **`user/repo@TAG`** (GitLab may
    also use nested namespaces), in which
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

    Leading `~` is expanded in output paths, including **`XGET_BIN`**,
    **`EGET_BIN`**, and stored output paths used for upgrades.

    Directories can also be specified as files to extract, and all files within
    them will be extracted. For example:

    ```sh
    xget https://go.dev/dl/go1.17.5.linux-amd64.tar.gz --file go --to ~/go1.17.5
    ```

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
    `xget upgrade` read from. Use `--untracked` to skip recording a single install
    without disturbing an existing record for that target.

    A package may be installed to several locations. Each install location is
    tracked separately, so installing to a new directory adds a record instead of
    replacing the existing one, while reinstalling to an already tracked directory
    updates that record in place. `xget upgrade` and `xget uninstall` act on every
    tracked location unless one is selected with `--to` or `--from`.

    Legacy executable-path install locations are normalized to their containing
    directory. Duplicate records for the same package and directory are merged
    when loading the store, retaining the newest install record.

    The behavior of xget is configurable in a number of ways via options.
    Documentation for these options is provided below.


.EXTERNALHELP
    https://camalot.github.io/xget/

.LINK
    https://github.com/camalot/xget

.NOTES
    AUTHOR:
    - Ryan Conrad <camalot@gmail.com> xget
    - Zachary Yedidia <zyedidia@gmail.com> [eget](https://github.com/zyedidia/eget)
#>

}

New-Alias -Name xget -Value Invoke-XGet