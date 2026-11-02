#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
fixture="$tmp/malformed.sh"
cp "$root/tests/fixtures/shell/malformed.sh.txt" "$fixture"
if PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-shell.sh" --check "$fixture" > /dev/null 2>&1; then
  echo 'shell checker accepted malformed script' >&2
  exit 1
fi
PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-shell.sh" --write "$fixture"
PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-shell.sh" --check "$fixture"
