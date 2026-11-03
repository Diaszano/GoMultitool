#!/usr/bin/env bash
set -Eeuo pipefail

version=$(sed -n 's/^zizmor=//p' .tools-versions)
case "$(uname -m)" in
  x86_64)
    asset_arch=x86_64-unknown-linux-gnu
    sha256=d185f726bdaf92df78273217dfdd24ebcd6b0ef48cf3ce776ff5762b6e138c99
    ;;
  aarch64)
    asset_arch=aarch64-unknown-linux-gnu
    sha256=37eebbd5762a44f262d94875431b14a6d2402991e90a685113ea97b8b41be833
    ;;
  *)
    echo "unsupported zizmor host architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
archive="zizmor-$asset_arch.tar.gz"
curl -fsSL "https://github.com/zizmorcore/zizmor/releases/download/$version/$archive" -o "$work/$archive"
printf '%s  %s\n' "$sha256" "$work/$archive" | sha256sum -c -
tar -xzf "$work/$archive" -C "$work"
install -m 0755 "$work/zizmor" "$(go env GOPATH)/bin/zizmor"
