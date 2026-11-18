# RK322x Multitool

This repository preserves the original Linux recovery environment for Rockchip TV boxes while adding a Go command-line application and automated repository tooling. The Go application provides version, storage inventory, and read-only diagnostics commands; the existing Bash recovery menu remains the default device recovery interface until hardware validation is complete.

## Getting started

Install Go 1.27.3 and the pinned development tools with [mise](https://mise.jdx.dev/), then run the local checks:

```sh
mise install
make check
```

See [local setup](docs/development/local-setup.md) for prerequisites and individual build commands.

## Repository guide

- [Architecture overview](docs/architecture/overview.md)
- [Contributing](CONTRIBUTING.md)
- [Recovery and hardware status](docs/recovery/README.md)
- [RK3229 recovery procedure](docs/recovery/rk3229.md)
- [Hardware validation matrix](docs/supported-boards/validation-matrix.md)
- [Build and release process](docs/releases/README.md)
- [Security policy](SECURITY.md)
- [Supported board configurations](docs/architecture/component-inventory.md)

## Builds

`make build` creates the host Go executable. `make build-armv7` cross-compiles it for Linux ARMv7. `make image` builds an experimental RK322x recovery image and requires Linux, root privileges, and the image tools used by CI. It writes a file-backed image; flashing a device is a separate operation and is not performed by this target.

The image workflow checks partition structure, filesystems, boot files, and embedded binaries before uploading an experimental workflow artifact. The new image pipeline has not established physical hardware compatibility. See the [recovery policy](docs/releases/release-policy.md) before distributing firmware.

## Upstream and license

The legacy recovery project and assets are preserved from the baseline recorded in [upstream-baseline.md](docs/architecture/upstream-baseline.md). The repository license is GPL-2.0; bundled firmware and binary components may have additional terms. Review [NOTICE](NOTICE) before redistributing those assets.
