#!/usr/bin/env bash
set -Eeuo pipefail

mode=${1:?usage: check-shell.sh --check|--write [files...]}
shift
(($#)) || {
  echo 'provide at least one shell file' >&2
  exit 2
}

case "$mode" in
  --write) shfmt -w -i 2 -ci -sr "$@" ;;
  --check) shfmt -d -i 2 -ci -sr "$@" ;;
  *)
    echo 'usage: check-shell.sh --check|--write [files...]' >&2
    exit 2
    ;;
esac
shellcheck -- "$@"
