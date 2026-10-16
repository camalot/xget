#!/usr/bin/env bash
# Verifies install/xget.* against their committed .sha256 files.
# Shared by 'task verify:sha256' and .github/actions/verify-sha256 so local
# runs and CI can never disagree about what "verified" means.
set -euo pipefail
cd "$(dirname "$0")/.."

failed=0
for f in install/xget.sh install/xget.ps1; do
  sum_file="$f.sha256"
  if [[ ! -f "$sum_file" ]]; then
    echo "❌ Missing checksum file $sum_file"
    failed=1
    continue
  fi
  expected="$(awk '{print $1}' "$sum_file")"
  actual="$(sha256sum "$f" | awk '{print $1}')"
  if [[ "$actual" != "$expected" ]]; then
    echo "❌ Checksum mismatch for $f"
    echo "   expected: $expected"
    echo "   actual  : $actual"
    failed=1
  else
    echo "✅ Checksum verified for $f"
  fi
done

if [[ $failed -eq 1 ]]; then
  echo "Run 'task sha256' to regenerate the checksum files." >&2
  exit 1
fi
