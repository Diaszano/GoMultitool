#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
for script in scripts/test-go.sh scripts/build-armv7.sh; do
  [[ -f "$root/$script" ]] || {
    echo "missing build helper: $script" >&2
    exit 1
  }
done
tmp=$(mktemp -d "$root/tests/failing.XXXXXX")
trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/failing_test.go" << 'GO'
package failing

import "testing"

func TestQualityGateRejectsFailure(t *testing.T) { t.Fatal("expected fixture failure") }
GO
relative=${tmp#"$root/"}
if make -C "$root" test GO_TEST_PACKAGES="./$relative" > /dev/null 2>&1; then
  echo 'Go quality gate accepted a failing test' >&2
  exit 1
fi
make -C "$root" build-armv7 > /dev/null
readelf -h "$root/build/multitool-linux-armv7" | grep -Eq 'Machine:[[:space:]]+ARM'
