# Architecture overview

The repository has two cooperating parts:

- `cmd/multitool` is the Go command-line entry point. It currently exposes `multitool version`; `internal/buildinfo` supplies version, commit, and build-time fields.
- `sources/multistrap`, `sources/rk*`, `create_image.sh`, and `tools/` preserve the existing Linux recovery environment, board configurations, boot assets, and image utilities.

The Go code is additive. It does not replace the legacy recovery menu or alter its boot flow. The [component inventory](component-inventory.md) lists the preserved files, and [the baseline record](upstream-baseline.md) identifies the source commit.

## Build and release boundaries

`make check` runs Go, shell, documentation, integration, and ARMv7 checks. GitHub Actions runs CI for pull requests. The software release workflow waits for CI on `main`, selects a version from Conventional Commits, creates its tag, and packages Linux amd64 and ARMv7 binaries with GoReleaser.

Recovery images use a separate workflow. It wraps the existing image builder, embeds the ARMv7 command, validates a loop-backed image file, and uploads it as an experimental workflow artifact. It does not flash physical media or publish firmware as a stable release. Hardware approval requirements are described in the [firmware release policy](../releases/release-policy.md).
