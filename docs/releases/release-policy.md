# Firmware release policy

Recovery image builds are experimental by default. The image workflow uploads only an artifact named `recovery-image-experimental-*`, and its provenance manifest records `channel: experimental`. It does not create or attach firmware to a GitHub Release.

The manifest identifies the source commit, image checksum, board target, and each boot component by its SHA-256 content identifier. The legacy kernel and bootloader blobs do not carry complete upstream version metadata, so their checksum is the version identifier.

Stable manifests require a hardware approval JSON record that matches the target and firmware version, reports `result: passed`, and includes the tester, test time, and evidence reference. `scripts/release-manifest.sh --channel stable` rejects missing or mismatched evidence. This repository does not yet configure an automated stable-image promotion; a reviewer must verify the hardware record before any stable distribution.

The current Go recovery path has not been booted on a physical RK3229. Follow the [hardware smoke test](../../tests/hardware/rk3229-smoke-test.md) and update the [validation matrix](../supported-boards/validation-matrix.md) before requesting stable approval.

The build does not generate an SBOM because the existing image toolchain has no pinned SBOM generator. The manifest and `.sha256` file provide source and artifact integrity metadata; they are not an SBOM.

A hardware approval record has this shape:

```json
{
  "board": "rk322x",
  "firmware_version": "v1.2.3",
  "result": "passed",
  "tested_by": "reviewer",
  "tested_at": "2026-11-02T00:00:00Z",
  "evidence": "hardware-test-report"
}
```

The manifest script validates these fields and requires the approval to match the image target and version. It does not authenticate the reviewer or publish stable firmware; stable distribution remains a separate human-controlled action.
