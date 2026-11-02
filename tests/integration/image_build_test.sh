#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$(mktemp -d)
image="$work/fixture.img"
loop_device=
cleanup() {
  local status=$?
  trap - EXIT
  if [[ -n "$loop_device" ]] && losetup "$loop_device" > /dev/null 2>&1; then
    for partition in "${loop_device}p1" "${loop_device}p2" "${loop_device}p3"; do
      while read -r mountpoint; do
        [[ -n "$mountpoint" ]] && umount "$mountpoint" || true
      done < <(findmnt --raw --noheadings --source "$partition" --output TARGET 2> /dev/null || true)
    done
    losetup -d "$loop_device" || status=1
  fi
  rm -rf "$work"
  exit "$status"
}
trap cleanup EXIT

if [[ $(uname -s) != Linux || $(id -u) -ne 0 ]]; then
  if bash "$root/scripts/build-image.sh" invalid > /dev/null 2>&1; then
    echo 'invalid image target unexpectedly succeeded' >&2
    exit 1
  fi
  if bash "$root/scripts/validate-image.sh" /dev/null > /dev/null 2>&1; then
    echo 'validator accepted a block device' >&2
    exit 1
  fi
  exit 0
fi

for tool in blkid losetup mount umount mountpoint findmnt sfdisk parted mkfs.ntfs mkfs.vfat mksquashfs readelf python3 truncate dd; do
  command -v "$tool" > /dev/null || {
    echo "missing image integration test tool: $tool" >&2
    exit 1
  }
done

truncate -s 256M "$image"
parted -s "$image" mktable msdos
parted -s "$image" unit MiB mkpart primary ntfs 1 65
parted -s "$image" unit MiB mkpart primary fat32 65 129
parted -s "$image" unit MiB mkpart primary 129 255
parted -s "$image" set 2 boot on
loop_device=$(losetup --find --show --partscan "$image")
for suffix in 1 2 3; do
  for _ in {1..30}; do
    [[ -b "${loop_device}p${suffix}" ]] && break
    sleep 1
  done
  [[ -b "${loop_device}p${suffix}" ]]
done
mkfs.ntfs -F -L MULTITOOL "${loop_device}p1" > /dev/null
mkfs.vfat -F 32 -n BOOTSTRAP "${loop_device}p2" > /dev/null
mkdir "$work/ntfs" "$work/boot" "$work/squash-root"
mount "${loop_device}p1" "$work/ntfs"
mkdir -p "$work/ntfs/images" "$work/ntfs/bsp"
cp "$root/LICENSE" "$work/ntfs/LICENSE"
printf 'fixture\n' > "$work/ntfs/bsp/uboot.img"
bash "$root/scripts/build-armv7.sh" "$work/ntfs/images/multitool"
chmod 0755 "$work/ntfs/images/multitool"
umount "$work/ntfs"
mount "${loop_device}p2" "$work/boot"
printf 'fixture\n' > "$work/boot/kernel.img"
printf 'fixture\n' > "$work/boot/rk322x-box.dtb"
mkdir "$work/boot/extlinux"
printf 'LINUX /kernel.img\nFDT /rk322x-box.dtb\nAPPEND root=PARTUUID=fixture boot=UUID=fixture\n' > "$work/boot/extlinux/extlinux.conf"
umount "$work/boot"
mkdir -p "$work/squash-root/usr/local/bin"
bash "$root/scripts/build-armv7.sh" "$work/squash-root/usr/local/bin/multitool"
mksquashfs "$work/squash-root" "$work/rootfs.squashfs" -noappend -all-root > /dev/null
dd if="$work/rootfs.squashfs" of="${loop_device}p3" bs=4k conv=sync,fsync status=none
losetup -d "$loop_device"
loop_device=
bash "$root/scripts/validate-image.sh" "$image"
[[ -z $(losetup --associated "$image" --noheadings --output NAME) ]]

# A damaged root filesystem must fail validation and release its loop device.
loop_device=$(losetup --find --show --partscan "$image")
for _ in {1..30}; do
  [[ -b "${loop_device}p3" ]] && break
  sleep 1
done
dd if=/dev/zero of="${loop_device}p3" bs=1M count=1 conv=notrunc status=none
losetup -d "$loop_device"
loop_device=
if bash "$root/scripts/validate-image.sh" "$image" > /dev/null 2>&1; then
  echo 'validator accepted an image with a damaged root filesystem' >&2
  exit 1
fi
[[ -z $(losetup --associated "$image" --noheadings --output NAME) ]]
