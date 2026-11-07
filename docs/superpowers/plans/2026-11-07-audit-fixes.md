# Audit Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the 11 findings of the 2026-11-07 repository audit: two broken `diagnose --json` results, an incomplete busy-device check, wrong file ownership in Armbian presets, duplicated copy loops, dead code, a missing directory sync after backup, and a redundant installer parameter.

**Architecture:** Small, independent changes inside the existing layers. One new shared helper (`internal/application/stream.go`) replaces four copy loops; nothing else adds files. Deletions outnumber additions.

**Tech Stack:** Go 1.27.3, standard library only. No new dependencies.

**Spec:** The audit report in the conversation that produced this plan. Finding numbers (#1–#11) are quoted in each task.

## Global Constraints

- No new modules in `go.mod`.
- `make check` (or at least `go test ./...` and `go vet ./...`) passes after every task.
- Keep the current fail-safe rule in `internal/infrastructure/linux`: when state cannot be read, treat the device as busy (`Mounted = true`) or as boot media.
- Short source data keeps returning an error that matches `errors.Is(err, io.ErrUnexpectedEOF)`. Existing tests in `backup/service_test.go:134` and `installer/writer_test.go:150` rely on it.
- Commit messages follow Conventional Commits (`fix(scope): ...`, `refactor(scope): ...`), as enforced by `.githooks/commit-msg`.
- Work on a branch, not on `main`.

## Review Focus

- `/proc/swaps` missing (kernel built without swap): `List` still succeeds and no device is marked busy only because of that. Test in Task 2.
- `/proc/swaps` lists a partition (`/dev/mmcblk1p3`): its parent disk `mmcblk1` is marked busy. Test in Task 2.
- `/proc/swaps` lists a swap file (`/swapfile`): no block device is marked busy. Test in Task 2.
- `diagnose --json` on a system whose `/var/log/syslog` cannot be read: exit code 0 with a full report. Test in Task 1.
- Context cancelled during a copy: the copy stops and returns `context.Canceled`. Test in Task 4.

---

### Task 1: Fix `diagnose --json` (findings #1, #2, #7)

**Files:**
- Modify: `cmd/multitool/main.go:46-98`
- Modify: `internal/application/diagnostics/boot.go:10-17`
- Modify: `internal/application/diagnostics/report.go:46-50`
- Delete: `internal/infrastructure/linux/logs.go`, `internal/infrastructure/linux/logs_test.go`
- Test: `cmd/multitool/main_test.go`, `internal/application/diagnostics/report_test.go`

**Interfaces:**
- Produces: `func run(args []string, root string, devices application.DeviceCatalog, messages i18n.Catalog, stdout, stderr io.Writer) int`. `main` passes `"/"`. The 7 existing `run(...)` calls in `main_test.go` pass `t.TempDir()` or `"/"` (they don't touch `diagnose`).
- Produces: `BootReport` JSON keys `extlinux_config_present`, `kernel_present`, `observations`, `missing`, `hypotheses`, `journal_available`.

- [ ] **Step 1: Write the failing tests**

In `report_test.go`, add `TestCollectBootReportCopiesBootAssetFlags`: temp root with `boot/zImage` and `boot/extlinux/extlinux.conf`; assert `r.KernelPresent && r.ExtlinuxConfigPresent`.

In `main_test.go`, add `TestDiagnoseJSON`: temp root containing `proc/cmdline` (`root=/dev/mmcblk0p2\n`), `boot/zImage`, and `var/log/syslog` written with mode `0o000`. Call `run([]string{"diagnose", "--json"}, root, testCatalog{}, i18n.For(i18n.LocalePtBR), &out, &stderr)`. Assert:

```go
if code != 0 { t.Fatalf("code=%d stderr=%s", code, stderr.String()) }
var got map[string]json.RawMessage
// decode out into got
// got has keys "boot" and "network" and no key "journal_available"
var boot struct{ KernelPresent bool `json:"kernel_present"` }
// decode got["boot"] into boot; boot.KernelPresent must be true
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `go test ./cmd/multitool ./internal/application/diagnostics`
Expected: compile error on the new `run` argument, and `KernelPresent` false.

- [ ] **Step 3: Implement**

- `main.go`: add the `root` parameter; read `filepath.Join(root, "proc/cmdline")`; call `diagnostics.CollectBootReport(root, ...)`; delete the `ReadJournal` call and the top-level `JournalAvailable` field. The report struct keeps only `Boot` and `Network`.
- `boot.go`: add `json` tags to `ExtlinuxConfigPresent`, `KernelPresent`, `Observations`, `Missing`, using the keys listed above.
- `report.go`: after `boot := InspectBootAssets(...)`, copy `boot.KernelPresent` and `boot.ExtlinuxConfigPresent` into `r`.
- Delete `logs.go` and `logs_test.go`. `ReadJournal` and `RedactLog` have no other callers.

- [ ] **Step 4: Run the tests and confirm they pass**

Run: `go test ./... && go vet ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add -A cmd/multitool internal/application/diagnostics internal/infrastructure/linux
git commit -m "fix(diagnostics): report boot assets and stop failing on unreadable syslog"
```

---

### Task 2: Treat active swap and held devices as busy (findings #3, #11)

**Files:**
- Modify: `internal/infrastructure/linux/procfs.go:13-25`
- Modify: `internal/infrastructure/linux/sysfs.go:43, 83`
- Test: `internal/infrastructure/linux/sysfs_test.go`

**Interfaces:**
- Produces: `func (c SysfsCatalog) isBusy(name, path string) bool`, which replaces `isMounted`. `readDevice` sets `d.Mounted = c.isBusy(name, d.Path)`. `Device.Mounted` keeps its name; its meaning widens to "in use".

- [ ] **Step 1: Write the failing tests**

All tests use `copyFixtureBlockDevices` and a temp `proc` with an empty `mounts` and a `cmdline` of `root=/dev/sda1`.

- `TestSysfsCatalogMarksActiveSwapPartitionBusy`: `proc/swaps` = header line plus `/dev/mmcblk1p3 partition 524284 0 -2`. Assert `mmcblk1.Mounted == true` and `mmcblk0.Mounted == false`.
- `TestSysfsCatalogIgnoresSwapFiles`: `proc/swaps` = header plus `/swapfile file 1048572 0 -2`. Assert no device has `Mounted == true`.
- `TestSysfsCatalogToleratesMissingSwaps`: no `proc/swaps`. Assert `List` returns no error and `mmcblk0.Mounted == false`.
- `TestSysfsCatalogMarksHeldPartitionBusy`: create the empty directory `sys/block/mmcblk0/mmcblk0p1/holders/dm-0`. Assert `mmcblk0.Mounted == true`.

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `go test ./internal/infrastructure/linux -run 'Swap|Held'`
Expected: FAIL on the swap-partition and holder tests.

- [ ] **Step 3: Implement `isBusy`**

Return true when any of these holds:
1. Current `/proc/mounts` logic matches, keeping "unreadable means busy".
2. The first field of any line after the header in `filepath.Join(c.ProcRoot, "swaps")` matches `sameDeviceOrPartition(field, path)`. If the file does not exist, skip this check; any other read error means busy.
3. `filepath.Join(c.SysRoot, "block", name, "holders")` or any `filepath.Glob(filepath.Join(c.SysRoot, "block", name, name+"*", "holders"))` directory has at least one entry.

Also delete the `sort.Slice` in `List` (`sysfs.go:43`). `os.ReadDir` already returns entries sorted by name, which is the ID, and `discovery.Service.Inventory` sorts again for any catalog.

- [ ] **Step 4: Run the tests and confirm they pass**

Run: `go test ./internal/infrastructure/linux ./internal/application/...`
Expected: PASS, including the existing `TestSysfsCatalogListsKnownDevices`.

- [ ] **Step 5: Commit**

```bash
git add internal/infrastructure/linux
git commit -m "fix(linux): treat active swap and held devices as busy"
```

---

### Task 3: Fix Armbian preset file ownership (finding #4, plus `Artifacts.FirstBoot` from #6)

**Files:**
- Modify: `internal/application/imageconfig/armbian.go:27-31, 70-80`
- Test: `internal/application/imageconfig/armbian_test.go:40-46`, `tests/integration/image_provision_test.go` (only if it references the old key path)

**Interfaces:**
- Produces: `type Artifacts struct { Files map[string]File }`, with the `FirstBoot` field removed.
- Produces: the SSH key goes to `etc/skel/.ssh/authorized_keys` with mode `0o600`. `etc/hostname` gets mode `0o644`.

Why `etc/skel`: Armbian's first login creates the user with `adduser`, which copies `/etc/skel` into the new home and gives it to that user. The home folder is then created by `adduser` with the right owner, and the key ends up owned by the user. Add above the entry: `// shortcut: every user created later also gets this key; switch to a first-boot chown if multi-user images matter.`

- [ ] **Step 1: Update the tests to the new contract**

In `armbian_test.go`, replace `result.FirstBoot` with `result.Files["root/.not_logged_in_yet"].Data`. Assert:

```go
result.Files["etc/skel/.ssh/authorized_keys"].Mode == 0o600
result.Files["etc/hostname"].Mode == 0o644
_, ok := result.Files["home/armbian/.ssh/authorized_keys"]; !ok
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `go test ./internal/application/imageconfig`
Expected: compile error on `FirstBoot`, or failures on the path and mode checks.

- [ ] **Step 3: Implement**

Change the two map entries and the `Artifacts` struct as listed in Interfaces. `writeArtifact` needs no change, because `MkdirAll` leaves the existing `etc/skel` alone and only creates `.ssh` with `0o700`.

- [ ] **Step 4: Run the tests and confirm they pass**

Run: `go test ./internal/application/imageconfig ./tests/integration`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/application/imageconfig tests/integration
git commit -m "fix(imageconfig): let first boot own the user home and keep hostname readable"
```

---

### Task 4: One shared copy helper and one `Target` type (findings #5, #10, and `BlockOpener`/`BlockWriter` from #6)

**Files:**
- Create: `internal/application/stream.go`
- Create: `internal/application/stream_test.go`
- Modify: `internal/application/ports.go:17-28`
- Modify: `internal/application/installer/writer.go:17-22, 64-80, 90-143`
- Modify: `internal/application/restore/service.go:23-28, 61, 89, 98, 108-143`
- Modify: `internal/application/backup/service.go:54-84`
- Modify: `internal/application/imageconfig/service.go:82, 194-217`
- Modify: `internal/infrastructure/compression/reader.go:50-62`
- Test: `tests/integration/install_test.go:66`, `tests/integration/restore_test.go:81`, and the unit tests of installer and restore that name `Target`

**Interfaces:**
- Produces, in `package application`:
  - `type Target interface { io.ReadWriteSeeker; Sync() error; Close() error }`, which replaces `BlockOpener` and `BlockWriter` (both deleted) and both `installer.Target` and `restore.Target` (deleted).
  - `func ContextReader(ctx context.Context, r io.Reader) io.Reader`: returns `ctx.Err()` before each `Read`.
  - `func CopyExact(ctx context.Context, dst io.Writer, src io.Reader, size int64) error`: uses `io.CopyN(dst, ContextReader(ctx, src), size)` and turns `io.EOF` into `io.ErrUnexpectedEOF`.

- [ ] **Step 1: Write the failing tests in `stream_test.go`**

- `TestCopyExactCopiesExactlySize`: src `"abcdef"`, size 4. `dst == "abcd"`, err nil.
- `TestCopyExactShortSource`: src `"ab"`, size 4. `errors.Is(err, io.ErrUnexpectedEOF)`.
- `TestCopyExactStopsOnCancel`: cancelled ctx. `errors.Is(err, context.Canceled)` and `dst` is empty.

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `go test ./internal/application`
Expected: FAIL, `CopyExact` undefined.

- [ ] **Step 3: Implement `stream.go` and update `ports.go`**

- [ ] **Step 4: Switch the callers**

- installer: `h := sha256.New()`, then `application.CopyExact(ctx, io.MultiWriter(target, h), source, size)` for the write and `application.CopyExact(ctx, h2, target, size)` for readback. Delete `copyAndHash` and `hashExact`.
- restore: same pattern. Delete `copyExact` and `hashExact`.
- backup: replace the loop at lines 54-84 with one `CopyExact(ctx, io.MultiWriter(dataTmp, hash), src, source.SizeBytes)`, keeping `_ = dataTmp.Close()` on error.
- imageconfig: replace `copyContext` with `io.Copy(tmp, application.ContextReader(ctx, src))`. Delete `copyContext`.
- compression: replace the loop with `size, err := io.Copy(io.Discard, r)`, keeping the `Close` handling.
- Tests: `installer.Target` and `restore.Target` become `application.Target`.

- [ ] **Step 5: Run the tests and confirm they pass**

Run: `go test ./... && go vet ./...`
Expected: PASS, including `ErrUnexpectedEOF` and `ErrShortWrite` in `writer_test.go`. `io.CopyN` returns `io.ErrShortWrite` itself.

- [ ] **Step 6: Commit**

```bash
git add -A internal tests
git commit -m "refactor(application): share one exact-copy helper and Target type"
```

---

### Task 5: Delete dead code (rest of finding #6)

**Files:**
- Delete: `internal/infrastructure/storage/` (all four files)
- Modify: `internal/domain/backup/status.go` (keep only `StatusComplete`); `internal/domain/backup/manifest_test.go:48` uses `Status("pending")`
- Modify: `internal/domain/device/device.go:14-15, 40` (remove `KindNAND`)
- Modify: `internal/application/imageconfig/memory.go:19, 42` (remove `MemoryArtifact.Priority`); `memory_test.go:31`
- Modify: `internal/application/imageconfig/preset.go:67-70` (remove `Redacted`); `preset_test.go:22`
- Modify: `internal/presentation/console/app.go:83-87` (remove `Confirm`); `app_test.go:85-93`
- Modify: `internal/platform/rk322x/profile.go:81-90` (remove `ValidateStorage`, `MayModifyBootloader`, `ErrUnsupportedStorage`); `profile_test.go:50-55`

The staged packages (installer, backup, restore, storageguard, imageconfig, filesystem, compression, bootloader, rk322x) stay. Removing them is a product decision and is out of scope.

- [ ] **Step 1: Confirm each symbol has no other caller**

Run: `grep -rn -E 'storage\.Open|StatusPending|StatusInterrupted|StatusFailed|KindNAND|\.Priority|Redacted|Confirm\(|ValidateStorage|MayModifyBootloader|ErrUnsupportedStorage' --include='*.go' --include='*.md' .`
Expected: hits only in the files listed above, plus docs. If a doc mentions one of them, update that line in the same commit.

- [ ] **Step 2: Delete the code and the tests that cover only it**

- [ ] **Step 3: Run the tests**

Run: `go test ./... && go vet ./... && make lint`
Expected: PASS, with no "unused" lint warnings.

- [ ] **Step 4: Commit**

```bash
git add -A internal
git commit -m "refactor: remove unused types, options and helpers"
```

---

### Task 6: Sync the backup folder after the final rename (finding #8)

**Files:**
- Modify: `internal/application/backup/service.go:129-133`

No new test. A test can't observe whether `fsync` reached the disk. The existing backup tests cover the success and error paths around it.

- [ ] **Step 1: Implement**

After `os.Rename(manifestTmpName, manifestPath)` succeeds, open `dir`, call `Sync()`, then `Close()`. On error, return it without deleting the files: both renames already happened, and the data is complete.

- [ ] **Step 2: Run the tests**

Run: `go test ./internal/application/backup ./tests/integration`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/application/backup/service.go
git commit -m "fix(backup): sync destination directory after publishing backup"
```

---

### Task 7: Remove the second plan from `Writer.Execute` (finding #9)

**Files:**
- Modify: `internal/application/installer/writer.go:28-47`
- Test: `internal/application/installer/writer_test.go`, `tests/integration/install_test.go:60-67`

**Interfaces:**
- Consumes: `application.Target` (Task 4).
- Produces: `func (Writer) Execute(ctx context.Context, plan installation.Plan, source io.Reader, confirmed bool, guard storageguard.Guard, openTarget func(string) (application.Target, error)) error`

- [ ] **Step 1: Update the tests to the new signature**

Drop the `storageguard.Plan` argument at every call. A test that passed a mismatched `targetPlan` to get `ErrIdentityMismatch` now changes the catalog to return a different device, and expects `storageguard.ErrDeviceChanged`.

- [ ] **Step 2: Run the tests and confirm they fail**

Run: `go test ./internal/application/installer ./tests/integration`
Expected: compile error, too many arguments.

- [ ] **Step 3: Implement**

Build `targetPlan := storageguard.Plan{Device: plan.Target}` inside `Execute`. Delete the `plan.Target != targetPlan.Device` comparison, but keep the nil checks for `guard.Catalog` and `openTarget`.

- [ ] **Step 4: Run all checks**

Run: `make check` (if a tool is missing locally: `go test ./... && go test -race ./... && go vet ./...`)
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/application/installer tests/integration
git commit -m "refactor(installer): derive guard plan from installation plan"
```
