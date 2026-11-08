#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
[[ -f "$root/.semrelrc" ]] || {
  echo 'missing semantic release configuration' >&2
  exit 1
}

run_case() {
  local message=$1 expected=$2 repo output status=0
  repo=$(mktemp -d)
  cp "$root/.semrelrc" "$repo/.semrelrc"
  git -C "$repo" init -q -b main
  git -C "$repo" config user.name 'Policy Test'
  git -C "$repo" config user.email 'policy@example.invalid'
  printf '%s\n' baseline > "$repo/README.md"
  git -C "$repo" -c user.name='Policy Test' -c user.email='policy@example.invalid' -C "$repo" add README.md
  git -C "$repo" -c user.name='Policy Test' -c user.email='policy@example.invalid' -C "$repo" commit -qm 'chore: baseline'
  git -C "$repo" tag v0.1.0
  printf '%s\n' "$message" >> "$repo/README.md"
  git -C "$repo" -c user.name='Policy Test' -c user.email='policy@example.invalid' -C "$repo" add README.md
  git -C "$repo" -c user.name='Policy Test' -c user.email='policy@example.invalid' -C "$repo" commit -qm "$message"
  output=$(cd "$repo" && semantic-release --provider git --ci-condition default --no-ci --dry --allow-initial-development-versions 2>&1) || status=$?
  if [[ "$expected" == none ]]; then
    [[ "$output" == *'no change'* && "$output" != *'new version:'* ]] || {
      printf '%s\n' "$output" >&2
      return 1
    }
  else
    if [[ $status -ne 0 ]] || ! grep -Fq "$expected" <<< "$output"; then
      printf '%s\n' "$output" >&2
      return 1
    fi
    local retry_status=0 retry_output
    retry_output=$(cd "$repo" && semantic-release --provider git --ci-condition default --no-ci --dry --allow-initial-development-versions 2>&1) || retry_status=$?
    [[ $retry_status -eq 0 && "$retry_output" == *"new version: $expected"* ]] || {
      printf '%s\n' "$retry_output" >&2
      return 1
    }
  fi
  [[ $(git -C "$repo" tag --list | wc -l) -eq 1 ]]
  rm -rf "$repo"
}

run_case 'fix: repair boot menu' '0.1.1'
run_case 'feat: add diagnostics' '0.2.0'
run_case 'feat!: replace recovery command' '0.2.0'
run_case 'docs: clarify setup' none
