#!/usr/bin/env bash
set -Eeuo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
image="$work/recovery.img.xz"
manifest="$work/recovery.manifest.json"
commit=$(git -C "$root" rev-parse HEAD)
printf 'recovery image fixture\n' | xz -c > "$image"

bash "$root/scripts/release-manifest.sh" --channel experimental "$image" "$manifest" rk322x v0.3.0 "$commit"
python3 - "$manifest" "$image" "$commit" << 'PY'
import hashlib
import json
import pathlib
import sys

manifest = json.loads(pathlib.Path(sys.argv[1]).read_text())
image = pathlib.Path(sys.argv[2])
assert manifest["channel"] == "experimental"
assert manifest["source_commit"] == sys.argv[3]
assert manifest["image_sha256"] == hashlib.sha256(image.read_bytes()).hexdigest()
assert set(manifest["boot_components"]) == {
    "kernel", "device_tree", "idbloader", "u_boot", "legacy_u_boot", "trust_os"
}
assert all(component["version_id"].startswith("sha256:") for component in manifest["boot_components"].values())
PY

if bash "$root/scripts/release-manifest.sh" --channel stable "$image" "$work/rejected.json" rk322x v0.3.0 "$commit" > /dev/null 2>&1; then
  echo 'stable image was accepted without hardware evidence' >&2
  exit 1
fi
[[ ! -e "$work/rejected.json" ]]

cat > "$work/approval.json" << 'JSON'
{
  "board": "rk322x",
  "firmware_version": "v0.3.0",
  "result": "passed",
  "tested_by": "hardware-reviewer",
  "tested_at": "2026-11-02T00:00:00Z",
  "evidence": "lab-run-42"
}
JSON
bash "$root/scripts/release-manifest.sh" --channel stable "$image" "$work/stable.json" rk322x v0.3.0 "$commit" "$work/approval.json"
python3 - "$work/stable.json" << 'PY'
import json
import pathlib
import sys

manifest = json.loads(pathlib.Path(sys.argv[1]).read_text())
assert manifest["channel"] == "stable"
assert manifest["hardware_validation"]["result"] == "passed"
PY

cat > "$work/failed-approval.json" << 'JSON'
{
  "board": "rk322x",
  "firmware_version": "v0.3.0",
  "result": "failed",
  "tested_by": "hardware-reviewer",
  "tested_at": "2026-11-02T00:00:00Z",
  "evidence": "lab-run-43"
}
JSON
if bash "$root/scripts/release-manifest.sh" --channel stable "$image" "$work/failed.json" rk322x v0.3.0 "$commit" "$work/failed-approval.json" > /dev/null 2>&1; then
  echo 'stable image was accepted after failed hardware validation' >&2
  exit 1
fi
[[ ! -e "$work/failed.json" ]]

if bash "$root/scripts/release-manifest.sh" --channel stable "$image" "$work/wrong-version.json" rk322x v0.4.0 "$commit" "$work/approval.json" > /dev/null 2>&1; then
  echo 'stable image was accepted with mismatched hardware evidence' >&2
  exit 1
fi
[[ ! -e "$work/wrong-version.json" ]]
