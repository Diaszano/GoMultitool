#!/usr/bin/env bash
set -Eeuo pipefail

version=$(sed -n 's/^goreleaser=//p' .tools-versions)
case "$(uname -m)" in
  x86_64)
    asset_arch=x86_64
    sha256=ebc3194ad1d1b1787614405285a0039665cd4c5bd3305215a1b201600f90630b
    ;;
  aarch64)
    asset_arch=arm64
    sha256=ef33741b684c9f62d6a21e7a3c8b051f139ea7b1d301693b67cc819c46226d53
    ;;
  *)
    echo "unsupported GoReleaser host architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
archive="goreleaser_Linux_$asset_arch.tar.gz"
curl -fsSL "https://github.com/goreleaser/goreleaser/releases/download/$version/$archive" -o "$work/$archive"
printf '%s  %s\n' "$sha256" "$work/$archive" | sha256sum -c -
tar -xzf "$work/$archive" -C "$work"
install -m 0755 "$work/goreleaser" "$(go env GOPATH)/bin/goreleaser"
