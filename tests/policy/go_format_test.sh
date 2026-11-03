#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
fixture="$tmp/malformed.go"
cp "$root/tests/fixtures/go/malformed.go.txt" "$fixture"
if PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-go-format.sh" --check "$fixture" > /dev/null 2>&1; then
  echo 'formatter accepted malformed Go' >&2
  exit 1
fi
PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-go-format.sh" --write "$fixture"
PATH="$(go env GOPATH)/bin:$PATH" bash "$root/scripts/check-go-format.sh" --check "$fixture"
