# Default branch protection

The ruleset in `.github/rulesets/main.json` targets GitHub's current default branch, requires pull requests and one approving review, requires code owner review and resolved conversations, and blocks deletion and force pushes. It requires the `policy` and `quality` checks.

Applied on 2026-11-02 as repository ruleset `default-branch-protection` (ID `26399784`), targeting the current default branch `main`. The API confirmed active enforcement, one required approval, code-owner review, last-push approval, resolved review threads, and required `policy` and `quality` checks. No bypass actors are configured.

The workflow defines both required checks in `.github/workflows/ci.yml`; GitHub reports their names as `policy` and `quality`. Repository administrators should review bypass access and keep it limited to emergency maintenance.
