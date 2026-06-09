# Astria updater rollback health gates implementation plan

## Steps

1. Add rollback/health gate manifest write and assert helpers in
   `scripts/validate_release_artifacts.sh`.
2. Add a `--updater-rollback-health-gates-smoke` entrypoint.
3. Run the smoke as part of `--npm-only --astria-local`.
4. Update Astria README and backend spec with the rollback/health gate contract.
5. Validate with focused smoke, local release validation, macOS shell smoke,
   Go tests, and diff checks.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-updater-rollback-health-gates`
- `scripts/validate_release_artifacts.sh --updater-rollback-health-gates-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`
