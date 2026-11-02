#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
bash "$root/scripts/build-armv7.sh" "$work/multitool"
readelf -h "$work/multitool" | grep -q 'Machine:.*ARM'
bash "$root/build/rootfs/install-multitool.sh" "$work/rootfs" "$work/multitool"
[[ -x "$work/rootfs/usr/local/bin/multitool" ]]
[[ -x "$work/rootfs/usr/local/sbin/multitool-init" ]]
readelf -h "$work/rootfs/usr/local/bin/multitool" | grep -q 'Machine:.*ARM'
grep -q 'multitool.sh' "$work/rootfs/usr/local/sbin/multitool-init"
grep -q '/usr/local/sbin/multitool-init' "$root/sources/multistrap/init"
grep -q 'packages=util-linux' "$root/sources/multistrap/multistrap.conf"
