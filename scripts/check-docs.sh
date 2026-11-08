#!/usr/bin/env bash
set -Eeuo pipefail

mode=${1:?usage: check-docs.sh --check [files...]}
shift
[[ "$mode" == --check ]] || {
  echo 'usage: check-docs.sh --check [files...]' >&2
  exit 2
}
(($#)) || {
  echo 'provide at least one documentation file' >&2
  exit 2
}

root=$(cd "$(dirname "$0")/.." && pwd)
markdown=()
yaml=()
for file in "$@"; do
  case "$file" in
    *.md) markdown+=("$file") ;;
    *.yaml | *.yml) yaml+=("$file") ;;
    *)
      echo "unsupported documentation file: $file" >&2
      exit 2
      ;;
  esac
done

if ((${#markdown[@]})); then
  markdownlint-cli2 --config "$root/.markdownlint.yaml" "${markdown[@]}"
fi
if ((${#yaml[@]})); then
  yamllint -c "$root/.yamllint.yaml" "${yaml[@]}"
fi
