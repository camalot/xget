# xget Documentation

## Code layout

xget is organized into internal packages so CLI concerns and runtime behavior
are separated:

- `cmd/xget/main.go`: application entrypoint.
- `internal/cli/root.go`: root Cobra command and CLI flag handling.
- `internal/config/config.go`: Viper-based configuration loading for TOML/YAML.
- `internal/options/options.go`: normalized runtime option structure.
- `internal/engine`: find/detect/verify/extract/download implementation.
- `internal/version/version.go`: version constant for runtime and release.

The runtime flow is unchanged from the user perspective.

xget works in four phases:

* Find: determine a list of assets that may be installed.
* Detect: determine which asset in the list should be downloaded for the target system.
* Verify: verify the checksum of the asset if possible.
* Extract: determine which file within the asset to extract.

If you are interested in reading the source code, the phase implementations are
in `internal/engine`, and orchestration is handled by `internal/engine/run.go`
via the Cobra root command in `internal/cli/root.go`.

## Configuration loading and precedence

xget now supports both TOML and YAML configuration formats while maintaining
backward compatibility with existing eget/xget TOML files.

Supported config filenames:

* `xget.toml`
* `xget.yaml`
* `xget.yml`

Search order:

1. `XGET_CONFIG` if defined.
2. Home file: `~/.xget.<ext>`.
3. Current directory: `./xget.<ext>`.
4. OS config path:
	* Linux/macOS: `$XDG_CONFIG_HOME/xget/xget.<ext>` or `~/.config/xget/xget.<ext>`
	* Windows: `%LOCALAPPDATA%/xget/xget.<ext>`

Resolution precedence:

1. CLI flags.
2. Repository section values (`"owner/repo"`).
3. Global section values (`global`).
4. Built-in defaults.

Config keys remain unchanged across TOML and YAML. For example, `target`,
`asset_filters`, `download_only`, and `verify_sha256` map to the same flags as
before.

## Find

If the input is a repo identifier, the Find phase queries `api.github.com` with
the repo and reads the list of assets from the response JSON. If a direct URL
is provided, the Find phase just returns the direct URL without doing any work.

## Detect

The Detect phase attempts to determine what OS and architecture each asset is
built for. This is done by matching a regular expression for each
OS/architecture that xget knows about. The match rules are shown below, and are
case insensitive.

| OS | Match Rule |
| ------------- | -------------------- |
| `darwin` | `darwin\|mac.?os\|osx` |
| `windows` | `win\|windows` |
| `linux` | `linux` |
| `netbsd` | `netbsd` |
| `openbsd` | `openbsd` |
| `freebsd` | `freebsd` |
| `android` | `android` |
| `illumos` | `illumos` |
| `solaris` | `solaris` |
| `plan9` | `plan9` |

| Architecture | Match Rule |
| ------------- | ----------------------------- |
| `amd64` | `x64\|amd64\|x86(-\|_)?64` |
| `386` | `x32\|amd32\|x86(-\|_)?32\|i?386` |
| `arm` | `arm` |
| `arm64` | `arm64\|armv8\|aarch64` |
| `riscv64` | `riscv64` |

If you would like a new OS/Architecture to be added, or find a case where the
auto-detection is not adequate (within reason), please open an issue.

Using the direct OS/Architecture (left column of the above tables) name in your
prebuilt zip file names will always allow xget to auto-detect correctly,
although xget will often auto-detect correctly for other names as well.

## Verify

During verification, xget will attempt to verify the checksum of the downloaded
asset. If the user has provided a checksum, or asked xget to simply print the
checksum, it will do so. Otherwise it may do auto-detection. If it is
downloading an asset called `xxx`, and there is another asset called
`xxx.sha256` or `xxx.sha256sum`, xget will automatically verify the SHA-256
checksum of the downloaded asset against the one contained in the
`.sha256`/`.sha256sum` file.

## Extract

During extraction, xget will detect the type of archive and compression, and
use this information to extract the requested file. If there is no requested
file, xget will extract a file with executable permissions, with priority given
to files that have the same name as the repo. If multiple files with executable
permissions exist and none of them match the repo name, xget will ask the user
to choose. Files ending in `.exe` are also assumed to be executable, regardless
of permissions within the archive.

xget supports the following filetypes for assets:

* `.tar.gz`/`.tgz`: tar archive with gzip compression.
* `.tar.bz2`: tar archive with bzip2 compression.
* `.tar.xz`: tar archive with xz compression.
* `.tar`: tar archive with no compression.
* `.zip`: zip archive.
* `.gz`: single file with gzip compression.
* `.bz2`: single file with bzip2 compression.
* `.xz`: single file with xz compression.
* otherwise: single file.

If a single file is "extracted" (no tar or zip archive), it will be marked
executable automatically.
