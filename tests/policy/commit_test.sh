#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
script="$root/tools/policy/commit.sh"
[[ -f "$script" ]] || {
  echo 'missing commit policy' >&2
  exit 1
}
for fixture in "$root"/tests/fixtures/commits/valid/*.txt; do
  bash "$script" --file "$fixture"
done
for fixture in "$root"/tests/fixtures/commits/invalid/*.txt; do
  if bash "$script" --file "$fixture" > /dev/null 2>&1; then
    echo "accepted invalid commit message: $fixture" >&2
    exit 1
  fi
done
printf '%s\n' 'docs: update the contribution guide' | bash "$script" --stdin

if printf '%s\n' 'Merge pull request #1 from feature/change' | bash "$script" --stdin > /dev/null 2>&1; then
  echo 'accepted a non-Conventional squash title' >&2
  exit 1
fi
