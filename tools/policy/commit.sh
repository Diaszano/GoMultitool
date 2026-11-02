#!/usr/bin/env bash
set -Eeuo pipefail

usage='usage: commit.sh --file MESSAGE_FILE | --stdin'
case "${1:-}" in
  --file)
    [[ $# == 2 ]] || {
      echo "$usage" >&2
      exit 2
    }
    message=$(cat "$2")
    ;;
  --stdin)
    [[ $# == 1 ]] || {
      echo "$usage" >&2
      exit 2
    }
    message=$(cat)
    ;;
  *)
    echo "$usage" >&2
    exit 2
    ;;
esac

subject=${message%%$'\n'*}
if [[ ! "$subject" =~ ^(feat|fix|refactor|perf|test|docs|build|ci|chore|style|revert)(\([a-z0-9]+(-[a-z0-9]+)*\))?!?:[[:space:]][^[:space:]].*$ ]]; then
  printf "Invalid commit subject: '%s'\n" "$subject" >&2
  exit 1
fi
