#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
target=${1:-rk322x}
[[ "$target" == rk322x && -f "$root/sources/$target.conf" ]] || {
  echo 'only the verified rk322x target is supported' >&2
  exit 2
}
[[ $(uname -s) == Linux && $(id -u) -eq 0 ]] || {
  echo 'image construction requires root on Linux; it only attaches the generated image file' >&2
  exit 1
}
for tool in go git multistrap mksquashfs fallocate losetup parted mkfs.vfat mkfs.ntfs mount umount mountpoint findmnt blkid sfdisk readelf xz sha256sum python3; do
  command -v "$tool" > /dev/null || {
    echo "missing required image build tool: $tool" >&2
    exit 1
  }
done

[[ ! -e "$root/sources/multistrap/overlays/usr/local/bin/multitool" && ! -L "$root/sources/multistrap/overlays/usr/local/bin/multitool" ]] || {
  echo 'refusing to overwrite an existing recovery rootfs binary' >&2
  exit 1
}
version=${IMAGE_VERSION:-$(git -C "$root" describe --tags --always --dirty)}
[[ "$version" =~ ^[[:alnum:]][[:alnum:].+_-]*$ ]] || {
  echo 'image version may contain only letters, numbers, dots, plus, underscores, and hyphens' >&2
  exit 2
}
commit=$(git -C "$root" rev-parse HEAD)
output_dir="$root/image-dist"
raw_dir="$root/dist-$target"
raw_image="$raw_dir/multitool.img"
archive="$output_dir/multitool-$target-$version.img.xz"
checksum="$archive.sha256"
manifest="$output_dir/multitool-$target-$version.manifest.json"
overlay_binary="$root/sources/multistrap/overlays/usr/local/bin/multitool"
for path in "$output_dir" "$raw_dir" "$root/sources/multistrap/overlays" "$root/sources/multistrap/overlays/usr" "$root/sources/multistrap/overlays/usr/local" "$root/sources/multistrap/overlays/usr/local/bin"; do
  [[ ! -L "$path" && (! -e "$path" || -d "$path") ]] || {
    echo "refusing unsafe image build directory: $path" >&2
    exit 1
  }
done
mkdir -p "$output_dir" "$raw_dir" "$(dirname "$overlay_binary")"
for path in "$archive" "$checksum" "$manifest" "$archive.tmp" "$checksum.tmp" "$manifest.tmp" "$raw_image"; do
  [[ ! -e "$path" && ! -L "$path" ]] || {
    echo "refusing to overwrite existing build output: $path" >&2
    exit 1
  }
done
temp_dir=$(mktemp -d)
loop_device=
cleanup() {
  local status=$?
  trap - EXIT
  if [[ -n "$loop_device" ]] && losetup "$loop_device" > /dev/null 2>&1; then
    losetup -d "$loop_device" || status=1
  fi
  while read -r device; do
    [[ -n "$device" ]] || continue
    for partition in "$device" "${device}p1" "${device}p2" "${device}p3"; do
      while read -r mountpoint; do
        [[ -n "$mountpoint" ]] && umount "$mountpoint" || true
      done < <(findmnt --raw --noheadings --source "$partition" --output TARGET 2> /dev/null || true)
    done
    losetup -d "$device" || true
  done < <(losetup --associated "$raw_image" --noheadings --output NAME 2> /dev/null || true)
  rm -rf "$temp_dir"
  rm -f "$overlay_binary"
  rmdir "$(dirname "$overlay_binary")" 2> /dev/null || true
  if [[ $status -ne 0 ]]; then
    rm -f "$archive.tmp" "$checksum.tmp" "$manifest.tmp" "$raw_image"
  fi
  rm -f "$raw_image"
  exit "$status"
}
trap cleanup EXIT

bash "$root/scripts/build-armv7.sh" "$temp_dir/multitool"
readelf -h "$temp_dir/multitool" | grep -q 'Machine:.*ARM'
bash "$root/build/rootfs/install-multitool.sh" "$root/sources/multistrap/overlays" "$temp_dir/multitool"
(cd "$root" && bash create_image.sh "$target")

loop_device=$(losetup --find --show --partscan "$raw_image")
for _ in {1..30}; do
  [[ -b "${loop_device}p1" ]] && break
  sleep 1
done
[[ -b "${loop_device}p1" ]] || {
  echo 'image data partition did not appear' >&2
  exit 1
}
mkdir "$temp_dir/data"
mount "${loop_device}p1" "$temp_dir/data"
[[ -d "$temp_dir/data/images" ]] || {
  echo 'image data partition is missing its images directory' >&2
  exit 1
}
install -m 0755 "$temp_dir/multitool" "$temp_dir/data/images/multitool"
umount "$temp_dir/data"
losetup -d "$loop_device"
loop_device=
bash "$root/scripts/validate-image.sh" "$raw_image"

xz -T0 -6 --check=crc64 -c "$raw_image" > "$archive.tmp"
mv "$archive.tmp" "$archive"
(cd "$output_dir" && sha256sum "$(basename "$archive")") > "$checksum.tmp"
mv "$checksum.tmp" "$checksum"
bash "$root/scripts/release-manifest.sh" --channel experimental "$archive" "$manifest" "$target" "$version" "$commit"
