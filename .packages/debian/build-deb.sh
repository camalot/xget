#!/usr/bin/env bash
#
# Build a Debian binary package (.deb) for xget from an already-extracted
# release archive, using .packages/debian/control.yml as the packaging
# definition.
#
# Usage:
#   build-deb.sh --version 2.2.0 --arch amd64 --srcdir <extracted-archive-dir> \
#                --outdir <dir> [--control .packages/debian/control.yml]
#
# Requires: yq (mikefarah v4), dpkg-deb, gzip, md5sum
set -euo pipefail

control_yml="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/control.yml"
version=""
arch=""
srcdir=""
outdir=""

while [ $# -gt 0 ]; do
	case "$1" in
	--version)
		version="$2"
		shift 2
		;;
	--arch)
		arch="$2"
		shift 2
		;;
	--srcdir)
		srcdir="$2"
		shift 2
		;;
	--outdir)
		outdir="$2"
		shift 2
		;;
	--control)
		control_yml="$2"
		shift 2
		;;
	*)
		echo "Unknown argument: $1" >&2
		exit 2
		;;
	esac
done

for required in version arch srcdir outdir; do
	if [ -z "${!required}" ]; then
		echo "Missing required argument: --${required}" >&2
		exit 2
	fi
done

for tool in yq dpkg-deb gzip md5sum; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "Required tool '$tool' is not installed" >&2
		exit 1
	fi
done

if [ ! -f "$control_yml" ]; then
	echo "Packaging definition not found: $control_yml" >&2
	exit 1
fi

# reads .control[<key>], joining sequences into a comma-separated field value
# (`// ""` also collapses missing keys and `false` into an empty, unemitted field)
field() {
	KEY="$1" yq -r '[(.control[strenv(KEY)] // "")] | flatten | map(tostring) | join(", ")' "$control_yml"
}

# __VERSION__/__ARCH__ placeholders are resolved for every field value
render() {
	printf '%s' "$1" | sed -e "s/__VERSION__/${version}/g" -e "s/__ARCH__/${arch}/g"
}

package="$(render "$(field package)")"
if [ -z "$package" ]; then
	echo "control.package is required in $control_yml" >&2
	exit 1
fi

staging="${outdir}/${package}_${version}_${arch}"
rm -rf "$staging"
mkdir -p "$staging/DEBIAN" "$outdir"

conffiles=""
asset_count="$(yq -r '.assets | length' "$control_yml")"
for ((i = 0; i < asset_count; i++)); do
	path="$(IDX="$i" yq -r '.assets[env(IDX)].path' "$control_yml")"
	type="$(IDX="$i" yq -r '.assets[env(IDX)].type // "doc"' "$control_yml")"
	mode="$(IDX="$i" yq -r '.assets[env(IDX)].mode // "0644"' "$control_yml")"
	source="$(IDX="$i" yq -r '.assets[env(IDX)].source // ""' "$control_yml")"
	[ -n "$source" ] || source="$(basename "$path")"

	if [ ! -f "${srcdir}/${source}" ]; then
		echo "Asset '${source}' (for ${path}) not found in ${srcdir}" >&2
		exit 1
	fi

	dest="${staging}${path}"
	install -D -m "$mode" "${srcdir}/${source}" "$dest"

	case "$type" in
	manpage)
		# Debian policy 12.1: manual pages must be installed compressed
		gzip -9n --force "$dest"
		chmod "$mode" "${dest}.gz"
		;;
	config)
		conffiles="${conffiles}${path}"$'\n'
		;;
	esac
done

if [ -n "$conffiles" ]; then
	printf '%s' "$conffiles" >"$staging/DEBIAN/conffiles"
	chmod 0644 "$staging/DEBIAN/conffiles"
fi

# Installed-Size is the disk usage in KiB of everything but the control files
installed_size="$(du -s -k --exclude=DEBIAN "$staging" | cut -f1)"

control_file="$staging/DEBIAN/control"
: >"$control_file"

emit() {
	local name="$1" value="$2"
	[ -n "$value" ] || return 0
	printf '%s: %s\n' "$name" "$value" >>"$control_file"
}

emit_yesno() {
	local name="$1" value="$2"
	case "$value" in
	true | yes) emit "$name" "yes" ;;
	esac
}

# field order follows debian-policy ch-controlfields; Description must be last
emit Package "$package"
emit Source "$(render "$(field source)")"
emit Version "$(render "$(field version)")"
emit Section "$(render "$(field section)")"
emit Priority "$(render "$(field priority)")"
emit Architecture "$(render "$(field architecture)")"
emit_yesno Essential "$(field essential)"
emit Multi-Arch "$(render "$(field multi-arch)")"
emit Pre-Depends "$(render "$(field pre-depends)")"
emit Depends "$(render "$(field depends)")"
emit Recommends "$(render "$(field recommends)")"
emit Suggests "$(render "$(field suggests)")"
emit Enhances "$(render "$(field enhances)")"
emit Breaks "$(render "$(field breaks)")"
emit Conflicts "$(render "$(field conflicts)")"
emit Provides "$(render "$(field provides)")"
emit Replaces "$(render "$(field replaces)")"
emit Built-Using "$(render "$(field built-using)")"
emit Installed-Size "$installed_size"
emit Origin "$(render "$(field origin)")"
emit Bugs "$(render "$(field bugs)")"
emit Homepage "$(render "$(field homepage)")"
emit Maintainer "$(render "$(field maintainer)")"

summary="$(render "$(field summary)")"
if [ -z "$summary" ]; then
	echo "control.summary is required in $control_yml" >&2
	exit 1
fi
printf 'Description: %s\n' "$summary" >>"$control_file"

# extended description: every line is indented by one space, blank lines become " ."
description="$(render "$(field description)")"
if [ -n "$description" ]; then
	while IFS= read -r line; do
		if [ -z "${line//[[:space:]]/}" ]; then
			printf ' .\n' >>"$control_file"
		else
			printf ' %s\n' "$line" >>"$control_file"
		fi
	done <<<"$description"
fi

# md5sums for the installed files, as dpkg/debsums expect
(
	cd "$staging"
	find . -path ./DEBIAN -prune -o -type f -printf '%P\n' | LC_ALL=C sort | xargs -r -d '\n' md5sum >DEBIAN/md5sums
)
chmod 0644 "$staging/DEBIAN/md5sums"

deb="${outdir}/${package}_${version}_${arch}.deb"
dpkg-deb --root-owner-group --build "$staging" "$deb" >/dev/null
rm -rf "$staging"

echo "$deb"
