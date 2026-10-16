#!/usr/bin/env bash
# Fails if 'go mod tidy' would change any module's go.mod/go.sum.
#
# Deliberately compares before/after copies rather than 'git diff --exit-code':
# that only works on a clean checkout (CI), and would false-fail locally
# whenever there are legitimate uncommitted dependency changes.
set -euo pipefail
cd "$(dirname "$0")/.."

failed=0
dirs=(.)
for mod in "${dirs[@]}"; do
  tmp="$(mktemp -d)"
  for f in go.mod go.sum; do
    [ -f "$mod/$f" ] && cp "$mod/$f" "$tmp/$f"
  done

  (cd "$mod" && go mod tidy)

  dirty=0
  for f in go.mod go.sum; do
    if [ -f "$mod/$f" ] || [ -f "$tmp/$f" ]; then
      cmp -s "$mod/$f" "$tmp/$f" || dirty=1
    fi
  done
  rm -rf "$tmp"

  if [ "$dirty" -eq 1 ]; then
    echo "❌ $mod: 'go mod tidy' changed go.mod/go.sum - commit the tidied files"
    failed=1
  else
    echo "✅ $mod: module metadata is tidy"
  fi
done

exit "$failed"