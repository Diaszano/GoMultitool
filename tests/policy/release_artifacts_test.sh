#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
config="$root/.goreleaser.yaml"
[[ -f "$config" ]] || {
  echo 'missing GoReleaser configuration' >&2
  exit 1
}

cd "$root"
PATH="$(go env GOPATH)/bin:$PATH" goreleaser check
PATH="$(go env GOPATH)/bin:$PATH" goreleaser release --snapshot --clean
archives=("$root"/dist/*.tar.gz)
[[ ${#archives[@]} == 2 ]] || {
  echo "expected two release archives, got ${#archives[@]}" >&2
  exit 1
}
for archive in "${archives[@]}"; do
  contents=$(tar -tzf "$archive")
  grep -qx 'multitool' <<< "$contents"
  grep -qx 'LICENSE' <<< "$contents"
  grep -qx 'README.md' <<< "$contents"
done
checksums=("$root"/dist/*_checksums.txt)
[[ ${#checksums[@]} == 1 ]]
(cd "$root/dist" && sha256sum -c "$(basename "${checksums[0]}")")
amd_archive=$(printf '%s\n' "${archives[@]}" | grep '_amd64\.tar\.gz$')
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
tar -xzf "$amd_archive" -C "$tmp"
version_output=$("$tmp/multitool" version)
grep -q 'RK322x Multitool 0.0.1-next' <<< "$version_output"
arm_archive=$(printf '%s\n' "${archives[@]}" | grep '_armv7\.tar\.gz$')
tar -xzf "$arm_archive" -C "$tmp"
readelf -h "$tmp/multitool" | grep -Eq 'Machine:[[:space:]]+ARM'
