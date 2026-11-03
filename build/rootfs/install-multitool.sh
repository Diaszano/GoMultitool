#!/usr/bin/env bash
set -Eeuo pipefail

if [[ $# -ne 2 || ! -f $2 ]]; then
  echo 'usage: install-multitool.sh ROOTFS ARM_BINARY' >&2
  exit 2
fi
install -D -m 0755 "$2" "$1/usr/local/bin/multitool"
install -D -m 0755 "$(dirname "$0")/multitool-init" "$1/usr/local/sbin/multitool-init"
