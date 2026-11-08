# Local development setup

## Prerequisites

- [mise](https://mise.jdx.dev/), GNU Make, Git, and Bash
- Network access for the first tool installation and Go module download

Install Go and every repository tool from the repository root:

```sh
mise install
go version
```

`.mise.toml` pins Go and the development tools (goimports, golangci-lint, ShellCheck, shfmt, actionlint, govulncheck, gitleaks, zizmor, GoReleaser, go-semantic-release, markdownlint-cli2, and yamllint), plus the Node.js and uv runtimes they need. `mise.lock` records each tool's download URL and SHA-256 for Linux x64 and macOS arm64, and `.mise/locks/` pins the npm and Python dependencies of markdownlint-cli2 and yamllint; CI installs from the same files with `jdx/mise-action`. `make setup` is an alias for `mise install`. Run the Make targets from a shell where mise is activated, or prefix them with `mise exec --`.

To change a tool version, edit `.mise.toml` and refresh the lockfile with `mise lock --platform linux-x64,macos-arm64`.

## Build and validate

```sh
make help
make check
make security
make build
make build-armv7
```

`make check` includes formatting, Go lint, shell lint, documentation lint, Go tests, race tests, integration checks, and ARMv7 cross-compilation. `make security` runs vulnerability, secret, and GitHub Actions workflow scans.

To build the host command directly, run `./build/multitool version` after `make build`. `make build-armv7` writes `build/multitool-linux-armv7`; it does not execute that ARM binary.

## Recovery image

`make image` currently accepts only `rk322x`. It requires a Linux host with root privileges, loop devices, and the packages listed in `.github/workflows/image-build.yml`. The target attaches only the generated regular image file to loop devices, validates it, and writes the compressed image, checksum, and manifest under `image-dist/`.

A documentation completeness checklist for a change is:

- [ ] Setup and validation instructions still match the Make targets.
- [ ] New commands or configuration have a linked explanation.
- [ ] Hardware support and release claims match available test evidence.
- [ ] Recovery or firmware changes describe their risks and required approval.
