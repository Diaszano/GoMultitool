#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
[[ -f "$root/scripts/install-zizmor.sh" ]] || {
  echo 'missing security tooling' >&2
  exit 1
}
PATH="$(go env GOPATH)/bin:$PATH" gitleaks --version > /dev/null
fixture_dir=$(mktemp -d "$root/tests/secret-test.XXXXXX")
trap 'rm -rf "$fixture_dir"' EXIT
printf 'token = "%s%s"\n' 'ghp_' '012345678901234567890123456789012345' > "$fixture_dir/secret.txt"
if PATH="$(go env GOPATH)/bin:$PATH" gitleaks dir --redact --no-banner "$fixture_dir" > /dev/null 2>&1; then
  echo 'secret scanner accepted a token fixture' >&2
  exit 1
fi
rm -rf "$fixture_dir"
trap - EXIT
[[ ! -e "$fixture_dir" ]]
