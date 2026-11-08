#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
fixture="$tmp/malformed.go"
cp "$root/tests/fixtures/go/malformed.go.txt" "$fixture"
if bash "$root/scripts/check-go-format.sh" --check "$fixture" > /dev/null 2>&1; then
  echo 'formatter accepted malformed Go' >&2
  exit 1
fi
bash "$root/scripts/check-go-format.sh" --write "$fixture"
bash "$root/scripts/check-go-format.sh" --check "$fixture"
