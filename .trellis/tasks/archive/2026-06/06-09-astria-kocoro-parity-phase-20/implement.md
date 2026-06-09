# Astria Kocoro parity phase 20 implementation plan

## Checklist

1. `astria-staged-updater-transaction-plan`
   - Add a local transaction planner for Astria updater release inputs.
   - Require replacement disabled, verified metadata fields, compatibility
     fields, and declared safety gates.
   - Add smoke coverage for valid, missing-gate, and replacement-enabled inputs.
2. `astria-updater-rollback-health-gates`
   - Define rollback and post-update health gate manifest fields.
   - Validate gate completeness without performing replacement.
3. `astria-release-acceptance-gates`
   - Extend release validation to reject unsafe production release metadata.
   - Keep local validation credential-free.
4. Final review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-20`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not replace the app or bundled daemon.
- Do not require Apple credentials in validation.
- Do not accept replacement-enabled metadata before rollback and health gates
  are implemented and explicitly approved.
- Do not weaken Phase19 redaction or compatibility checks.
