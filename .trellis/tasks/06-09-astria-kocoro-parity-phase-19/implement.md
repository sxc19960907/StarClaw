# Astria Kocoro parity phase 19 implementation plan

## Checklist

1. `astria-signed-updater-dry-run`
   - Add verifier/dry-run command or script path for Astria updater metadata.
   - Keep replacement disabled.
   - Add smoke coverage for valid, invalid, and replacement-blocked metadata.
2. `astria-release-compatibility-manifest`
   - Add release compatibility manifest generation/checking.
   - Cover app and bundled daemon versions.
3. `astria-local-os-crash-artifact-collection`
   - Add explicit local artifact collection boundary.
   - Reuse redaction and local-only export behavior.
4. Final review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-19`
- `scripts/validate_release_artifacts.sh --updater-boundary-smoke`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not replace the app or bundled daemon.
- Do not require Apple credentials in validation.
- Do not accept updater metadata without checksum/signature/compatibility
  fields.
- Do not collect or upload crash artifacts without explicit user action.
