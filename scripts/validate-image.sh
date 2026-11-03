#!/usr/bin/env bash
set -Eeuo pipefail

image=${1:?usage: validate-image.sh IMAGE}
[[ -f "$image" && ! -L "$image" ]] || {
  echo 'image must be a regular file' >&2
  exit 1
}
[[ $(uname -s) == Linux && $(id -u) -eq 0 ]] || {
  echo 'image validation requires root on Linux for isolated loop mounts' >&2
  exit 1
}
for tool in blkid losetup mount umount mountpoint findmnt sfdisk readelf python3; do
  command -v "$tool" > /dev/null || {
    echo "missing required image validation tool: $tool" >&2
    exit 1
  }
done

loop_device=
mount_dir=$(mktemp -d)
cleanup() {
  local status=$?
  trap - EXIT
  for mountpoint in "$mount_dir"/ntfs "$mount_dir"/boot "$mount_dir"/rootfs; do
    if mountpoint -q "$mountpoint"; then
      umount "$mountpoint" || status=1
    fi
  done
  if [[ -n "$loop_device" ]]; then
    losetup -d "$loop_device" || status=1
  fi
  rmdir "$mount_dir" 2> /dev/null || true
  exit "$status"
}
trap cleanup EXIT
mkdir "$mount_dir/ntfs" "$mount_dir/boot" "$mount_dir/rootfs"

python3 - "$image" << 'PY'
import json
import subprocess
import sys

layout = json.loads(subprocess.check_output(["sfdisk", "--json", sys.argv[1]]))["partitiontable"]
parts = layout["partitions"]
if len(parts) != 3 or not parts[1].get("bootable"):
    raise SystemExit("expected three partitions with partition 2 bootable")
if [part["node"] for part in parts] != [sys.argv[1] + str(i) for i in (1, 2, 3)]:
    raise SystemExit("unexpected partition order")
PY

loop_device=$(losetup --find --show --partscan "$image")
for suffix in 1 2 3; do
  for _ in {1..30}; do
    [[ -b "${loop_device}p${suffix}" ]] && break
    sleep 1
  done
  [[ -b "${loop_device}p${suffix}" ]] || {
    echo "missing image partition ${loop_device}p${suffix}" >&2
    exit 1
  }
done
[[ $(blkid -s TYPE -o value "${loop_device}p1") == ntfs ]]
[[ $(blkid -s TYPE -o value "${loop_device}p2") == vfat ]]
[[ $(blkid -s TYPE -o value "${loop_device}p3") == squashfs ]]
mount -o ro,nosuid,nodev,noexec "${loop_device}p1" "$mount_dir/ntfs"
mount -o ro,nosuid,nodev,noexec "${loop_device}p2" "$mount_dir/boot"
mount -t squashfs -o ro,nosuid,nodev,noexec "${loop_device}p3" "$mount_dir/rootfs"

for path in LICENSE bsp/uboot.img; do
  [[ -s "$mount_dir/ntfs/$path" ]] || {
    echo "missing data partition file: $path" >&2
    exit 1
  }
done
[[ -d "$mount_dir/ntfs/images" && -s "$mount_dir/ntfs/images/multitool" ]]
readelf -h "$mount_dir/ntfs/images/multitool" | grep -q 'Machine:.*ARM'
for path in kernel.img rk322x-box.dtb extlinux/extlinux.conf; do
  [[ -s "$mount_dir/boot/$path" ]] || {
    echo "missing boot partition file: $path" >&2
    exit 1
  }
done
if grep -Eq '#(FAT|SQUASHFS)_PARTUUID#' "$mount_dir/boot/extlinux/extlinux.conf"; then
  echo 'boot configuration has unresolved partition UUIDs' >&2
  exit 1
fi
[[ -s "$mount_dir/rootfs/usr/local/bin/multitool" ]]
readelf -h "$mount_dir/rootfs/usr/local/bin/multitool" | grep -q 'Machine:.*ARM'
