#!/usr/bin/env bash
set -Eeuo pipefail

version=$(sed -n 's/^semantic-release=//p' .tools-versions)
case "$(uname -m)" in
  x86_64)
    asset_arch=amd64
    sha256=b8f518b1aeb1d1742f4e6e91a179b707cea4309a64fe3854f465bf724f04f6d3
    ;;
  aarch64)
    asset_arch=arm64
    sha256=74fb627a2ec43a97986bcae43830b146c24c711b15e5db7cc2f89e205548045c
    ;;
  *)
    echo "unsupported semantic-release host architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
asset="semantic-release_${version}_linux_${asset_arch}"
curl -fsSL "https://github.com/go-semantic-release/semantic-release/releases/download/${version}/semantic-release_${version}_checksums.txt" -o "$work/checksums.txt"
curl -fsSL "https://github.com/go-semantic-release/semantic-release/releases/download/${version}/$asset" -o "$work/$asset"
printf '%s  %s\n' "$sha256" "$work/$asset" | sha256sum -c -
install -m 0755 "$work/$asset" "$(go env GOPATH)/bin/semantic-release"
