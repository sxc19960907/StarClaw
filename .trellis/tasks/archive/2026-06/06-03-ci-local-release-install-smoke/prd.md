# CI local release install smoke

## Goal

Ensure CI exercises the local release install smoke so archive install regressions are caught before release.

## Requirements

- Add `scripts/smoke_release_local.sh` to the GitHub Actions CI workflow.
- Keep the step after basic Go build/test checks and before browser-heavy Web UI smoke.
- Preserve existing CI smoke steps.
- Keep the workflow readable and bounded; do not add GoReleaser snapshot builds to PR CI.

## Acceptance Criteria

- [x] CI workflow includes a local release install smoke step.
- [x] The step runs `scripts/smoke_release_local.sh`.
- [x] Existing app launch and Web UI core smoke steps remain.
- [x] Local validation passes for workflow syntax-relevant shell scripts and the local release smoke.

## Notes

- This is a lightweight task; PRD-only planning is sufficient.
