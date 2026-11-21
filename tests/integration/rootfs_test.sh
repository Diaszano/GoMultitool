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
grep -q 'MULTITOOL_GO_RUNTIME=1 setsid --ctty --wait /usr/local/sbin/multitool-init' "$root/sources/multistrap/init"
grep -q '/usr/local/bin/multitool.sh' "$root/sources/multistrap/init"
grep -q 'packages=util-linux' "$root/sources/multistrap/multistrap.conf"

boot_environment=$(sed '/^# Mount the sys/,$d' "$root/sources/multistrap/init")
env -i HOME=/ TERM=linux /bin/bash -ec '
  eval "$1"
  /usr/bin/printenv PATH > /dev/null
  /usr/bin/env blkid --version > /dev/null
' bash "$boot_environment"

# Exercise font selection without accessing the host framebuffer or console.
# shellcheck disable=SC1090
source <(sed -n '/^configure_console_font()/,/^}/p' "$root/sources/multistrap/init")
fbset() { printf '    geometry %s\n' "$geometry"; }
setfont() { printf '%s\n' "$*"; }
expect_font() {
  local geometry=$1 expected=$2 actual
  actual=$(configure_console_font)
  if [[ "$actual" != "$expected" ]]; then
    printf 'geometry %s: expected %s, got %s\n' "$geometry" "$expected" "$actual" >&2
    return 1
  fi
}
expect_font '1280 720 1280 720 32' '-C /dev/tty1 /usr/share/consolefonts/Lat15-Terminus16.psf.gz'
expect_font '1920 1080 3840 2160 32' '-C /dev/tty1 /usr/share/consolefonts/Lat15-Terminus24x12.psf.gz'
expect_font '3840 2160 3840 2160 32' '-C /dev/tty1 /usr/share/consolefonts/Lat15-Terminus32x16.psf.gz'
expect_font '640 480 3840 2160 32' ''
expect_font 'unknown' ''
