# Releases

## Software releases

Pull requests must pass the `CI / policy` and `CI / quality` checks. Successful changes merged to `master` are analyzed from Conventional Commit messages. `fix` selects a patch, `feat` a minor, and breaking changes a major after `1.0.0`; before `1.0.0`, breaking changes select a minor. Documentation-only changes do not create a release.

The release workflow creates the version tag after CI succeeds. GoReleaser packages Linux amd64 and ARMv7 binaries, archives, and SHA-256 checksums. See the [versioning policy](versioning.md) and [contributor guide](../../CONTRIBUTING.md).

Validate packaging locally without publishing:

```sh
goreleaser check
goreleaser release --snapshot --clean
```

## Recovery images

The image workflow is separate from the software release workflow. It builds the RK322x image, validates its partitions and boot files, and uploads a `recovery-image-experimental-*` workflow artifact. It does not attach firmware to a GitHub Release. Images without documented hardware validation must remain experimental; see the [firmware release policy](release-policy.md).

The image manifest records the source commit and SHA-256 identifiers for boot components. A pinned SBOM generator is not currently part of the image toolchain, so the workflow does not claim to provide an SBOM.
