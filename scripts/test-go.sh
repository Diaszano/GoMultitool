#!/usr/bin/env bash
set -Eeuo pipefail

packages=("$@")
((${#packages[@]})) || packages=(./...)
exec go test "${packages[@]}"
