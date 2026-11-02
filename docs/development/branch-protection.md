# Default branch protection

The ruleset in `.github/rulesets/main.json` targets GitHub's current default branch, requires pull requests and one approving review, requires code owner review and resolved conversations, and blocks deletion and force pushes. It requires the `CI / policy` and `CI / quality` checks.

Review the required check names against `.github/workflows/ci.yml` before applying the ruleset. Apply or update it with an administrator token after CI is available:

```sh
gh api --method POST repos/Diaszano/GoMultitool/rulesets --input .github/rulesets/main.json
```

Check the result under **Settings → Rules → Rulesets** and verify that each required check is reported by the CI workflow. Repository administrators should review bypass access and keep it limited to emergency maintenance.
