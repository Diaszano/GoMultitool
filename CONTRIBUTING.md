# Contributing

## Development setup

Install Go 1.27.3 and [mise](https://mise.jdx.dev/), then run:

```sh
make setup
make check
```

Run `make help` to see the available local commands. Use Conventional Commits, for example `feat: add device discovery` or `fix(storage): reject mounted targets`.

## Shell script checks

Install ShellCheck 0.11.0 and run `make lint-shell`; `make setup` installs the pinned shfmt version. The legacy `create_image.sh`, multistrap init/menu scripts, and board boot installers are excluded because they manipulate image and boot components; they remain unchanged until a separate behavior-preserving audit can verify them.

A ShellCheck 0.11.0 audit found pre-existing findings in the excluded legacy files: 54 in `create_image.sh`, 126 in `sources/multistrap/multitool.sh`, 13 in its completion hook, and 39 across `init` and the four board boot installers. These findings are left intact pending behavior-preserving review.

## ARMv7 build

`make build-armv7` creates `build/multitool-linux-armv7` with CGO disabled. The policy check inspects its ELF machine type without running it on physical hardware.

## Branch names

Use `feature/`, `fix/`, `hotfix/`, `refactor/`, `docs/`, `chore/`, `test/`, `ci/`, or `release/` followed by a lowercase hyphen-separated description. The protected default branch and semantic release channels (`main`, `next`, `next-major`, `beta`, `alpha`, and `N.x`) are also valid.

## Commit messages

Use Conventional Commit subjects such as `feat(storage): add safe device discovery` or `fix: handle an empty input`. Mark breaking changes with `!` or a `BREAKING CHANGE:` footer. Run `bash scripts/install-git-hooks.sh` to enable the optional local commit hook; CI validates pull request titles with the same policy.
