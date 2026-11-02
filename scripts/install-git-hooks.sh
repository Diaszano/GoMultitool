#!/usr/bin/env bash
set -Eeuo pipefail

git rev-parse --show-toplevel > /dev/null
git config core.hooksPath .githooks
