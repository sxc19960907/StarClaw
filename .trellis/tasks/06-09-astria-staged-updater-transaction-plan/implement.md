# Astria staged updater transaction plan implementation plan

## Steps

1. Locate Phase19 updater metadata validation helpers.
2. Add a transaction planner adjacent to existing release validation code.
3. Require rollback and post-update health gate declarations while keeping
   replacement disabled.
4. Add a smoke entrypoint and wire it into `--npm-only --astria-local`.
5. Update backend spec for the new transaction planning boundary.
6. Validate with focused smoke, release validation, Go tests, and diff checks.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-staged-updater-transaction-plan`
- `scripts/validate_release_artifacts.sh --updater-transaction-plan-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`
