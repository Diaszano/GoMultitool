# Supply-chain security

Pull requests, pushes to the default branch, and a weekly schedule run `make security`. It checks Go vulnerabilities, scans tracked and working-tree files for secrets with redacted output, and audits GitHub Actions with actionlint and zizmor. Local zizmor checks run offline; CI adds its online audits with a read-only workflow token. Pull requests also run GitHub dependency review and fail on moderate or higher severity advisories.

Third-party GitHub Actions are pinned to full commit SHAs. Go and every development tool and scanner binary are pinned in `.mise.toml`; `mise.lock` records the SHA-256 of each download, and mise verifies it before installation, locally and in CI. Dependabot proposes weekly Go module and Actions updates for review.

Do not add credentials, private host keys, or production tokens to this repository. If a secret is detected, revoke or rotate it first, then remove it from the affected files and history using the repository's incident procedure. Preserve firmware provenance and verify license terms before redistributing binary assets.
