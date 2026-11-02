# Versioning policy

Releases follow Conventional Commits. `fix` produces a patch release, `feat` a minor release, and a breaking change a major release. Documentation, tests, and maintenance changes do not release unless they contain a breaking-change marker.

The project starts in development mode: the first release is `0.1.0`; before `1.0.0`, breaking changes advance the minor version. After `1.0.0`, breaking changes advance the major version. A change with no release impact leaves the version unchanged.

`.semrelrc` uses go-semantic-release with the Git provider to create version tags. GoReleaser creates the GitHub release and uploads its artifacts. `CHANGELOG.md` is the configured generated changelog. Local version-policy tests use an isolated Git repository and dry-run mode; they do not create tags or releases.
