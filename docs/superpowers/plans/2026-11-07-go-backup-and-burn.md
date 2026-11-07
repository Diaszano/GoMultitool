# Go Backup and Burn Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move backup and image burn from the Bash menu into the Go TUI, make the TUI the boot screen once hardware validation passes, and fix the second audit's findings on that path.

**Architecture:** The TUI (`internal/presentation/tui`) only calls functions passed in a `tui.Services` struct; `cmd/multitool` wires them to the real services:
- `internal/infrastructure/multitool`: MULTITOOL partition, lock, free space.
- `internal/infrastructure/compression`: image formats and size.
- `internal/platform/rk322x`: DDR command rate.
- `installer.Burn`: write, verify, command rate.
- `backup.Service`: gzip backup.

The Bash menu is fixed in place and launched from the TUI.

**Tech Stack:** Go 1.27.3, tview v0.42.0, tcell v2.8.1, standard library. Bash, shellcheck, shfmt.

**Spec:** `docs/superpowers/specs/2026-11-07-go-backup-and-burn-design.md`

## Global Constraints

- No new Go modules.
- Linux ARMv7, runs as root on `tty1`. Tests must also pass on macOS (dev machines): use `syscall` calls that exist on both (`Flock`, `Statfs`), with `int64(...)` conversions on struct fields.
- Automated tests never write to a real block device; use temp files and fake sysfs/procfs trees.
- Never show "concluído" / "Backup salvo" / success unless every step succeeded.
- Conventional Commits; no AI co-author trailers (the repo hook rejects them). Branch: `feature/go-backup-and-burn`.
- Run `golangci-lint fmt` and `golangci-lint run --config .golangci.yml ./...` (golines limit is 100 columns) before each commit.
- pt-BR is the default locale; every new user-facing string exists in both `pt-BR` and `en` in `internal/presentation/i18n/catalog.go`.

## Review Focus

- Backup runs out of space on MULTITOOL halfway: the partial file and its `.json` are gone and the screen shows "Sem espaço na partição MULTITOOL". Test in Task 6.
- A burn source image larger than the destination: refused before the target is opened. Test in Task 8.
- An image name with spaces and an apostrophe (`Armbian (1) d'x.img.xz`): listed, burned, and handled by the Bash picker. Tests in Tasks 1 and 4.
- A second copy of the tool starts a burn while one is running: the second gets "outra operação em andamento". Test in Task 7.
- `blkid` hangs during device listing: `List` returns within the 5-second timeout and treats the device as boot media. Test in Task 2.

---

### Task 1: Fix the Bash menu exit codes and file-name quoting; lint it; stream the release-manifest hash

**Files:**
- Modify: `sources/multistrap/multitool.sh`: `choose_file` (296-320); gauge pipelines at 425, 505, 802, 933, 1036; `get_decompression_cli` command lines (593-682); `get_decompression_cli $IMAGE_SOURCE` calls (763 and the steP-nand twin).
- Modify: `Makefile:31` (`lint-shell` adds `sources/multistrap/multitool.sh`)
- Modify: `scripts/release-manifest.sh:91`
- Create: `tests/policy/legacy_menu_test.sh`
- Modify: `Makefile` `integration-test` target, to run the new test

- [ ] **Step 1: Write the failing policy test `tests/policy/legacy_menu_test.sh`**

Follow the style of the other `tests/policy/*.sh` files (`set -Eeuo pipefail`, `root=...`). Assertions:
1. Every line matching `) 2>&1 | dialog` is preceded, inside the same subshell, by `set -o pipefail;`. Check that each `(` that opens such a pipeline starts with `(set -o pipefail;`.
2. The first non-blank line after each gauge pipeline block is `ERR=${PIPESTATUS[0]}`.
3. `choose_file` works with tricky names. Extract the function with `sed -n '/^function choose_file()/,/^}/p'`, define a fake `dialog` that prints `1` to stderr, and set `TTY_CONSOLE=/dev/null`. Create a temp dir with `a b.img` and `c'd.img`, and call `choose_file t t "$dir/*"`. It must print `$dir/c'd.img`.

- [ ] **Step 2: Run it and confirm it fails**

Run: `bash tests/policy/legacy_menu_test.sh`
Expected: FAIL on the first gauge pipeline.

- [ ] **Step 3: Fix the script**

- Each gauge: `(set -o pipefail; <pipeline>) 2>&1 | dialog ...`, then `ERR=${PIPESTATUS[0]}`. Do the same for the two un-gauged `(eval "$OPERATION_CLI")` calls (keep `ERR=$?` there, plus `set -o pipefail;` inside).
- `choose_file`: `FILES+=("$FILE")`, `BASENAME=$(basename "$FILE")`, `echo "${FILES[$CHOICE]}"`. Quote callers: `IMAGE_SOURCE=$(choose_file ...)`, `get_decompression_cli "$IMAGE_SOURCE"`, `BASENAME=$(basename "$IMAGE_SOURCE")`.
- In `get_decompression_cli`, build every path inside `D_COMMAND_LINE` with `$(printf '%q' "$TARGET")` and `$(printf '%q' "$CANDIDATE_IMAGE")` instead of `'$TARGET'`.
- Add the file to `lint-shell`. Run shellcheck. Add one `# shellcheck disable=SCxxxx,...` line after the shebang, listing only the rules that still fire in code this task doesn't touch. Do not reformat the file with `shfmt -w`; exclude it from the shfmt check in `scripts/check-shell.sh` by running shfmt only on the files that aren't `sources/multistrap/multitool.sh`.
- `release-manifest.sh`: `"image_sha256": hashlib.file_digest(open(image, "rb"), "sha256").hexdigest()`.

- [ ] **Step 4: Run the checks**

Run: `bash tests/policy/legacy_menu_test.sh && make lint-shell && bash tests/policy/release_manifest_test.sh`
Expected: all pass. (`release_manifest_test.sh` may need `xz`; skip it locally if missing and note that in the commit body.)

- [ ] **Step 5: Commit**

```bash
git add sources/multistrap/multitool.sh Makefile scripts tests/policy/legacy_menu_test.sh
git commit -m "fix(legacy): report real exit status and quote image file names"
```

---

### Task 2: Device listing: one snapshot per `List`, `blkid` timeout, skip pseudo-devices, NAND

**Files:**
- Modify: `internal/infrastructure/linux/sysfs.go`, `internal/infrastructure/linux/procfs.go`
- Modify: `internal/domain/device/device.go`, `internal/domain/device/safety.go`
- Test: `internal/infrastructure/linux/sysfs_test.go`, `internal/domain/device/safety_test.go`

**Interfaces:**
- Produces: `device.KindNAND Kind = "nand"`. `Device.Validate` accepts it. `Device.ValidateWriteTarget` rejects it with `ErrUnsafeDevice`.
- Produces: `SysfsCatalog.BlkidTimeout time.Duration` (zero means 5s).
- Internal: `type hostState struct { mounts, swaps []string; swapsErr, mountsErr bool; bootRoot string; bootKnown bool }`, built once by `func (c SysfsCatalog) snapshot(ctx context.Context) hostState`. `isBusy(name, path string, s hostState) bool` and `isBootMedia(path string, s hostState) bool` read only from it. The fail-safe rules stay as they are: unreadable mounts → busy; unreadable cmdline, unknown root, `blkid` error, timeout, or ambiguous result → boot media.

- [ ] **Step 1: Write the failing tests**

All tests build trees with `copyFixtureBlockDevices` plus extra dirs.
- `TestSysfsCatalogSkipsPseudoDevices`: add `block/mmcblk0boot0`, `block/mmcblk0rpmb`, `block/zram0`, each with `size` 8192 and `queue/logical_block_size` 512, and `block/sdc` with `size` 0. Assert none of these IDs appear in `List`.
- `TestSysfsCatalogClassifiesRknandAsNAND`: add `block/rknand0` (size 8192, block size 512, no `device/type`). Assert `Kind == device.KindNAND`.
- `TestSysfsCatalogBlkidTimeoutMeansBootMedia`: set PATH to a fake `blkid` that runs `sleep 5`, `cmdline` = `root=PARTUUID=x`, no dev links, `BlkidTimeout: 200*time.Millisecond`. Assert `List` returns in under 2s and every device has `BootMedia == true`.
- `TestSysfsCatalogRunsBlkidOncePerList`: use a fake `blkid` that appends a line to a counter file and prints `/dev/mmcblk1p3`. After one `List` the file has exactly 1 line.
- `TestSysfsCatalogUnreadableHoldersMeansBusy`: create `block/mmcblk0/holders` as a regular file instead of a directory. `ReadDir` then fails with something other than not-exist, so assert `mmcblk0.Mounted == true`.
- In `safety_test.go`, `TestNANDIsNotAWriteTarget`: a valid NAND device gets `ErrUnsafeDevice` from `ValidateWriteTarget`, and `Validate()` returns nil.

- [ ] **Step 2: Run them and confirm they fail**

Run: `go test ./internal/infrastructure/linux ./internal/domain/device`
Expected: FAIL. `KindNAND` is undefined and the pseudo-devices are listed.

- [ ] **Step 3: Implement**

- `List` skip rule: drop names with prefix `loop`, `ram`, or `zram`, names matching `boot[0-9]+$` or ending in `rpmb`, and devices whose `size` is 0. Read the size inside `readDevice`; return a sentinel so `List` can skip.
- `classify`: names with prefix `rknand` → `KindNAND`.
- Run `blkid` via `exec.CommandContext` with `context.WithTimeout(ctx, timeout)`.
- In `isBusy`, an `os.ReadDir` error other than `os.IsNotExist` on a holders path → busy.

- [ ] **Step 4: Run the tests**

Run: `go test ./internal/... && go vet ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/linux internal/domain/device
git commit -m "fix(linux): list real disks once per scan and bound blkid"
```

---

### Task 3: One device-line formatter, i18n cleanup, and `tui`/`--console` dispatched inside `run`

**Files:**
- Create: `internal/presentation/i18n/format.go`, `internal/presentation/i18n/format_test.go`
- Modify: `internal/presentation/i18n/catalog.go`, `internal/presentation/tui/app.go:72-88`, `internal/presentation/console/app.go`, `cmd/multitool/main.go`
- Test: `cmd/multitool/main_test.go`, `internal/presentation/console/app_test.go`, `internal/presentation/i18n/i18n_test.go`

**Interfaces:**
- Produces in `i18n`:
  - `func (c Catalog) DeviceLine(d device.Device) string`. Format: `"<path>  <kind>  <size>"`, plus `"  · " + c.InUseLabel` when `d.Mounted`, and `"  · " + c.BootMediaLabel` when `d.BootMedia`.
  - `func (c Catalog) Size(bytes int64) string`. Uses decimal GB with one decimal place and `c.DecimalSeparator`. Values under 1 GB are shown in MB, with no decimal places.
- Catalog field changes:
  - Remove `ConsoleTitle`, `ConsoleHardwareOption`, `ConsoleExitOption`, `VersionLabel` and `MountedLabel`.
  - Add `InUseLabel` (`"em uso"` / `"in use"`), `BootMediaLabel` (`"mídia de boot"` / `"boot media"`) and `DecimalSeparator` (`","` / `"."`).
  - Add `Kinds map[device.Kind]string`: pt-BR `emmc→"eMMC", sd→"SD", usb→"USB", nand→"NAND", unknown→"desconhecido"`; en uses the same values except `unknown→"unknown"`.
- Console uses `ProductName`, `Menu` labels by ID (`MenuHardware`, `MenuExit`) and `DeviceLine`. The version output uses `ProductName`.
- Produces: `func run(args []string, root string, devices application.DeviceCatalog, messages i18n.Catalog, stdin io.Reader, stdout, stderr io.Writer, startTUI func() error) int`. `["tui"]` calls `startTUI` and returns 1 on error. `["--console"]` calls `console.Run(stdin, stdout, ...)`. Both share one `discovery.Service`. `main` passes `os.Stdin` and a closure that builds and runs the TUI (Task 9 replaces that closure).

- [ ] **Step 1: Write the failing tests**
  - `TestDeviceLineLocalizesAndLabels`: for pt-BR, `{Path:"/dev/mmcblk0", Kind:KindSD, SizeBytes:14_800_000_000, Mounted:true, BootMedia:true}` → `"/dev/mmcblk0  SD  14,8 GB  · em uso  · mídia de boot"`. For en, the same device without labels → `"/dev/mmcblk0  SD  14.8 GB"`.
  - `TestSizeUnderOneGB`: `Size(4_194_304)` → `"4 MB"`.
  - `TestRunDispatchesTUIAndConsole`: `run([]string{"tui"}, ...)` calls the fake `startTUI` once and returns 0. `run([]string{"--console"}, ...)` with stdin `"0\n"` writes `ProductName` to stdout and returns 0.
  - Update the existing console and TUI hardware tests to expect `DeviceLine` output.

- [ ] **Step 2: Run them and confirm they fail**

Run: `go test ./internal/presentation/... ./cmd/multitool`
Expected: compile failures on `DeviceLine`, and on the new `run` arity.

- [ ] **Step 3: Implement** as specified in Interfaces. Delete the two old per-screen formatters.

- [ ] **Step 4: Run the tests**

Run: `go test ./... && go vet ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/presentation cmd/multitool
git commit -m "refactor(presentation): share one localized device line"
```

---

### Task 4: Image formats: `.img.gz`, exact size, and listing the images folder

**Files:**
- Modify: `internal/domain/image/image.go`, `internal/domain/installation/plan.go`, `internal/infrastructure/compression/reader.go`
- Create: `internal/infrastructure/compression/catalog.go`
- Create: `tests/fixtures/images/sample.img.gz` (gzip of `sample.img`)
- Test: `internal/domain/image/image_test.go`, `internal/infrastructure/compression/reader_test.go`, `internal/infrastructure/compression/catalog_test.go`

**Interfaces:**
- Produces: `image.FormatGzip Format = "gz"`. `FormatFromPath` accepts `.img.gz` case-insensitively and still rejects a bare `.gz`. `installation.Plan.Validate` accepts `FormatGzip`.
- Produces: `compression.Open` returns `gzip.NewReader` over the file for `FormatGzip`, and its `Close` closes both.
- Produces: `func Size(ctx context.Context, path string) (int64, error)`:
  - `.img`: `os.Stat` size.
  - `.img.xz`: parse `xz --robot --list -- <path>`. Take the line starting with `totals\t`; column 5 is the uncompressed size.
  - `.img.gz`: returns `0, nil`, meaning unknown. The burn flow calls `Inspect` (Task 9).
- Produces: `func ListImages(ctx context.Context, dir string) ([]image.Image, error)`. Regular files only (no symlinks), whose names pass `FormatFromPath`, sorted by name. `SizeBytes` comes from `Size`. A missing `dir` returns an empty list and nil.

- [ ] **Step 1: Write the failing tests**
  - `TestFormatFromPathGzip`: `"A.IMG.GZ"` → `FormatGzip`; `"a.gz"` → `ErrUnsupportedFormat`.
  - `TestOpenGzipStreamsImage`: reading the fixture `.img.gz` gives the same bytes as `sample.img`.
  - `TestSizeXZReadsIndex`: the `sample.img.xz` size equals the size of `sample.img` (20).
  - `TestListImagesHandlesSpacesAndSkipsOthers`: a temp dir containing `Armbian (1) d'x.img.xz` (a copy of the fixture), `b.img`, `notes.txt`, and a symlink `c.img`. Expect exactly two entries, in this order: `Armbian (1) d'x.img.xz` with size 20, then `b.img`.
- [ ] **Step 2: Run** `go test ./internal/domain/image ./internal/infrastructure/compression`. Expected: FAIL.
- [ ] **Step 3: Implement** as specified in Interfaces.
- [ ] **Step 4: Run** `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5: Commit**

```bash
git add internal/domain internal/infrastructure/compression tests/fixtures/images/sample.img.gz
git commit -m "feat(compression): support gzip images and list the images folder"
```

---

### Task 5: DDR command rate (rk322x)

**Files:**
- Create: `internal/platform/rk322x/commandrate.go`, `internal/platform/rk322x/commandrate_test.go`

**Interfaces:**
- Produces:
  - `const CommandRateScan = 128 * 1024`
  - `var ErrNoCommandRate = errors.New("command rate not found")`
  - `func FindCommandRate(head []byte) (offset int64, rate byte, err error)`: scan `head[:min(len, CommandRateScan)]` at offsets divisible by 4 for bytes `65 65 86 a7`. Take the first match `S`; the offset is `S-16`. `S < 16`, no match, or `head[S-16]` not in {0, 1} → `ErrNoCommandRate`.

- [ ] **Step 1: Write the failing tests** (table-driven on 128 KiB slices)
  - The signature at 4096 with byte 1 at 4080 → `(4080, 1, nil)`.
  - The signature at 4098 (not 4-aligned) → `ErrNoCommandRate`.
  - No signature → `ErrNoCommandRate`.
  - The signature at 4096 with byte 7 at 4080 → `ErrNoCommandRate`.
  - The signature only at offset 128*1024 (just past the scan window, in a longer slice) → `ErrNoCommandRate`.
- [ ] **Step 2: Run** `go test ./internal/platform/rk322x`. Expected: FAIL (undefined).
- [ ] **Step 3: Implement** `FindCommandRate`.
- [ ] **Step 4: Run** `go test ./internal/platform/rk322x`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(rk322x): locate the DDR command rate byte`

---

### Task 6: Gzip backup with no-space handling

**Files:**
- Modify: `internal/application/backup/service.go`, `internal/application/stream.go`
- Test: `internal/application/backup/service_test.go`, `internal/application/stream_test.go`, `tests/integration/*` callers of `Create`

**Interfaces:**
- Produces: `func (Service) Create(ctx context.Context, source device.Device, src io.Reader, destination string) (backup.Manifest, error)`. The `availableBytes` parameter and `ErrInsufficientSpace` are removed.
  - `destination` is the final `.img.gz` path; the manifest goes to `destination + ".json"`.
  - The data is gzip-compressed (`gzip.NewWriterLevel(tmp, gzip.BestSpeed)`). The manifest `SHA256`/`SizeBytes` cover the uncompressed device data.
  - Any write error that `errors.Is(err, syscall.ENOSPC)` is returned wrapped as `ErrNoSpace = errors.New("no space left on backup partition")`.
  - A folder fsync error matching `syscall.EINVAL` or `syscall.ENOTSUP` is ignored.
  - Existing outputs are still refused before anything is written (current behavior).
- Produces: `CopyExact` maps `io.EOF` to `io.ErrUnexpectedEOF` only when `n < size` (PR #10 deferred minor).

- [ ] **Step 1: Write the failing tests**
  - `TestCreateWritesGzipReadableByGunzip`: back up 4096 bytes. `gzip.NewReader` on the output yields the same bytes, and the manifest SHA equals `sha256` of the raw input.
  - `TestCreateNoSpaceRemovesPartialFiles`: back up through a writer wrapper that fails with `syscall.ENOSPC` after 1024 bytes. To make that possible, add an unexported hook `var dataSink = func(f *os.File) io.Writer { return f }`, which the gzip writer writes through. The test replaces it with a wrapper that fails. Assert `errors.Is(err, ErrNoSpace)`, and that neither `destination`, its `.json`, nor any `.backup-*` temp file exists.
  - `TestCopyExactKeepsWriterEOF`: a `dst` whose `Write` returns `io.EOF` → the error `errors.Is(err, io.EOF)` and not `io.ErrUnexpectedEOF`.
- [ ] **Step 2: Run** `go test ./internal/application/...`. Expected: FAIL.
- [ ] **Step 3: Implement** as specified in Interfaces. Update every `Create` caller.
- [ ] **Step 4: Run** `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(backup): write gzip backups and report a full partition`

---

### Task 7: MULTITOOL workspace and the single-operation lock

**Files:**
- Create: `internal/infrastructure/multitool/workspace.go`, `internal/infrastructure/multitool/workspace_test.go`

**Interfaces:**
- Consumes: `filesystem.CommandRunner` (`internal/infrastructure/filesystem/image_mount.go:16`).
- Produces:

```go
type Locator struct {
    Runner    filesystem.CommandRunner // nil → exec
    ProcRoot  string                   // "/proc"
    MountPoint string                  // "/mnt"
}
type Workspace struct {
    Root     string // mounted MULTITOOL directory
    Platform string // trimmed contents of Root/TARGET, "" if absent
}
func (l Locator) Open(ctx context.Context) (Workspace, func() error, error)
func (w Workspace) ImagesDir() string  // Root/images
func (w Workspace) BackupsDir() string // Root/backups, created 0755 by Open
func (w Workspace) FreeBytes() (int64, error) // syscall.Statfs Bavail*Bsize
var ErrBusy = errors.New("another operation is running")
func Lock(path string) (unlock func() error, err error) // flock LOCK_EX|LOCK_NB; EWOULDBLOCK → ErrBusy
```

`Open` behavior:
1. Run `blkid -l -o device --label MULTITOOL`. Empty output → `errors.New("MULTITOOL partition not found")`.
2. If that device is the first field of a line in `ProcRoot/mounts`, use the second field as `Root`; the close func does nothing.
3. Otherwise run `mount <dev> <MountPoint>`; the close func runs `umount <MountPoint>`.

- [ ] **Step 1: Write the failing tests** (use a fake runner that records calls)
  - `TestOpenUsesExistingMount`: mounts contain `/dev/mmcblk0p1 /media/mt ntfs ...`. Expect `Root == "/media/mt"`, no `mount` call, and a close func that makes no `umount` call.
  - `TestOpenMountsAndUnmounts`: no mounts line. The calls are `blkid`, then `mount /dev/mmcblk0p1 <tmp>`, and close calls `umount <tmp>`. `Platform` reads `rk322x\n` from `<tmp>/TARGET` → `"rk322x"`.
  - `TestOpenMissingPartition`: blkid prints nothing → error, and no `mount` call.
  - `TestLockSecondCallerGetsBusy`: two `Lock` calls on the same temp path. The second returns `ErrBusy`; after the first unlocks, a third succeeds.
- [ ] **Step 2: Run** `go test ./internal/infrastructure/multitool`. Expected: FAIL.
- [ ] **Step 3: Implement** as specified in Interfaces.
- [ ] **Step 4: Run** `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(multitool): locate the MULTITOOL partition and serialize operations`

---

### Task 8: Burn use case (write, verify, command rate)

**Files:**
- Create: `internal/application/installer/burn.go`, `internal/application/installer/burn_test.go`

**Interfaces:**
- Consumes: `Writer.Execute` (current signature), `rk322x.FindCommandRate`, `rk322x.CommandRateScan`, `application.Target`.
- Produces:

```go
type Burn struct {
    Guard      storageguard.Guard
    OpenTarget func(string) (application.Target, error)
    Platform   string // "rk322x" enables command-rate preservation
}
type BurnResult struct{ CommandRateNotApplied bool }
func (b Burn) Run(ctx context.Context, plan installation.Plan, source io.Reader) (BurnResult, error)
```

`Run` steps:
1. `plan.Validate()`.
2. If `Platform == "rk322x"`: open the target, read the first `CommandRateScan` bytes (`io.ReadFull`; `io.ErrUnexpectedEOF` is fine), run `FindCommandRate`, close. A not-found result means `saved = none`.
3. `Writer{}.Execute(ctx, plan, source, true, b.Guard, b.OpenTarget)`.
4. If something was saved: open the target, read the head again, and run `FindCommandRate`. Not found or invalid → `BurnResult{CommandRateNotApplied: true}`, nil. Found → seek to the offset, write the saved byte, `Sync`, close. A write, sync or close error → return it.

- [ ] **Step 1: Write the failing tests** (targets are temp files opened with `os.OpenFile`; the catalog fake returns the device)
  - `TestBurnRestoresCommandRate`: the target has the signature at 4096 with rate byte 1, and the image has the signature at 8192 with rate byte 0. After `Run`, the target byte at 8176 is 1, the rest equals the image, and the result is `{false}`.
  - `TestBurnWarnsWhenImageHasNoCommandRate`: the target has a rate and the image has no signature → nil error and `{true}`.
  - `TestBurnSkipsCommandRateOffRK322x`: `Platform: ""`, both have signatures → the target equals the image byte-for-byte.
  - `TestBurnRejectsImageLargerThanTarget`: the plan image is 1024 bytes and the target is 512 → `installation.ErrImageTooLarge`, and `OpenTarget` is never called.
- [ ] **Step 2: Run** `go test ./internal/application/installer`. Expected: FAIL.
- [ ] **Step 3: Implement** `Run`.
- [ ] **Step 4: Run** `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(installer): burn images and preserve the rk322x command rate`

---

### Task 9: TUI shell: new menu, async inventory, Bash item, wiring

**Files:**
- Modify: `internal/presentation/tui/app.go`, `internal/presentation/tui/screens/home.go`, `internal/presentation/i18n/catalog.go`
- Create: `cmd/multitool/wire.go`
- Modify: `cmd/multitool/main.go`
- Test: `internal/presentation/tui/app_test.go`, `internal/presentation/tui/screens/home_test.go`, `internal/presentation/i18n/i18n_test.go`

**Interfaces:**
- Produces in `tui`:

```go
type Progress func(done, total int64) // total 0 = unknown
type Services struct {
    Inventory  func(context.Context) ([]device.Device, error)
    Images     func(context.Context) ([]image.Image, error)
    FreeBytes  func(context.Context) (int64, error)
    Burn       func(ctx context.Context, img image.Image, target device.Device, p Progress) (installer.BurnResult, error)
    Backup     func(ctx context.Context, source device.Device, name string, p Progress) error
    RunLegacy  func() error
}
func New(s Services, c i18n.Catalog) *tview.Application
```

- Menu IDs: `MenuBurn`, `MenuBackup`, `MenuHardware`, `MenuLegacy`, `MenuExit`, in that order. Labels and hints:
  - pt-BR: "Gravar imagem no eMMC" / "Instalar Armbian ou outra imagem"; "Fazer backup" / "Salvar o eMMC na partição MULTITOOL"; "Informações de hardware" / "Ver dispositivos de armazenamento conectados"; "Menu antigo (Bash)" / "Restaurar, apagar, steP-nand"; "Sair" / "Fechar o GoMultitool".
  - en: "Burn image to eMMC" / "Install Armbian or another image"; "Back up" / "Save the eMMC to the MULTITOOL partition"; "Hardware information" / "Inspect connected storage devices"; "Legacy menu (Bash)" / "Restore, erase, steP-nand"; "Exit" / "Leave GoMultitool".
  - `Menu` becomes `[5]MenuItem`. The old future IDs are deleted.
- Every service call runs in a goroutine and updates widgets through `app.QueueUpdateDraw`. While it runs, a "Carregando…" / "Loading…" text is shown. `q`/Ctrl+C stay responsive.
- `MenuLegacy` calls `app.Suspend(func(){ err = s.RunLegacy() })`; an error is shown in a modal.
- `wire.go`: `func services(catalog linuxsys.SysfsCatalog, locator multitool.Locator) tui.Services`.
  - `Burn` and `Backup` take `multitool.Lock("/run/multitool.lock")` first (`ErrBusy` → the localized "Outra operação em andamento." / "Another operation is running.").
  - They open the workspace, build the `installation.Plan`, and fill a `.img.gz` size of 0 with `compression.Inspect`.
  - They wrap the source in a counting reader that calls `Progress`, run `installer.Burn` / `backup.Service.Create` with `storage` from `os.OpenFile(path, os.O_RDWR, 0)`, and then close the workspace.
  - `RunLegacy` runs `exec.Command("/usr/local/bin/multitool.sh")` with the standard streams.

- [ ] **Step 1: Write the failing tests**
  - `TestMenuShowsFiveItems` (pt-BR and en): labels and hints exactly as listed above.
  - `TestHardwareLoadsAsyncAndStaysResponsive`: an `Inventory` fake that blocks on a channel. Press Enter on Hardware; the screen shows "Carregando…"; then `q` stops the app before the channel is released.
  - `TestLegacyItemRunsBash`: `RunLegacy` fake counts calls. Select `MenuLegacy` → count is 1 and the menu is drawn again afterwards.
- [ ] **Step 2: Run** `go test ./internal/presentation/... ./cmd/multitool`. Expected: FAIL.
- [ ] **Step 3: Implement** the shell, menu and wiring. Burn and Backup menu entries open placeholder primitives that Tasks 10 and 11 replace.
- [ ] **Step 4: Run** `go test ./... && go vet ./... && make build-armv7`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(tui): boot menu with async inventory and legacy menu entry`

---

### Task 10: Burn screen (single screen)

**Files:**
- Create: `internal/presentation/tui/burn.go`, `internal/presentation/tui/burn_test.go`
- Modify: `internal/presentation/i18n/catalog.go`

**Interfaces:**
- Consumes: `tui.Services.Inventory`, `.Images`, `.Burn`, and `Catalog.DeviceLine`/`Size` (Tasks 3 and 9).
- Screen layout (mockup B):
  - Destination list. Devices with `BootMedia` are omitted. Devices with `Mounted` or `Kind == KindNAND` or `Kind == KindUnknown` are shown via `DeviceLine` but can't be selected.
  - Image list: name plus `Size`, or "tamanho calculado ao gravar" / "size computed when burning" for 0.
  - The warning "Tudo em <path> será apagado." / "Everything on <path> will be erased."
  - Buttons Gravar/Cancelar (Burn/Cancel).
- While burning:
  - Progress text `"<pct>%  <done> / <total>"` using `Size`, plus a bar.
  - Inputs are disabled.
  - Esc does nothing.
- Results:
  - Success: "Gravado e conferido. Pode desligar." / "Burned and verified. You can power off." If `CommandRateNotApplied`, add "A imagem não tem o command rate da DDR; o valor anterior não foi aplicado." / "The image has no DDR command rate; the previous value was not applied."
  - Error: the catalog `ErrorPrefix` plus the error. Never show the success line.
  - Empty states: "Nenhuma imagem em MULTITOOL/images." / "No images in MULTITOOL/images." and "Nenhum destino disponível." / "No available destination."

- [ ] **Step 1: Write the failing tests** (simulation screen, fakes)
  - `TestBurnScreenHidesBootMediaAndBlocksBusy`: the inventory has a boot SD, a busy USB and a free eMMC. The screen has no line for the SD; Enter on the USB does not select it.
  - `TestBurnScreenShowsSuccessOnlyOnSuccess`: the `Burn` fake returns `{}` and nil → success text. The `Burn` fake returns an error → the error text and no success text.
  - `TestBurnScreenWarnsOnCommandRate`: the `Burn` fake returns `{CommandRateNotApplied: true}` → both lines are shown.
  - `TestBurnScreenEmptyImages`: `Images` returns an empty list → the empty-state text, and the Gravar button is disabled.
- [ ] **Step 2: Run** `go test ./internal/presentation/tui`. Expected: FAIL.
- [ ] **Step 3: Implement** using tview `List`s plus a `Flex`.
- [ ] **Step 4: Run** `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(tui): single-screen image burn`

---

### Task 11: Backup screen (single screen)

**Files:**
- Create: `internal/presentation/tui/backup.go`, `internal/presentation/tui/backup_test.go`
- Modify: `internal/presentation/i18n/catalog.go`

**Interfaces:**
- Consumes: `tui.Services.Inventory`, `.FreeBytes`, `.Backup`.
- Layout:
  - Source list, with the same filter as burn: boot media omitted; busy, NAND and unknown devices shown but not selectable.
  - A name input, defaulting to `tvbox-backup-` + `time.Now().Format("2006-01-02")`, with the suffix `.img.gz` shown after it.
  - The line "Livre na MULTITOOL: <Size>" / "Free on MULTITOOL: <Size>".
  - Buttons "Fazer backup"/"Cancelar" ("Back up"/"Cancel").
- Name validation runs before calling `Backup`: empty, contains `/`, or starts with `.` → "Nome inválido." / "Invalid name."
- Results:
  - Success: "Backup salvo: <name>.img.gz" / "Backup saved: <name>.img.gz".
  - `backup.ErrNoSpace` → "Sem espaço na partição MULTITOOL." / "No space left on the MULTITOOL partition."
  - An existing file → "Já existe um backup com esse nome." / "A backup with this name already exists." (`Create` returns an error containing "already exists"; add `var ErrExists` in `backup` and wrap it.)
  - Any other error → `ErrorPrefix` plus the error.

- [ ] **Step 1: Write the failing tests**
  - `TestBackupScreenDefaultsNameWithDate`: the input text starts with `tvbox-backup-` and is followed by today's date.
  - `TestBackupScreenRejectsBadNames`: `""`, `"a/b"`, `".x"` → "Nome inválido." each time, and the `Backup` fake is never called.
  - `TestBackupScreenNoSpaceMessage`: the fake returns `backup.ErrNoSpace` → the no-space text, and no "Backup salvo".
  - `TestBackupScreenNameWithSpaces`: the name `"tv box"` reaches the fake unchanged.
- [ ] **Step 2: Run** `go test ./internal/presentation/tui ./internal/application/backup`. Expected: FAIL.
- [ ] **Step 3: Implement**, including `backup.ErrExists`.
- [ ] **Step 4: Run** `go test ./... && go vet ./... && go test -race ./...`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(tui): single-screen backup`

---

### Task 12: Boot wiring behind the opt-in flag

**Files:**
- Modify: `build/rootfs/multitool-init`, `tests/integration/rootfs_test.sh`, `docs/architecture/legacy-runtime.md`

- [ ] **Step 1: Update the test** so `rootfs_test.sh` asserts `multitool-init` contains `exec /usr/local/bin/multitool tui` and still contains the `multitool.sh` fallback.
- [ ] **Step 2: Run** `bash tests/integration/rootfs_test.sh`. Expected: FAIL.
- [ ] **Step 3: Change** `--console` to `tui` in `multitool-init`. In `legacy-runtime.md`, update the sentence about `MULTITOOL_GO_RUNTIME=1` to say it starts the Go TUI.
- [ ] **Step 4: Run** `bash tests/integration/rootfs_test.sh && make lint-docs`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(rootfs): start the Go TUI when the Go runtime is enabled`

---

### Task 13: Make the Go TUI the boot screen (only after hardware approval)

**Do not execute until the owner reports that backup and burn worked on a real box.** The executor stops after Task 12 and asks.

**Files:**
- Modify: `sources/multistrap/init:118-122`, `build/rootfs/multitool-init`, `tests/integration/rootfs_test.sh`

- [ ] **Step 1: Update the test**: `sources/multistrap/init` launches `/usr/local/sbin/multitool-init` with `MULTITOOL_GO_RUNTIME=1`.
- [ ] **Step 2: Run it and confirm it fails.**
- [ ] **Step 3: Change** the `init` loop to `setsid env MULTITOOL_GO_RUNTIME=1 /usr/local/sbin/multitool-init <> /dev/tty1 >&0 2>&1`, keeping the loop and the `-f /usr/local/bin/multitool.sh` guard.
- [ ] **Step 4: Run** `bash tests/integration/rootfs_test.sh`. Expected: PASS.
- [ ] **Step 5: Commit** `feat(rootfs): boot into the Go TUI`
