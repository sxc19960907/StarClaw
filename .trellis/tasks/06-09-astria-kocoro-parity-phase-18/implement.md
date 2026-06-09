# Astria Kocoro parity phase 18 implementation plan

## Checklist

1. `astria-local-crash-reporter-summaries`
   - Add local crash summary/report affordance.
   - Reuse diagnostics redaction boundaries.
   - Add smoke coverage for local-only and redaction behavior.
2. `astria-native-notification-readiness`
   - Add notification status/guidance boundary.
   - Avoid surprise permission prompts.
   - Add smoke coverage.
3. `astria-signed-updater-release-boundary`
   - Harden updater/release metadata validation.
   - Keep validation credential-free.
4. Final review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-18`
- `scripts/smoke_macos_astria_shell.sh`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not upload crash reports automatically.
- Do not trigger notification permission prompts from passive readiness checks.
- Do not commit signing/notarization/updater secrets.
- Do not loosen local-only diagnostic redaction.
