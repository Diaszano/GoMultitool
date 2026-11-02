# Legacy recovery runtime

The recovery image starts `sources/multistrap/init`, mounts its labeled `MULTITOOL` partition at `/mnt`, reads `ISSUE` and `TARGET`, displays the license, then invokes the installed `multitool-init` dispatcher on the active console. The dispatcher defaults to `sources/multistrap/multitool.sh` through `dialog`; its Go console path requires `MULTITOOL_GO_RUNTIME=1`. The shell discovers MMC devices from `/sys/bus/mmc/devices` and has a special path for `/dev/rknand0`.

## Existing menu operations

The RK322x menu offers backup flash, restore flash, erase flash, Bash shell, burn image, change DDR command rate, reboot, and shutdown. NAND detection additionally exposes Jump start installation and steP-nand Armbian installation. Other board configurations may expose a different set. The exact RK322x menu labels are recorded in [`../../tests/fixtures/legacy-menu.txt`](../../tests/fixtures/legacy-menu.txt).

The script backs up selected MMC devices as gzip streams into `MULTITOOL/backups`, restores selected gzip backups, burns supported image/archive inputs, and has board-specific loader operations. These operations can erase or overwrite storage. The legacy code does not provide a versioned integrity manifest; a successful dialog progress display is not proof of a bootable result.

## Migration boundary

The Go application may replace menu presentation, read-only inventory, image inspection, backup/restore orchestration, and diagnostics only as corresponding safeguards and recovery fallbacks are implemented. `create_image.sh`, board boot components, and the existing shell flow remain the fallback until equivalent behavior is validated on the reference hardware. Bootloader, kernel, and TrustZone firmware are preserved as supplied.

The recovery root filesystem now includes the ARMv7 Go executable and an opt-in `multitool-init` dispatcher. The existing `sbin/init` invokes the dispatcher; setting `MULTITOOL_GO_RUNTIME=1` in a controlled test environment selects the Go console, while the default path remains the legacy menu. `util-linux` provides the existing `mount`, `blkid`, and `lsblk` commands.

## Source and image evidence

The repository contains source and boot-component files but no complete known-working recovery-card image. SHA-256 values for the tracked RK322x components are recorded in [`../supported-boards/rk3229-reference.md`](../supported-boards/rk3229-reference.md); they identify repository inputs, not a verified assembled image. No claim of successful physical boot is made by this source review.
