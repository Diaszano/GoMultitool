# Contributing

## Development setup

Install Go 1.27.3 with mise, then install the pinned tools and run the full local checks:

```sh
mise install
make setup
make check
make security
```

See [local setup](docs/development/local-setup.md) for prerequisites and individual Make targets.

## Branches and commits

Use a branch prefix such as `feature/`, `fix/`, `hotfix/`, `refactor/`, `docs/`, `chore/`, `test/`, `ci/`, or `release/`, followed by a lowercase hyphen-separated description. The default branch is reserved for reviewed pull requests.

Use Conventional Commit subjects such as `feat(storage): add safe device discovery` or `fix: handle empty input`. Mark breaking changes with `!` or a `BREAKING CHANGE:` footer. Run `bash scripts/install-git-hooks.sh` to enable the optional local commit hook; CI checks pull request titles and commit messages.

## Pull requests and checks

Open a pull request for changes. CI runs policy and quality jobs; the separate security workflow scans dependencies, secrets, and GitHub Actions. Recovery image changes also run the isolated image build workflow. Keep documentation, tests, and release metadata aligned with behavior changes.

## Shell and ARMv7 checks

`make lint-shell` checks maintained scripts. The legacy `create_image.sh`, multistrap init/menu scripts, and board boot installers are excluded because they manipulate image and boot components; they remain unchanged pending a behavior-preserving audit. Existing findings are recorded in this repository's history and should not be expanded.

`make build-armv7` creates `build/multitool-linux-armv7` with CGO disabled. The policy check inspects its ELF machine type without running it on physical hardware.

## Releases and recovery images

Software releases use Conventional Commits and are published after CI succeeds on `main`. Recovery images are built separately and remain experimental until hardware tests and approval are recorded. See the [release guide](docs/releases/README.md), [recovery status](docs/recovery/README.md), and [firmware release policy](docs/releases/release-policy.md).
