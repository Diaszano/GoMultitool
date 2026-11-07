# Go backup and image burn — design

Date: 2026-11-07
Base branch: `feature/go-backup-and-burn`, cut from `fix/audit-findings` (PR #10). It depends on PR #10's `application.CopyExact`, `application.Target`, and busy-device detection.

## Intent

The person using this tool puts a Linux image (Armbian) on the eMMC of an RK322x TV box to turn it into a home server. Their real path through the legacy Bash menu (`sources/multistrap/multitool.sh`) is: **back up the eMMC, then burn an image with the normal burn option**. Today that path has two dangerous bugs:

- The menu reports success when backup or burn fails (it reads `dialog`'s exit status, not `dd`'s).
- File names with spaces break the file picker.

This project moves exactly that path into the Go binary, makes the Go TUI the screen the box boots into, and keeps the Bash menu available from inside the Go menu for everything not yet migrated. The whole Bash menu will be removed later; this is the first step.

## Goals

1. **Go TUI is the boot screen** once hardware validation passes (see Rollout). Main menu, in this order:
   - Gravar imagem
   - Fazer backup
   - Informações de hardware
   - Menu antigo (Bash)
   - Sair

   The current "Em breve" items are removed.
2. **Burn image (single screen).** Pick the destination and the image on one screen, then press Gravar. Requirements:
   - Destination list: boot media is not offered. Busy devices (mounted, active swap, holders) are shown as "em uso" and cannot be selected.
   - Image list: files in `MULTITOOL/images` ending in `.img`, `.img.xz`, or `.img.gz` (case-insensitive). Names with spaces and apostrophes work. Show the name and the uncompressed size.
   - Uncompressed size (needed by `installer.Writer`, the too-large check and the percent bar) *(builder's call)*:
     - `.img`: the file size.
     - `.img.xz`: `xz --robot --list`, which reads the exact size from the xz index without decompressing.
     - `.img.gz`: a counting pass (`compression.Inspect`) when the user picks the file, with a "calculando tamanho…" note. The gzip trailer only stores the size modulo 4 GiB, so it isn't used. `.gz` images are rare; Armbian ships `.img.xz`.
   - Show the warning "Tudo em <dispositivo> será apagado." before writing.
   - Progress bar with bytes done/total and speed, updated without freezing the UI.
   - After writing: sync, read back, and compare SHA-256. This is the existing `installer.Writer` behavior.
   - **DDR command rate (rk322x only):** `MULTITOOL/TARGET` contains `rk322x`. Before writing, read the command rate from the destination; after writing, put it back. See "Command rate" below. If it can't be read before writing, skip silently, as Bash does. If it was read, but the new image has no signature or the byte there is not 0/1, the burn still counts as successful and the result screen adds a warning line: "A imagem não tem o command rate da DDR; o valor anterior (1T/2T) não foi aplicado." *(builder's call: the image was written and verified, so this isn't a failure)*. A failed write of the byte itself is an error.
   - Success is shown only when every step succeeded. Any failure shows the error and never says "concluído".
3. **Backup (single screen).** Pick the source device and type a file name, then press Fazer backup. Requirements:
   - Sources: the same filter as burn destinations. Boot media and busy devices are not offered, because a busy device can change while it's being read *(builder's call)*.
   - The name defaults to `tvbox-backup-YYYY-MM-DD`. Spaces are allowed. Reject empty names, names containing `/`, and names starting with `.`.
   - Output is `MULTITOOL/backups/<name>.img.gz`, gzip-compressed, so the Bash "Restore flash" (which reads `backups/*.gz` through `pigz -d`) can restore it. Next to it, write `<name>.img.gz.json`: the existing `backup.Manifest` with the SHA-256 and size of the **uncompressed** device data.
   - If either file already exists, refuse and ask for another name. Never overwrite.
   - Free space: show the free space on the MULTITOOL partition next to the device size, but don't refuse up front. The gzip size can't be known in advance, and the MULTITOOL partition is often smaller than the eMMC even when a compressed backup fits. If the disk fills up, stop, delete the partial file, and show "Sem espaço na partição MULTITOOL" *(builder's call; replaces the worst-case pre-check)*. The `availableBytes` check in `backup.Service.Create` must be removed or changed to match.
   - Progress bar like the burn screen.
   - Write to a temp file in the same folder, fsync, rename, then fsync the folder (PR #10 behavior). If the folder fsync returns `EINVAL` or `ENOTSUP` (some NTFS/FUSE drivers), ignore it; any other error is reported *(builder's call; confirm on the box)*. On any failure, remove the partial file and show the error.
4. **Menu antigo (Bash).** Suspend the TUI, run `/usr/local/bin/multitool.sh` on the same terminal, and return to the Go menu when it exits. The Bash menu keeps all its current items, including its own backup and burn.
5. **Fix the second audit's findings that touch this path:**
   - Bash: every gauge pipeline (`multitool.sh` backup, restore, burn, steP-nand burn, erase) reports the real exit status (`set -o pipefail` inside the subshell, read `${PIPESTATUS[0]}`).
   - Bash: quote file names in `choose_file` and the decompression command lines (`printf '%q'` instead of `'$TARGET'`).
   - Add `sources/multistrap/multitool.sh` to `make lint-shell` (shellcheck). Turn off noisy rules with a top-of-file `# shellcheck disable=` listing each rule.
   - Device listing: resolve the boot device and read `/proc/mounts`, `/proc/swaps` and `/proc/cmdline` once per `List`. Run `blkid` with a 5-second timeout; a timeout means "boot media" (fail-safe). Load the device list off the UI goroutine so the TUI never freezes.
   - Device listing: skip `*boot[0-9]`, `*rpmb`, `zram*`, and size-0 devices. Classify `rknand*` as NAND (re-add `KindNAND`).
   - One shared device-line formatter for TUI and console: path, translated kind, size in GB with one decimal (`7,3 GB` in pt-BR, `7.3 GB` in en), and labels "em uso" / "mídia de boot".
   - i18n: remove `ConsoleTitle`, `VersionLabel`, `ConsoleHardwareOption`, `ConsoleExitOption`; reuse `ProductName` and the menu labels.
   - `cmd/multitool/main.go`: handle `tui` and `--console` inside `run`, with one `discovery.Service`, and test both dispatches.
   - `scripts/release-manifest.sh`: hash the image with `hashlib.file_digest` instead of `read_bytes()`.
   - PR #10 deferred minors: a `holders` read error other than "not exist" means busy, and `CopyExact` maps `io.EOF` to `io.ErrUnexpectedEOF` only when fewer than `size` bytes were copied.

## Non-goals

- Restore, erase, steP-nand, jump start, and the manual "Change DDR Command Rate" menu stay in Bash.
- Image formats beyond `.img`, `.img.xz`, `.img.gz` (no zip, 7z, tar, bz2, lzma).
- NAND as a write destination. NAND is listed and labeled but not offered for burn or backup in this project, because Bash's idbloader-skip logic for `rknand` is not migrated.
- The work LED (`/sys/class/leds/led:state1`) state changes.
- Reboot and shutdown items in the Go menu.

## Anti-goals (failure even if it "works")

- Any screen saying "concluído" or "Backup salvo" when a step failed.
- Writing to the device the box booted from, or to a busy device.
- Overwriting an existing backup.
- A frozen TUI that ignores `q` and Ctrl+C.
- Automated tests that write to a real block device.

## Constraints

- Go 1.27.3, existing deps only (`tview`, `tcell`). No new modules.
- Linux ARMv7 target. The binary runs as root on `tty1` (and over SSH).
- `xz` and `gzip` decompression:
  - `.img.xz` keeps using the external `xz -dc` (`internal/infrastructure/compression`). It is already in the image (`create_image.sh` uses it).
  - `.img.gz` uses Go's `compress/gzip`.
  - The backup's gzip output uses `compress/gzip`.
- The MULTITOOL partition is NTFS, found by label `MULTITOOL` (`blkid -l --label MULTITOOL`, as Bash does). If it isn't mounted, mount it at `/mnt` and unmount it after each operation, as Bash does with `mount_mt_partition`/`unmount_mt_partition`. If it is already mounted, use that mount point and leave it mounted.
- `MULTITOOL/TARGET` holds the platform name (`rk322x`). `MULTITOOL/ISSUE` holds the build.
- Commit messages follow Conventional Commits and must not include AI co-author trailers (the repo hook rejects them).

## One operation at a time

The Go binary takes an exclusive `flock` on `/run/multitool.lock` for the whole of a backup or burn. A second copy (for example over SSH) gets "outra operação em andamento" instead of starting. The Bash menu does not use the lock; running both at once stays an accepted risk until the Bash menu is removed *(builder's call)*.

## Command rate (exact behavior to port)

From `multitool.sh` `get_command_rate_offset`, `get_command_rate`, `set_command_rate`:

- Read the first 128 KiB of the device.
- Find the first 4-byte-aligned offset `S` where bytes are `65 65 86 a7`. That is the 32-bit little-endian word `0xa7866565`, which `od -tx4` prints as `a7866565`.
- The command-rate byte is at `S - 16`. Its value must be `0` (1T) or `1` (2T); anything else means "not found".
- Before burn: read the byte from the destination device. Not found means skip restoring.
- After burn and readback: find the signature again on the destination. If it is found and the current byte is 0 or 1, write the saved byte at the new `S - 16`, then sync. Otherwise show the warning described in Goals 2.
- Unit-test the find/read/write logic on in-memory byte slices: signature found, not aligned, missing, invalid byte, and round trip.

## Rollout

- `build/rootfs/multitool-init` already gates the Go runtime behind `MULTITOOL_GO_RUNTIME=1`. Change it so the gated path launches `multitool tui` (instead of `--console`).
- `sources/multistrap/init` keeps launching Bash by default until the owner validates backup and burn on a real box. Then a final commit switches `init` to launch `/usr/local/sbin/multitool-init` with the Go runtime on. That commit is part of this project but is made only after the owner reports hardware success.
- `init` restarts whatever it launched in a loop, so "Sair" in the Go menu returns to the same Go menu. That is acceptable.

## Decided by the owner

- Backup and burn first; everything else later.
- Boot into the Go screen, with an item to open the old Bash menu.
- Burn uses a single screen (mockup B), and backup uses the same style.
- The Bash menu keeps showing its own backup and burn items until the whole Bash menu is removed.

## Left to the builder

- TUI widget layout details (tview `Form` vs custom `List`s), as long as it matches the single-screen mockups.
- How progress is reported from the copy into the UI (a counting writer plus `app.QueueUpdateDraw` is the expected shape).
- Package placement for the MULTITOOL partition locator and the command-rate code, following the existing `application` / `infrastructure` split.
- Exact wording of error messages, in both locales.
