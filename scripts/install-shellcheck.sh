#!/usr/bin/env bash
set -Eeuo pipefail

version=$(sed -n 's/^shellcheck=//p' .tools-versions)
case "$(uname -m)" in
  x86_64)
    asset_arch=x86_64
    sha256=8c3be12b05d5c177a04c29e3c78ce89ac86f1595681cab149b65b97c4e227198
    ;;
  aarch64)
    asset_arch=aarch64
    sha256=12b331c1d2db6b9eb13cfca64306b1b157a86eb69db83023e261eaa7e7c14588
    ;;
  *)
    echo "unsupported ShellCheck host architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
archive="shellcheck-$version.linux.$asset_arch.tar.xz"
curl -fsSL "https://github.com/koalaman/shellcheck/releases/download/$version/$archive" -o "$work/$archive"
printf '%s  %s\n' "$sha256" "$work/$archive" | sha256sum -c -
tar -xJf "$work/$archive" -C "$work"
install -m 0755 "$work/shellcheck-$version/shellcheck" "$(go env GOPATH)/bin/shellcheck"
