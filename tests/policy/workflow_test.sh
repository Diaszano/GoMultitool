#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
workflow="$root/.github/workflows/ci.yml"
[[ -f "$workflow" ]] || {
  echo 'missing CI workflow' >&2
  exit 1
}
grep -q '^  pull_request:' "$workflow"
grep -q '^  push:' "$workflow"
grep -q '^  policy:' "$workflow"
grep -q '^  quality:' "$workflow"
grep -Eq 'uses: actions/checkout@[0-9a-f]{40}' "$workflow"
grep -Eq 'uses: actions/setup-go@[0-9a-f]{40}' "$workflow"
grep -Eq 'uses: actions/upload-artifact@[0-9a-f]{40}' "$workflow"
if grep -Eq 'uses: [^ ]+@(v[0-9]+|main|master)$' "$workflow"; then
  echo 'CI action is not pinned to a commit SHA' >&2
  exit 1
fi
