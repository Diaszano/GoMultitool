# Local development setup

## Prerequisites

- Go 1.27.3, selected by `.mise.toml`
- GNU Make, Git, Bash, curl, tar, Node.js/npm, and Python 3/pip
- Network access for the first tool installation and Go module download

Install Go and the repository tools from the repository root:

```sh
mise install
go version
make setup
```

`make setup` installs versions from `.tools-versions`; checksum-verified shell scripts install ShellCheck, zizmor, GoReleaser, and go-semantic-release. It installs tools under the current Go `GOPATH` and does not modify the project module requirements.

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
