# RK3229 hardware smoke test

Run this only with a verified recovery card and a disposable target for restore checks. Record photos or serial logs and keep the original recovery card unchanged.

## Record before boot

- Board model/revision and SoC marking
- Bootloader and kernel identifiers/checksums
- RAM size and storage devices/capacities
- Recovery card checksum and known-good boot result
- Tester, UTC time, image version, and evidence location

## Checks

1. Boot from microSD and confirm HDMI output and keyboard input.
2. Confirm the TUI starts; exercise the console fallback.
3. Compare eMMC, SD, and USB inventory against `lsblk`, sysfs, and device labels.
4. Confirm the active recovery medium is rejected by every write path.
5. Back up a disposable device and verify its manifest/checksum.
6. Restore only to a disposable device, then verify the restored data.
7. Export boot/network diagnostics and inspect them for secrets.
8. Measure idle and peak memory on a 1 GiB board.
9. Confirm the original recovery card still boots.

Record each result as pass, fail, or not run. A stable release requires all safety checks and a documented recovery route.
