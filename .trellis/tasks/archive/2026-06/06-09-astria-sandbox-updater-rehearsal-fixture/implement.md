# Astria sandbox updater rehearsal fixture implementation plan

## Steps

1. Add `RUN_SANDBOX_UPDATER_REHEARSAL_SMOKE` flag and argument parsing.
2. Add `run_astria_sandbox_updater_rehearsal_smoke`.
3. Implement sandbox path guard helper in the smoke.
4. Add valid rehearsal and outside-path rejection cases.
5. Run the smoke under `--npm-only --astria-local`.
6. Update README and backend spec.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-sandbox-updater-rehearsal-fixture`
- `scripts/validate_release_artifacts.sh --sandbox-updater-rehearsal-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`
