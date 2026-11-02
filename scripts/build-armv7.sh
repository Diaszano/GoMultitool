#!/usr/bin/env bash
set -Eeuo pipefail

output=${1:-build/multitool-linux-armv7}
mkdir -p "$(dirname "$output")"
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -o "$output" ./cmd/multitool
