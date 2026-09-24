#!/bin/sh
#
# Configure abuild and build the signed xget apk. Runs as root *inside* an
# Alpine chroot created by jirutka/setup-alpine, so this must stay POSIX sh.
#
# Required environment:
#   APKBUILD_DIR      directory holding the rendered APKBUILD
#   REPODEST          directory the built apk is written to
#   SRCDEST           directory holding the pre-downloaded release archives
#   PACKAGER          "Name <email>" recorded in the package
#   SIGNING_KEY_NAME  file name of the abuild RSA key pair
# Optional environment:
#   SIGNING_PRIVATE_KEY / SIGNING_PUBLIC_KEY
#                     PEM contents of the signing key pair; when either is
#                     empty an ephemeral key is generated instead
set -eu

: "${APKBUILD_DIR:?}"
: "${REPODEST:?}"
: "${SRCDEST:?}"
: "${PACKAGER:?}"
: "${SIGNING_KEY_NAME:?}"

abuild_home="$HOME/.abuild"
mkdir -p "$abuild_home" "$REPODEST" "$SRCDEST"

cat >"$abuild_home/abuild.conf" <<EOF
PACKAGER="$PACKAGER"
REPODEST="$REPODEST"
SRCDEST="$SRCDEST"
EOF

if [ -n "${SIGNING_PRIVATE_KEY:-}" ] && [ -n "${SIGNING_PUBLIC_KEY:-}" ]; then
	echo "Using the provided abuild signing key '$SIGNING_KEY_NAME'"
	(
		umask 077
		printf '%s\n' "$SIGNING_PRIVATE_KEY" >"$abuild_home/$SIGNING_KEY_NAME"
	)
	printf '%s\n' "$SIGNING_PUBLIC_KEY" >"$abuild_home/$SIGNING_KEY_NAME.pub"
	chmod 644 "$abuild_home/$SIGNING_KEY_NAME.pub"
	echo "PACKAGER_PRIVKEY=\"$abuild_home/$SIGNING_KEY_NAME\"" >>"$abuild_home/abuild.conf"
	# lets apk install and verify the package without --allow-untrusted
	install -m644 "$abuild_home/$SIGNING_KEY_NAME.pub" /etc/apk/keys/
else
	echo "No signing key supplied - generating an ephemeral one for this build"
	# not `-i`: that installs the public key via doas/sudo, which the chroot lacks
	abuild-keygen -a -n
	install -m644 "$abuild_home"/*.rsa.pub /etc/apk/keys/
fi

cd "$APKBUILD_DIR"
# -F: abuild refuses to run as root otherwise, and the chroot has no build user
abuild -F -r
