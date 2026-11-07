#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
script="$root/sources/multistrap/multitool.sh"

# Gauge pipelines must run under pipefail and report the pipeline status.
awk '
  function fail(msg) { print msg " (line " NR ")" > "/dev/stderr"; bad = 1 }
  want {
    if ($0 ~ /^[[:space:]]*$/) next
    if ($0 !~ /^[[:space:]]*ERR=\$\{PIPESTATUS\[0\]\}$/) fail("gauge not followed by ERR=${PIPESTATUS[0]}")
    want = 0
  }
  inblock && $0 !~ /\\$/ { inblock = 0; want = 1 }
  /\) 2>&1 \| dialog/ {
    if ($0 !~ /^[[:space:]]*\(set -o pipefail; /) fail("gauge pipeline without set -o pipefail")
    if ($0 ~ /\\$/) inblock = 1; else want = 1
  }
  END { exit bad }
' "$script"

# choose_file must cope with spaces and apostrophes in file names.
dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
touch "$dir/a b.img" "$dir/c'd.img"
eval "$(sed -n '/^function choose_file()/,/^}/p' "$script")"
dialog() { echo 1 >&2; }
export TTY_CONSOLE=/dev/null
export BACKTITLE=t
got=$(choose_file t t "$dir/*")
[[ "$got" == "$dir/c'd.img" ]] || {
  echo "choose_file returned: $got" >&2
  exit 1
}
