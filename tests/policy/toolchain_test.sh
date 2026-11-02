#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
help=$(make -C "$root" help)
for target in help setup fmt fmt-check lint test test-race build build-armv7 integration-test image security check clean; do
  grep -Eq "^  $target([[:space:]]|$)" <<<"$help" || { echo "missing Make target: $target" >&2; exit 1; }
done

output=$(PATH=/nonexistent /bin/bash "$root/scripts/check-toolchain.sh" 2>&1) && { echo 'toolchain check unexpectedly passed' >&2; exit 1; }
grep -q 'Missing required tools: go make git' <<<"$output"
