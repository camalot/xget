#!/bin/sh
#
# Install the built xget apk, check every packaged file landed on disk, make
# sure the binary runs, then remove it again. Runs as root *inside* an Alpine
# chroot created by jirutka/setup-alpine, so this must stay POSIX sh.
#
# Required environment:
#   APK_FILE  path to the .apk to verify
#   VERSION   plain version the package should report
set -eu

: "${APK_FILE:?}"
: "${VERSION:?}"

echo "==> apk verify"
apk verify "$APK_FILE"

echo "==> apk add"
apk add "$APK_FILE"

echo "==> packaged files"
# `apk info -L` prints a "<pkg> contains:" header followed by paths without a leading slash
apk info -L xget | tail -n +2 | while read -r path; do
	[ -n "$path" ] || continue
	if [ ! -e "/$path" ]; then
		echo "Installed package is missing /$path" >&2
		exit 1
	fi
	echo "ok: /$path"
done

for path in /usr/bin/xget /usr/share/man/man1/xget.1.gz \
	/usr/share/licenses/xget/LICENSE /usr/share/doc/xget/README.md; do
	if [ ! -e "$path" ]; then
		echo "Expected $path to be installed" >&2
		exit 1
	fi
done

echo "==> xget --version"
test -x /usr/bin/xget
xget --version | tee /tmp/xget-version.txt
grep -q "xget $VERSION" /tmp/xget-version.txt

echo "==> apk del"
apk del xget
test ! -e /usr/bin/xget
echo "verified $(basename "$APK_FILE")"
