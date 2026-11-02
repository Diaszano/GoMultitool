# RK322x Multitool — Repository Standards, CI/CD and Automated Releases Design

**Status:** Proposed  
**Date:** 2026-11-02  
**Type:** Superpowers Design Specification  
**Priority:** Critical  
**Implementation Order:** First

## 1. Overview

Establish a professional, maintainable, secure repository for the Go-based RK322x Multitool.

The repository must have enforced coding standards, reproducible builds, automated checks, predictable releases, dependency management, and complete contributor documentation.

Use Conventional Commits, Conventional Branches, Semantic Versioning, and GitHub Actions.

## 2. Goals

1. Preserve upstream history, attribution, and licensing.
2. Establish an idiomatic Go repository structure.
3. Implement consistent formatting and linting.
4. Automate tests and ARMv7 cross-compilation.
5. Enforce Conventional Commits.
6. Enforce Conventional Branches.
7. Configure GitHub branch protection.
8. Automate semantic version calculation.
9. Publish Go binaries and GitHub Releases.
10. Produce checksums, changelogs, and provenance.
11. Introduce separately gated recovery-image publication.
12. Automate dependency and security maintenance.
13. Document contribution, recovery, and release procedures.

## 3. Development Standards

### Language

- All source code, comments, documentation, and Git metadata in English.
- User-facing strings default to Brazilian Portuguese.
- Idiomatic Go naming for packages, types, interfaces, and functions.
- Avoid unnecessary abstractions.
- Follow SOLID, Clean Architecture, and pragmatic DDD.
- Prefer explicit error handling.
- Never ignore errors in storage, boot, or recovery operations.
- No embedded credentials or production secrets.

### Project naming

Proposed Go module:

`github.com/Diaszano/rk322x-multitool`

Confirm the final repository path before creating `go.mod`.

### Repository layout

```text
.github/
  workflows/
    ci.yml
    release.yml
    image-build.yml
    security.yml
  ISSUE_TEMPLATE/
  PULL_REQUEST_TEMPLATE.md
  CODEOWNERS
  dependabot.yml

cmd/
internal/
configs/
build/
scripts/
tests/
tools/

docs/
  architecture/
  development/
  recovery/
  releases/
  supported-boards/
  superpowers/
    specs/
    plans/

.editorconfig
.gitattributes
.gitignore
.golangci.yml
.goreleaser.yaml
.semrelrc
.markdownlint.yaml
.yamllint.yaml
go.mod
go.sum
Makefile
README.md
CONTRIBUTING.md
SECURITY.md
CODE_OF_CONDUCT.md
LICENSE
NOTICE
CHANGELOG.md
```

Preserve upstream directories until their functionality has been audited and migrated.

## 4. Go Formatting and Linting

Use:

- gofmt
- goimports
- golangci-lint v2
- go vet

Recommended linters:

- errcheck
- govet
- staticcheck
- ineffassign
- unused
- revive
- misspell
- bodyclose
- durationcheck
- errorlint
- gocritic

All tools must use pinned versions.

Formatting violations and meaningful lint violations must fail CI.

Lint suppressions must be narrow and justified.

## 5. Shell Quality

Retained build and initialization scripts must use:

- ShellCheck
- shfmt

Prefer strict Bash handling when compatible:

```bash
#!/usr/bin/env bash
set -Eeuo pipefail
```

Existing scripts must be audited before applying strict mode.

Scripts must quote arguments, validate input, handle errors, and clean up temporary mounts and loop devices.

## 6. Additional Quality Tools

- EditorConfig
- yamllint
- markdownlint
- actionlint
- zizmor
- gitleaks
- govulncheck
- Dependabot
- Trivy where appropriate for images or filesystem scanning

Avoid redundant tools or configuration files with no actual use.

## 7. Developer Experience

Provide these Make targets:

```text
make help
make setup
make fmt
make fmt-check
make lint
make test
make test-race
make build
make build-armv7
make integration-test
make image
make security
make check
make clean
```

Support Linux Mint as the primary development environment.

Support Go cross-compilation from macOS.

Privileged image construction may be Linux-only.

Unit tests must not require root. Privileged integration tests must use isolated loop devices.

## 8. Conventional Commits

Use Conventional Commits 1.0.0.

Format:

```text
<type>(<scope>): <description>
```

Accepted types:

- feat
- fix
- refactor
- perf
- test
- docs
- build
- ci
- chore
- style
- revert

Examples:

```text
feat(storage): add safe eMMC detection
fix(boot): handle missing root partition
refactor(installer): extract block writer adapter
test(backup): cover interrupted reads
ci(release): automate version publishing
```

Breaking changes must be marked with `!` or a `BREAKING CHANGE:` footer.

Enforce PR titles and merged commit messages.

Prefer squash merges with Conventional Commit titles.

## 9. Conventional Branches

Protected default branch: `main`.

Accepted working branch prefixes:

```text
feature/
fix/
hotfix/
refactor/
docs/
chore/
test/
ci/
release/
```

Examples:

```text
feature/go-tui
fix/emmc-detection
refactor/storage-adapters
ci/semantic-release
docs/recovery-guide
```

Names must be lowercase with hyphen-separated descriptions.

Branch naming must be validated in CI.

## 10. Pull Request Governance

Every change should use a pull request.

Required protections:

- No force pushes to `main`.
- No default-branch deletion.
- Required CI checks.
- Required reviews for sensitive changes.
- Resolved review conversations.
- Restricted bypass permissions.
- Protected release workflows.

CODEOWNERS must cover storage writing, firmware, bootloader, and release logic.

Pull request descriptions must document:

- Summary
- Motivation
- Tests
- Safety implications
- Hardware validation
- Breaking changes
- Documentation updates

## 11. CI Pipeline

GitHub Actions workflow: `.github/workflows/ci.yml`

Run on pull requests and pushes to `main`.

Required jobs:

1. Validate branch naming.
2. Validate commit messages and PR titles.
3. Check formatting.
4. Run Go linting.
5. Run ShellCheck and shfmt.
6. Validate YAML and GitHub workflows.
7. Run unit tests.
8. Run supported race tests.
9. Run integration tests.
10. Cross-compile Linux ARMv7.
11. Run supported ARMv7 smoke tests.
12. Produce test artifacts.

Never execute destructive tests against host disks.

## 12. Build Pipeline

GitHub Actions workflow: `.github/workflows/image-build.yml`

Build stages:

1. Resolve pinned toolchains and dependencies.
2. Cross-compile the Go ARMv7 binary.
3. Prepare the minimal Linux rootfs.
4. Integrate verified boot components.
5. Build the recovery disk image.
6. Inspect the partition table and filesystem layout.
7. Validate boot files.
8. Compress with XZ.
9. Generate SHA-256 checksums.
10. Generate a release manifest.

Output names:

```text
multitool-rk322x-vX.Y.Z.img.xz
multitool-rk322x-vX.Y.Z.sha256
multitool-linux-armv7-vX.Y.Z.tar.gz
```

Include source commit, kernel version, and bootloader provenance in the manifest.

Builds must be isolated and clean up all temporary loop devices and mounts.

## 13. Semantic Versioning

Use Semantic Versioning 2.0.0.

Tags:

```text
vMAJOR.MINOR.PATCH
```

Mapping:

| Commit            | Default impact         |
| ----------------- | ---------------------- |
| fix               | Patch                  |
| feat              | Minor                  |
| Breaking change   | Major                  |
| docs, test, chore | None                   |
| ci, build         | None unless configured |

Define and test special behavior for versions below `v1.0.0`.

Initialize the first release intentionally.

## 14. Automated Releases

Use go-semantic-release for commit analysis and version selection.

Use GoReleaser v2 for binary packaging.

Workflow:

1. CI succeeds on `main`.
2. Analyze merged Conventional Commits.
3. Determine whether a release is needed.
4. Calculate the next version.
5. Create one authoritative Git tag.
6. Build and verify artifacts.
7. Generate release notes.
8. Publish a GitHub Release.
9. Upload binaries and checksums.
10. Record provenance.

The process must be idempotent.

Two tools must not independently create competing tags.

Repeated workflow execution must not duplicate a release.

## 15. Release Channels and Safety

Application binaries may be released automatically once CI passes.

Recovery images require additional validation because they can modify flash storage.

An image must not be advertised as stable without:

- Valid image structure
- Boot component validation
- Integrity checks
- Relevant hardware compatibility evidence
- Documented recovery instructions

Images without physical validation must remain experimental or require manual promotion.

## 16. Changelog and Version Information

Generate changelogs from Conventional Commits.

Organize relevant changes into:

- Features
- Fixes
- Performance
- Breaking Changes
- Security
- Other Changes

The executable must expose:

```text
multitool version
```

Example:

```text
RK322x Multitool v1.2.3
Commit: <commit>
Build: <timestamp>
```

## 17. Supply-Chain Security

Use least-privilege GitHub Actions permissions.

Requirements:

- Pin third-party actions.
- Avoid write-capable tokens for untrusted PRs.
- Do not expose release credentials to forked PRs.
- Scan for secrets.
- Check Go vulnerabilities.
- Validate dependencies.
- Audit third-party firmware blobs.
- Generate checksums.
- Add artifact provenance and SBOM where supported.
- Review dependency updates before merging.

Use scheduled security checks.

## 18. Dependency Automation

Configure Dependabot for:

- Go modules
- GitHub Actions

Group compatible updates where appropriate.

Major upgrades and boot-related dependency updates require explicit review.

Firmware binaries must have their own versioned provenance and compatibility process.

## 19. Documentation

Required files:

- `README.md`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `CODE_OF_CONDUCT.md`
- `NOTICE`
- Architecture documentation
- Development setup
- Supported hardware matrix
- Recovery instructions
- Release policy

The README must distinguish verified hardware support from experimental support.

Document all storage operations capable of causing data loss.

## 20. Licensing

Preserve upstream GPL-2.0 obligations, copyright notices, and attribution.

Record the upstream source revision.

Audit licenses and redistribution conditions for firmware and bootloader binaries separately.

Do not assume proprietary firmware inherits the application source license.

## 21. Testing Policy

Use:

- Go unit tests
- Linux integration tests with temporary loop devices
- ARMv7 cross-compilation
- Supported emulated smoke tests
- Hardware validation for firmware releases

Race tests run only on supported architectures.

Test coverage must focus on meaningful behaviors, particularly failure paths and storage safety.

Global coverage percentage must not replace actual safety testing.

## 22. Workflow Summary

```text
Feature Branch
     |
     v
Pull Request
     |
     v
Policy + Format + Lint
     |
     v
Tests + Security + ARMv7 Build
     |
     v
Review and Approval
     |
     v
Squash Merge to main
     |
     v
Semantic Release
     |
     v
vX.Y.Z + GoReleaser
     |
     v
GitHub Release
     |
     v
Binaries + Checksums
     |
     v
Recovery Image Validation Gate
```

## 23. Implementation Phases

**Phase 1:** Preserve upstream history and initialize Go.

**Phase 2:** Establish formatters, linters, and developer tooling.

**Phase 3:** Enforce Git branch, commit, and PR rules.

**Phase 4:** Implement continuous integration.

**Phase 5:** Implement GoReleaser and semantic versioning.

**Phase 6:** Automate recovery-image builds.

**Phase 7:** Introduce security scanning, provenance, and documentation.

## 24. Acceptance Criteria

- [ ] Repository has a documented Go structure.
- [ ] `make check` runs all required local checks.
- [ ] Format and lint violations fail CI.
- [ ] Unit and integration tests pass.
- [ ] Linux ARMv7 builds work.
- [ ] Conventional Commits enforced.
- [ ] Conventional Branches enforced.
- [ ] Protected `main` configured.
- [ ] Semantic versions follow `vX.Y.Z`.
- [ ] Eligible merges trigger releases automatically.
- [ ] No duplicate tags or releases.
- [ ] Go binaries have injected version information.
- [ ] Changelog and release notes are generated.
- [ ] Image artifacts have checksums and manifests.
- [ ] Unverified firmware cannot be promoted as stable.
- [ ] Dependency automation is enabled.
- [ ] Security checks are operational.
- [ ] GPL and third-party firmware obligations are documented.
- [ ] Contributor and recovery documentation is complete.

## 25. Definition of Done

The repository foundation is complete when all required quality gates pass, the protected branch is configured, semantic release behavior is validated, and a release can be reproduced and audited from its source revision.

**Execution order:** Complete this specification before the main Go modernization work.
