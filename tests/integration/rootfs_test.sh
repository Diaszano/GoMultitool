#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
bash "$root/scripts/build-armv7.sh" "$work/multitool"
LC_ALL=C readelf -h "$work/multitool" | grep -E 'Machine:[[:space:]]+ARM$' > /dev/null
bash "$root/build/rootfs/install-multitool.sh" "$work/rootfs" "$work/multitool"
[[ -x "$work/rootfs/usr/local/bin/multitool" ]]
[[ -x "$work/rootfs/usr/local/sbin/multitool-init" ]]
LC_ALL=C readelf -h "$work/rootfs/usr/local/bin/multitool" | grep -E 'Machine:[[:space:]]+ARM$' > /dev/null
grep -q 'multitool.sh' "$work/rootfs/usr/local/sbin/multitool-init"
grep -q 'exec /usr/local/bin/multitool tui' "$work/rootfs/usr/local/sbin/multitool-init"
grep -q '/usr/local/bin/multitool.sh' "$root/sources/multistrap/init"
grep -q 'packages=util-linux' "$root/sources/multistrap/multistrap.conf"
