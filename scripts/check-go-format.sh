#!/usr/bin/env bash
set -Eeuo pipefail

mode=${1:?usage: check-go-format.sh --check|--write [files...]}
shift
(($#)) || {
  echo 'provide at least one Go file' >&2
  exit 2
}

case "$mode" in
  --write) goimports -w "$@" ;;
  --check)
    unformatted=$(goimports -l "$@")
    if [[ -n "$unformatted" ]]; then
      printf 'Go files need formatting:\n%s\n' "$unformatted" >&2
      exit 1
    fi
    ;;
  *)
    echo 'usage: check-go-format.sh --check|--write [files...]' >&2
    exit 2
    ;;
esac
