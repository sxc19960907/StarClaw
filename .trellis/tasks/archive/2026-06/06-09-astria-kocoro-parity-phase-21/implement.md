# Astria Kocoro parity phase 21 implementation plan

## Checklist

1. `astria-sandbox-updater-rehearsal-fixture`
   - Add a release validation smoke for sandbox-only fixture replacement and
     rollback.
   - Prove all touched paths are under a temporary sandbox.
   - Wire smoke into `--npm-only --astria-local`.
2. `astria-sandbox-updater-health-rehearsal`
   - Add simulated health checks against the fixture.
3. `astria-sandbox-updater-rollback-rehearsal`
   - Add failed replacement rollback rehearsal.
4. Final review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-21`
- `scripts/validate_release_artifacts.sh --sandbox-updater-rehearsal-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not touch real installed app paths.
- Do not introduce download/install behavior.
- Do not require Apple credentials.
- Do not weaken Phase20 release gates.
