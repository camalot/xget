# Release Platforms

Linux package jobs use [platforms.json](platforms.json). Each entry defines the
Go target, release archive suffix, container platform, and package architecture
names. A missing or `null` package-format field means that format is unsupported.

| Release Suffix | Go Target | Docker Platform | Debian | RPM | Alpine | Arch Linux ARM |
| --- | --- | --- | --- | --- | --- | --- |
| `amd64` | `GOARCH=amd64` | `linux/amd64` | `amd64` | `x86_64` | `x86_64` | `x86_64` |
| `arm64` | `GOARCH=arm64` | `linux/arm64` | `arm64` | `aarch64` | `aarch64` | `aarch64` |
| `armv7` | `GOARCH=arm GOARM=7` | `linux/arm/v7` | `armhf` | `armv7hl` | `armv7` | `armv7h` |

ARMv7 uses Go's hard-float build. Debian package installation is tested on
Bookworm, which supports armhf; a newer Debian baseline may require ARMv7
capabilities beyond Go's minimum. RPMs for x86-64 and ARM64 retain the AlmaLinux 9
build/test environment and `.el9` suffix. ARMv7 RPMs use openSUSE Tumbleweed,
which supports this CPU, and consequently have a different distribution suffix.
The package architecture is not a promise that every RPM distribution supports
the CPU. Alpine packages use the configured abuild signing keys, or an ephemeral
key for local/dry-run testing. Arch Linux ARM packages are binary repacks built
with the matching `CARCH`; they are not official Arch Linux packages. Publishing
the PKGBUILD to AUR remains disabled until the existing account issue is resolved.

## Adding a Platform

1. Add the target to the appropriate build in [.goreleaser.yaml](../.goreleaser.yaml).
   ARM variants use `goarch: [arm]` and explicit `goarm` values. Archive names
   include `armvN`, so ARMv6 and ARMv7 cannot overwrite one another.
2. Add its entry to [platforms.json](platforms.json). Only declare package formats
   with supported CPU/ABI combinations. Set `rpm_image` to a maintained image
   that supports the platform and provides `dnf` or `zypper` and `rpm-build`.
3. Add installer detection and runtime asset matching when needed. Update the
   supported-platform documentation and regenerate installer checksums after
   changing an installer.
4. Run `go test ./scripts`, `goreleaser check`, and a release dry run. The contract
   tests reject missing Linux targets and duplicate package architecture names.
   The workflows derive their matrices from the manifest; no per-architecture
   build, download, SBOM, attestation, or upload steps need to be duplicated.

For future ARMv6, add Go's `goarm: ["6", "7"]` and a separate `armv6` entry.
Do not label an ARMv6 binary `armhf` indiscriminately: Debian armhf requires
ARMv7, while Alpine uses a different baseline for its `armhf` port. Leave Debian
and other incompatible formats unset. Match the image ABI and hardware baseline
before advertising a package. Windows and macOS do not gain 32-bit ARM targets.
The Nix flake in this directory is an unrelated `hello` prototype, not an xget
release publisher; no ARMv7 xget package is claimed for it.

## Package Artifacts

The main release and standalone publish workflows use per-format job matrices.
Each internal package action takes one required `architecture` (the release
suffix) and exposes `outdir`. Each verified package is uploaded with its own
`.sha256` and `.sbom.json` files, build provenance, and SBOM attestation. Separate
package checksums avoid parallel jobs overwriting a shared checksum file.
Dry runs upload uniquely named workflow artifacts per format and release suffix.

Run the local Docker-backed package tests with PowerShell 7, Bash (Git Bash on
Windows), Go, pandoc, and yq on `PATH`:

```powershell
& ./.packages/Test-Packages.ps1 -Architecture armv7
& ./.packages/Test-Packages.ps1
```

The default run derives all architectures from the manifest. Use `-Format` to
select `debian`, `rpm`, `alpine`, or `aur`. Docker must support the selected
container platforms through native execution or binfmt/QEMU. The runner creates
real CLI archives in a temporary directory, builds and inspects each package,
and checks installation, execution, and removal for Debian, RPM, and Alpine.
Pacman packages are cross-repacked and inspected; they are not install-tested
on Arch Linux ARM. These tests download container images and build dependencies.

The shared release downloader verifies exactly the selected archive names against
`checksums.txt`; checksum-like SBOM entries, duplicate hashes, missing assets,
and unsupported platform names cannot silently satisfy that verification.
