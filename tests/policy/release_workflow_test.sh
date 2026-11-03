#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
workflow="$root/.github/workflows/release.yml"
ci="$root/.github/workflows/ci.yml"
[[ -f "$workflow" && -f "$ci" ]] || {
  echo 'missing release or reusable CI workflow' >&2
  exit 1
}
grep -q '^  push:' "$workflow"
grep -q 'branches: \[main\]' "$workflow"
grep -q 'uses: ./\.github/workflows/ci.yml' "$workflow"
grep -q 'needs: \[ci\]' "$workflow"
grep -q 'semantic-release --provider git' "$workflow"
grep -q "git tag --points-at HEAD" "$workflow"
grep -q 'goreleaser release --clean' "$workflow"
grep -q 'contents: write' "$workflow"
grep -q 'cancel-in-progress: false' "$workflow"
grep -q '^  workflow_call:' "$ci"
[[ $(grep -c 'semantic-release --provider git' "$workflow") -eq 1 ]]
[[ $(grep -c 'goreleaser release --clean' "$workflow") -eq 1 ]]
grep -q 'replace_existing_artifacts: true' "$root/.goreleaser.yaml"
