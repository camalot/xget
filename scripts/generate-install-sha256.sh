#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../install"
rm -f xget.*.sha256
for f in xget.*; do
  shasum -a 256 "$f" >"$f.sha256"
	cat "$f.sha256"
  sha256sum --check "$f.sha256"
done
