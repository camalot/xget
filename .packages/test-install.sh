#!/usr/bin/env bash
set -euo pipefail

download="$(cd "$1" && pwd)"
mkdir -p "$2"
destination="$(cd "$2" && pwd)"

curl() {
        local output="" url=""
        while [ "$#" -gt 0 ]; do
                case "$1" in
                -o) output="$2"; shift 2 ;;
                -*) shift ;;
                *) url="$1"; shift ;;
                esac
        done
        test -n "$output"
        cp "$download/$(basename "$url")" "$output"
}

uname() {
        case "$1" in
        -s) echo Linux ;;
        -m) echo armv7l ;;
        -o) echo GNU/Linux ;;
        *) return 1 ;;
        esac
}

export download
export -f curl uname
bash install/xget.sh --version v0.0.0 --no-checksum --dir "$destination/bin" --man-dir "$destination/man"
tar -xzf "$download/xget_0.0.0_linux_armv7.tar.gz" -C "$destination" xget
cmp "$destination/xget" "$destination/bin/xget"
test -f "$destination/man/xget.1"