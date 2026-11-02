#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
script="$root/scripts/build-image.sh"
dollar='$'
if grep -Fq "rm -f \"${dollar}archive\" \"${dollar}checksum\" \"${dollar}manifest\"" "$script" || grep -Fq 'dist-rk322x/root.img' "$script"; then
  echo 'image build must preserve existing release artifacts and rootfs data' >&2
  exit 1
fi
grep -q 'refusing to overwrite existing build output' "$script"
grep -q 'ROOTFS_IMAGE=' "$script"
grep -q 'ROOTFS_IMAGE=' "$root/create_image.sh"
