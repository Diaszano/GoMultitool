# Upstream baseline

This repository starts from commit `d7dc392fe33850f5975aa466538fd889629c430c` (recorded with `git log -1 --format=%H` on 2026-11-02). The configured upstream is `git@github.com:Diaszano/GoMultitool.git`.

The original recovery tooling remains intact in that commit. Application modernization is additive and must retain a bootable legacy recovery path until compatibility is verified.

The root `LICENSE` is GNU GPL version 2. Preserve upstream copyright and per-file notices; bundled firmware and binary artifacts may have separate terms and should be reviewed before redistribution.

No built recovery image is checked into the repository. `create_image.sh` builds one from the board configuration and files in `sources/`, using tools in `tools/`.

## Baseline verification

Compare the tracked source inventory against the recorded commit:

```sh
git ls-tree -r --name-only d7dc392fe33850f5975aa466538fd889629c430c -- create_image.sh sources tools
```
