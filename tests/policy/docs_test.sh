#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
[[ -f "$root/scripts/check-docs.sh" ]] || {
  echo "missing documentation checker" >&2
  exit 1
}
tmp=$(mktemp -d "$root/docs-test.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
cp "$root/tests/fixtures/docs/invalid.md.txt" "$tmp/invalid.md"
cp "$root/tests/fixtures/docs/invalid.yml.txt" "$tmp/invalid.yml"
for file in "$tmp/invalid.md" "$tmp/invalid.yml"; do
  if bash "$root/scripts/check-docs.sh" --check "$file" > /dev/null 2>&1; then
    echo "documentation checker accepted invalid file: $file" >&2
    exit 1
  fi
done
