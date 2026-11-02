#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
script="$root/tools/policy/branch.sh"
[[ -f "$script" ]] || {
  echo 'missing branch policy' >&2
  exit 1
}
for branch in feature/go-tui fix/emmc-detection hotfix/cve-123 refactor/storage-adapters docs/recovery-guide chore/tooling test/policy ci/pipeline release/1.x main next next-major beta alpha 1.x; do
  bash "$script" "$branch"
done
for branch in Feature/go-tui feature/go_tui feature/ feature/has--double ci/Empty; do
  if bash "$script" "$branch" > /dev/null 2>&1; then
    echo "accepted invalid branch: $branch" >&2
    exit 1
  fi
done
DEFAULT_BRANCH=trunk bash "$script" trunk
