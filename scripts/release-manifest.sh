#!/usr/bin/env bash
set -Eeuo pipefail

usage() {
  echo 'usage: release-manifest.sh --channel experimental IMAGE OUTPUT TARGET VERSION COMMIT' >&2
  echo '       release-manifest.sh --channel stable IMAGE OUTPUT TARGET VERSION COMMIT HARDWARE_APPROVAL.json' >&2
  exit 2
}

[[ ${1:-} == --channel && $# -ge 2 ]] || usage
channel=$2
shift 2
case "$channel" in
  experimental) [[ $# -eq 5 ]] || usage ;;
  stable) [[ $# -eq 6 ]] || usage ;;
  *) usage ;;
esac

image=$1
output=$2
target=$3
version=$4
commit=$5
approval=${6:-}
root=$(cd "$(dirname "$0")/.." && pwd)
[[ "$target" == rk322x && -d "$root/sources/$target" ]] || {
  echo 'unsupported recovery image target' >&2
  exit 1
}
[[ "$version" =~ ^[[:alnum:]][[:alnum:].+_-]*$ && "$commit" =~ ^[[:xdigit:]]{40}$ ]] || {
  echo 'invalid image version or source commit' >&2
  exit 1
}
[[ -f "$image" && ! -L "$image" && -s "$image" ]] || {
  echo 'image must be a non-empty regular file' >&2
  exit 1
}
output_tmp="$output.tmp"
[[ ! -L "$output" && ! -L "$output_tmp" ]] || {
  echo 'refusing to write through a manifest symlink' >&2
  exit 1
}
trap 'rm -f "$output_tmp"' EXIT
xz -t "$image"

if [[ "$channel" == stable ]]; then
  [[ -f "$approval" && ! -L "$approval" ]] || {
    echo 'stable firmware requires a hardware approval record' >&2
    exit 1
  }
  python3 - "$approval" "$target" "$version" << 'PY'
import datetime
import json
import pathlib
import sys

record = json.loads(pathlib.Path(sys.argv[1]).read_text())
try:
    tested_at = datetime.datetime.fromisoformat(record.get("tested_at", "").replace("Z", "+00:00"))
except (AttributeError, ValueError):
    raise SystemExit("hardware approval has an invalid test timestamp")
if (
    record.get("board") != sys.argv[2]
    or record.get("firmware_version") != sys.argv[3]
    or record.get("result") != "passed"
    or not record.get("tested_by")
    or tested_at.tzinfo is None
    or not record.get("evidence")
):
    raise SystemExit("hardware approval does not match this firmware version")
PY
fi

python3 - "$image" "$output_tmp" "$target" "$version" "$commit" "$root" "$channel" "$approval" << 'PY'
import hashlib
import json
import pathlib
import sys

image, output, target, version, commit, root, channel, approval = sys.argv[1:]
root = pathlib.Path(root)
def component(path):
    digest = hashlib.sha256((root / path).read_bytes()).hexdigest()
    return {"file": path, "version_id": f"sha256:{digest}", "sha256": digest}
manifest = {
    "channel": channel,
    "target": target,
    "version": version,
    "source_commit": commit,
    "image": pathlib.Path(image).name,
    "image_sha256": hashlib.sha256(pathlib.Path(image).read_bytes()).hexdigest(),
    "boot_components": {
        name: component(path)
        for name, path in {
            "kernel": "sources/rk322x/kernel.img",
            "device_tree": "sources/rk322x/rk322x-box.dtb",
            "idbloader": "sources/rk322x/idbloader_ddr_1.11.img",
            "u_boot": "sources/rk322x/u-boot-dtb.bin",
            "legacy_u_boot": "sources/rk322x/legacy-rk-uboot.bin",
            "trust_os": "sources/rk322x/rk322x_optee_os_4.5.0.bin",
        }.items()
    },
}
if approval:
    manifest["hardware_validation"] = json.loads(pathlib.Path(approval).read_text())
pathlib.Path(output).write_text(json.dumps(manifest, indent=2) + "\n")
PY
mv "$output_tmp" "$output"
trap - EXIT
