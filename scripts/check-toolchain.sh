#!/usr/bin/env bash
set -Eeuo pipefail

missing=()
for tool in go make git; do
  command -v "$tool" >/dev/null 2>&1 || missing+=("$tool")
done

if ((${#missing[@]})); then
  printf 'Missing required tools: %s\n' "${missing[*]}" >&2
  exit 1
fi
