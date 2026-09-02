#!/usr/bin/env bash
# Downloads and installs the latest (or a specific) xget release binary
# for this machine's OS/architecture from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/camalot/xget/develop/install/xget.sh | bash
#   ./xget.sh [-d|--dir <path>] [-v|--version <tag>] [-h|--help]
set -euo pipefail
SKIP_CHECKSUM=0

REPO="camalot/xget"
BINARY="xget"
DEFAULT_INSTALL_DIR="$HOME/.local/bin"

INSTALL_DIR="$DEFAULT_INSTALL_DIR"
VERSION=""

usage() {
	cat <<EOF
Usage: install.sh [options]

Downloads and installs the latest (or a specific) $REPO release binary
for this machine's OS/architecture from GitHub Releases
(https://github.com/${REPO}/releases).

Options:
  -d, --dir <path>       Install directory (default: ${DEFAULT_INSTALL_DIR})
  -v, --version <tag>    Install a specific release tag, e.g. v0.1.0 (default: latest)
	-x, --no-checksum      Skip script checksum verification (not recommended)
  -h, --help             Show this help message
EOF
}

while [ $# -gt 0 ]; do
	case "$1" in
	-d | --dir)
		INSTALL_DIR="$2"
		shift 2
		;;
	-v | --version)
		VERSION="$2"
		shift 2
		;;
	-x | --no-checksum)
		SKIP_CHECKSUM=1
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "error: unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
done

need_cmd() {
	command -v "$1" >/dev/null 2>&1 || {
		echo "error: '$1' is required but not found on PATH" >&2
		exit 1
	}
}
need_cmd curl
need_cmd tar

TAG=""
ASSET=""

# Prints a block of diagnostics the user can paste directly into the
# GitHub issue linked by fail_unsupported.
issue_block() {
	cat <<EOF
### Install script diagnostics

- Repository: ${REPO}
- Script: install.sh
- Requested version: ${VERSION:-latest}
- uname -s: ${os_raw}
- uname -m: ${arch_raw}
- Detected OS: ${os:-<unrecognized>}
- Detected Arch: ${arch:-<unrecognized>}
- Resolved tag: ${TAG:-<none>}
- Attempted asset: ${ASSET:-<none>}
EOF
}

fail_unsupported() {
	echo "error: $1" >&2
	echo "" >&2
	echo "No supported $BINARY release was found for this machine." >&2
	echo "Please open an issue: https://github.com/${REPO}/issues/new?title=$(printf '%s' "Unsupported platform: ${os_raw}/${arch_raw}" | sed 's/ /+/g')" >&2
	echo "" >&2
	echo "Paste the block below into the issue:" >&2
	echo "" >&2
	issue_block >&2
	exit 1
}

get_script_checksum() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$0" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$0" | awk '{print $1}'
	else
		echo "warning: no sha256sum/shasum found; cannot compute script checksum" >&2
		return 1
	fi
}

script_checksum() {
	[ "$SKIP_CHECKSUM" -eq 1 ] && return 0

	# download the script checksum from GitHub and compare it to the local checksum
	# (this is a basic integrity check to ensure the script hasn't been tampered with)
	checksum_url="https://raw.githubusercontent.com/${REPO}/main/install/xget.sh.sha256"
	if curl -fsSL -o /tmp/xget.sh.sha256 "$checksum_url"; then
		expected_checksum="$(awk '{print $1}' /tmp/xget.sh.sha256)"
		actual_checksum="$(get_script_checksum)"
		if [ "$actual_checksum" != "$expected_checksum" ]; then
			echo "error: script checksum mismatch (expected ${expected_checksum}, got ${actual_checksum})" >&2
			exit 1
		fi
	else
		echo "warning: could not download script checksum from ${checksum_url}; use --no-checksum to skip this check" >&2
		exit 1
	fi
}

script_checksum

os_raw="$(uname -s)"
case "$os_raw" in
Linux) os="linux" ;;
Darwin) os="darwin" ;;
*) os="" ;;
esac

arch_raw="$(uname -m)"
case "$arch_raw" in
x86_64 | amd64) arch="amd64" ;;
aarch64 | arm64) arch="arm64" ;;
*) arch="" ;;
esac

if [ -z "$os" ] || [ -z "$arch" ]; then
	fail_unsupported "unsupported OS/architecture: ${os_raw}/${arch_raw}"
fi

if [ -n "$VERSION" ]; then
	TAG="$VERSION"
else
	echo "Looking up the latest release..."
	TAG="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name": *"[^"]*"' | head -n1 | sed -E 's/.*"([^"]+)"$/\1/')"
fi

if [ -z "$TAG" ]; then
	fail_unsupported "could not determine the latest release tag from the GitHub API"
fi

VERSION_NUM="${TAG#v}"
ASSET="${BINARY}_${VERSION_NUM}_${os}_${arch}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "Downloading ${ASSET} (${TAG})..."
if ! curl -fsSL -o "$tmpdir/$ASSET" "$URL"; then
	fail_unsupported "no release asset found at ${URL}"
fi

if curl -fsSL -o "$tmpdir/checksums.txt" "$CHECKSUMS_URL" 2>/dev/null; then
	expected="$(grep " ${ASSET}\$" "$tmpdir/checksums.txt" | awk '{print $1}')"
	if [ -n "$expected" ]; then
		if command -v sha256sum >/dev/null 2>&1; then
			actual="$(sha256sum "$tmpdir/$ASSET" | awk '{print $1}')"
		elif command -v shasum >/dev/null 2>&1; then
			actual="$(shasum -a 256 "$tmpdir/$ASSET" | awk '{print $1}')"
		else
			actual=""
			echo "warning: no sha256sum/shasum found; skipping checksum verification" >&2
		fi
		if [ -n "$actual" ] && [ "$actual" != "$expected" ]; then
			echo "error: checksum mismatch for ${ASSET} (expected ${expected}, got ${actual})" >&2
			exit 1
		fi
	fi
else
	echo "warning: could not download checksums.txt; skipping checksum verification" >&2
fi

echo "Extracting..."
tar -xzf "$tmpdir/$ASSET" -C "$tmpdir" "$BINARY"

mkdir -p "$INSTALL_DIR"
cp "$tmpdir/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"

echo "Installed ${BINARY} ${TAG} to ${INSTALL_DIR}/${BINARY}"
case ":$PATH:" in
*":${INSTALL_DIR}:"*) ;;
*)
	echo ""
	echo "note: ${INSTALL_DIR} is not on your PATH - add it, e.g.:"
	echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
	;;
esac
