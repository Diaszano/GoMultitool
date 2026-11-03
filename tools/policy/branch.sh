#!/usr/bin/env bash
set -Eeuo pipefail

branch=${1:?usage: branch.sh BRANCH_NAME}
default_branch=${DEFAULT_BRANCH:-main}

case "$branch" in
  "$default_branch" | main | next | next-major | beta | alpha) exit 0 ;;
esac
if [[ "$branch" =~ ^[0-9]+\.x$|^release/[0-9]+\.x$ ]]; then
  exit 0
fi
if [[ "$branch" =~ ^(feature|fix|hotfix|refactor|docs|chore|test|ci|release)/[a-z0-9]+(-[a-z0-9]+)*$ ]]; then
  exit 0
fi

printf "Invalid branch '%s'; use a lowercase conventional prefix and hyphen-separated description.\n" "$branch" >&2
exit 1
