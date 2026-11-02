# RK322x Multitool — Go Modernization Design

**Status:** Proposed  
**Date:** 2026-11-02  
**Type:** Superpowers Design Specification  
**Priority:** High  
**Dependency:** Repository Standards & CI/CD Design

## 1. Overview

Modernize the existing RK322x Multitool into a secure, maintainable, modular Go application for recovering ARM-based devices and installing Linux distributions.

The project will preserve the verified Rockchip boot chain while progressively migrating application-level Bash functionality to Go.

Initial reference hardware:

- MXQ Pro 4K TV Box
- Rockchip RK3229 (ARMv7)
- Approximately 1 GiB RAM
- Approximately 8 GB internal eMMC
- MicroSD and USB interfaces

Original project: https://github.com/paolosabatino/multitool

The original recovery image and source code must remain available throughout migration.

## 2. Goals

1. Migrate menu and application-level Bash logic to Go.
2. Implement a keyboard-driven terminal interface.
3. Use pragmatic DDD, SOLID principles, and hexagonal architecture.
4. Preserve existing verified bootloader and kernel compatibility.
5. Detect eMMC, SD, USB, and supported NAND storage safely.
6. Implement reliable backup and restore operations.
7. Support installation of compatible raw and compressed Linux images.
8. Provide offline configuration of compatible Armbian distributions.
9. Implement boot, filesystem, and network diagnostics.
10. Provide optional ZRAM and swap configuration.
11. Operate offline and without a graphical desktop.
12. Ensure bounded memory usage on hardware with 1 GiB RAM.
13. Provide testable, independently usable application services.

All source code, identifiers, comments, tests, and documentation must be in English. User-facing messages must default to Brazilian Portuguese.

## 3. Non-Goals

The project will not initially:

- Rewrite the Linux kernel or U-Boot in Go.
- Implement a new Rockchip hardware boot chain.
- Replace known-working firmware with unverified binaries.
- Guarantee compatibility with every RK322x board.
- Automatically flash unknown NAND or bootloader layouts.
- Support arbitrary ARM SoCs.
- Require internet access to perform recovery operations.
- Reimplement mature Linux filesystem utilities.
- Provide a graphical desktop.
- Advertise successful image writing as proof of successful hardware boot.

## 4. Technology Stack

- Go with a pinned stable version
- Linux ARMv7, `GOARCH=arm`, `GOARM=7`
- `CGO_ENABLED=0` where feasible
- Bubble Tea and Lip Gloss for TUI
- Minimal console fallback
- Go standard library for I/O, hashing, and streaming
- `log/slog` for structured logging
- Linux sysfs and procfs for discovery
- Validated Linux utilities for filesystem creation and repair
- BusyBox and minimal Linux root filesystem
- Existing verified Rockchip firmware, kernel, U-Boot, and device trees

No Python or Node.js runtime dependency is allowed in the recovery image.

## 5. Architecture

Use a pragmatic hexagonal architecture with four layers.

### Domain

Responsibilities:

- Block-device identity and classification
- Device capacity and geometry
- Image metadata and integrity
- Backup manifests
- Installation safety rules
- Board compatibility
- Operation states and typed errors

Domain packages must not depend on Linux adapters, the TUI, or external tools.

### Application

Responsibilities:

- Discover devices
- Plan installations
- Create and verify backups
- Restore storage
- Install operating system images
- Configure supported Linux images
- Collect diagnostics
- Apply optional server settings

All destructive operations must separate planning from execution.

### Infrastructure

Responsibilities:

- Read Linux sysfs and procfs
- Access block devices
- Mount filesystems and loop devices
- Stream compressed images
- Execute vetted Linux utilities
- Read bootloader metadata
- Collect logs

External processes must use explicit argument arrays and must not interpolate untrusted input into shell commands.

### Presentation

Responsibilities:

- Keyboard-driven TUI
- Hardware information
- Progress reporting
- Explicit confirmations
- Localized error messages
- Cancellation and operation status
- Console fallback

Presentation must depend on application interfaces, not raw devices.

## 6. Proposed Structure

```text
cmd/
  multitool/
    main.go

internal/
  domain/
    device/
    image/
    backup/
    installation/
    boot/

  application/
    discovery/
    backup/
    restore/
    installer/
    imageconfig/
    diagnostics/
    storageguard/

  infrastructure/
    linux/
    storage/
    compression/
    filesystem/
    bootloader/
    process/

  presentation/
    tui/
      screens/
      components/
      styles/
    console/
    i18n/

  platform/
    rk322x/

configs/
  boards/
  installers/

build/
  rootfs/
  kernel/
  uboot/
  image/

scripts/
  build-image.sh

tests/
  integration/
  fixtures/
  hardware/

docs/
  architecture/
  development/
  supported-boards/
  recovery/
  superpowers/
    specs/
    plans/
```

Avoid unnecessary interfaces, factories, or abstractions that are not justified by actual use cases.

## 7. Hardware Detection

The application must discover:

- Processor architecture and SoC information
- Physical memory
- Kernel version
- Internal and removable storage
- eMMC, SD, USB, and NAND where recognizable
- Device capacity and logical sector size
- Mounted and busy filesystems
- Active recovery boot device
- Network interfaces
- Board model, revision, and relevant device-tree information

Never hardcode a device such as `/dev/mmcblk2` as internal storage.

Classification must be based on kernel-provided metadata. Unknown or inconsistent hardware must remain explicitly unknown.

## 8. Storage Safety

Every destructive operation must:

1. Build a read-only execution plan.
2. Identify the source and destination.
3. Validate the destination's physical identity.
4. Reject the active recovery medium.
5. Reject mounted or busy destinations.
6. Reject ambiguous hardware identification.
7. Validate image size and destination capacity.
8. Display the exact operation and affected device.
9. Require explicit confirmation.
10. Revalidate device identity immediately before opening it for writing.
11. Synchronize completed writes.
12. Verify the result.
13. Distinguish completed, interrupted, and failed operations.

Never automatically erase flash to resolve boot failures.

Unknown board profiles must prevent bootloader modification.

## 9. Terminal Interface

Proposed menu:

```text
RK322x Multitool

1. Hardware Information
2. Storage Management
3. Backup & Restore
4. Install Linux
5. Configure Linux Image
6. Boot Diagnostics
7. System Configuration
8. Terminal
9. Reboot / Shutdown
```

The interface must operate with a USB keyboard and standard Linux console.

Required capabilities:

- Keyboard navigation
- Progress indicators
- Cancellation
- Explicit destructive-operation confirmation
- Small-screen rendering
- Brazilian Portuguese messages
- Actionable error reporting

If Bubble Tea is incompatible with the actual Linux console, the core application must remain available through a simple text-based frontend.

## 10. Backup and Restore

### Backup requirements

- Stream block-device data using bounded buffers.
- Verify destination capacity.
- Generate SHA-256 checksums.
- Support optional compression.
- Generate versioned JSON metadata.
- Support cancellation without falsely reporting success.
- Track source identity, capacity, sector size, and tool version.
- Use temporary output files and finalize only after complete verification.

### Restore requirements

- Verify backup integrity before restoration.
- Confirm target device identity and capacity.
- Revalidate the destination at execution time.
- Require an explicit destructive confirmation.
- Synchronize and verify restored data.
- Reject incomplete or invalid backup manifests.
- Record success, interruption, or failure separately.

## 11. Image Installation

Initial supported formats:

- `.img`
- `.img.xz`

Installation pipeline:

1. Identify the selected image.
2. Validate its format.
3. Validate compression and metadata.
4. Determine uncompressed size.
5. Identify and validate the destination.
6. Show an execution preview.
7. Require explicit confirmation.
8. Stream data to the device.
9. Synchronize writes.
10. Perform readback verification.
11. Inspect resulting partitions and filesystems.
12. Report installation results.

A successfully written image is not automatically a bootable image.

Bootloader handling must use explicit, reviewed board profiles.

## 12. Offline Linux Provisioning

Provide optional preinstallation configuration for verified Armbian images.

Initial settings:

- Root password
- Normal username and password
- SSH authentication
- Hostname
- Locale
- Timezone
- Ethernet DHCP
- Optional static IP configuration

The implementation must inspect the target Armbian version and confirm that each first-boot configuration variable is supported.

Secrets must never appear in application logs.

Any temporary plaintext credential files must have restrictive permissions and be removed when safe.

Support SSH public-key authentication where possible.

Provisioning must be atomic from the user's perspective: failed modifications must not leave the original image presented as ready to install.

## 13. Boot Diagnostics

Implement read-only diagnostics for:

- Partition tables
- Filesystem signatures
- Root filesystem UUIDs
- Kernel and initramfs presence
- Device-tree configuration
- Bootloader metadata
- Armbian boot settings
- `fstab` configuration
- Persistent journal logs
- Network interface state
- Available storage

Diagnostic reports must distinguish verified observations from hypotheses.

Reports must redact passwords, private keys, tokens, and other sensitive data.

Optional UART capture may be added later using external hardware.

## 14. Server Configuration

Provide optional post-installation settings for:

- ZRAM
- Disk swap
- Swap priorities
- SSH
- Hostname
- Network configuration

The application must detect existing managers and avoid conflicting ZRAM services.

Disk swap on eMMC must never be created automatically. Explicitly warn about limited flash capacity, write amplification, and performance implications.

## 15. Resource Requirements

Target a 1 GiB RAM Linux environment.

Engineering objectives:

- Prefer no more than 64 MiB steady-state application memory.
- Use streaming rather than whole-image buffering.
- Start with copy buffers no larger than 8 MiB.
- Minimize unnecessary goroutines.
- Avoid background work during recovery.
- Release file descriptors and mounts on cancellation.
- Never depend on disk swap to run the recovery application.

Memory budgets must be validated on physical hardware before becoming release requirements.

## 16. Testing Strategy

### Unit tests

- Domain validation
- Device classification
- Image parsing
- Capacity checks
- Storage safety
- Manifest validation
- Error handling
- Cancellation
- Secret redaction

### Integration tests

Use temporary files and Linux loop devices for:

- Image installation
- Backup and restore
- Integrity verification
- Corrupted images
- Short writes
- Interrupted operations
- Filesystem mounting
- Offline provisioning
- Cleanup after failure

Never use physical developer storage for automated destructive tests.

### Physical hardware tests

Initial board: MXQ Pro 4K RK3229, 1 GiB RAM, 8 GB eMMC.

Validate:

- Boot from microSD
- HDMI and keyboard
- Storage detection
- Protection of the recovery medium
- Backup and restore
- Image installation to approved test storage
- Boot diagnostics
- Recovery after interruption
- Compatibility with known-working firmware

Keep the original recovery image available throughout testing.

## 17. Implementation Phases

**Phase 1 — Baseline:** Inventory the original source, boot components, licensing, and compatible hardware.

**Phase 2 — Foundation:** Implement Go domain models, application interfaces, Linux discovery, and safety validation.

**Phase 3 — Storage:** Implement backup, restore, image inspection, and installation.

**Phase 4 — Presentation:** Replace original menus with the Go TUI and console fallback.

**Phase 5 — Provisioning:** Implement version-aware Armbian image configuration.

**Phase 6 — Diagnostics:** Add boot, filesystem, network, and optional memory configuration.

**Phase 7 — Integration:** Embed the Go application in the existing recovery rootfs.

**Phase 8 — Validation:** Complete emulation, physical hardware testing, and release-readiness review.

## 18. Acceptance Criteria

- [ ] Go application compiles for Linux ARMv7.
- [ ] Domain and use cases remain independent of TUI implementation.
- [ ] Application operates without internet access.
- [ ] Working boot components remain intact.
- [ ] Hardware detection identifies devices accurately.
- [ ] Unknown hardware fails safely.
- [ ] Active boot media cannot be overwritten.
- [ ] Backup and restore use validated manifests.
- [ ] Image writes are verified.
- [ ] Unsupported bootloader operations are blocked.
- [ ] Offline provisioning works for documented Armbian versions.
- [ ] Diagnostic output redacts secrets.
- [ ] TUI works over HDMI with USB keyboard.
- [ ] Console fallback remains usable.
- [ ] Automated unit and integration tests pass.
- [ ] Physical RK3229 recovery testing succeeds.
- [ ] Recovery documentation is complete.
- [ ] Original recovery image remains available.

## 19. Definition of Done

The modernization is complete when the Go application replaces the required application-level Bash functionality, all storage safety tests pass, the recovered image boots on verified RK3229 hardware, and the project can publish traceable releases without losing its recovery capabilities.

The project must preserve upstream GPL-2.0 obligations and all applicable third-party firmware license requirements.

**Execution order:** Complete the repository foundation first, then implement this specification through the approved Superpowers plan.
