# Harden Release Checklist for npm

## Problem

The npm installer now works locally, but release workflow and release documentation do not yet make npm verification explicit. This creates risk that a GitHub release succeeds while the npm package wrapper or installer is broken.

## Requirements

- Add release documentation that describes GoReleaser assets, npm pack verification, npm publish preconditions, and smoke commands.
- Extend release artifact validation to include npm package packaging checks.
- Ensure CI/release-local validation has a clear npm smoke entry point.
- Do not add npm auth/token publishing automation in this task.

## Acceptance Criteria

- [x] Release checklist documentation exists and includes npm verification/publish steps.
- [x] `scripts/validate_release_artifacts.sh` validates npm pack metadata/content.
- [x] Release workflow runs npm smoke or artifact validation where appropriate.
- [x] `scripts/validate_release_artifacts.sh` and `scripts/smoke_npm_install.sh` pass locally where tool prerequisites are available.
- [x] `git diff --check` passes.
