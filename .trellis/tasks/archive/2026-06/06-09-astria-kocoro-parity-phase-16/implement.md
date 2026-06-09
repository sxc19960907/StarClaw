# Astria Kocoro parity phase 16 implementation plan

## Checklist

1. `astria-native-menu-dock-window`
   - Add command model for New Window, Reload, Diagnostics, and Retry Daemon.
   - Wire SwiftUI commands to shared app action state.
   - Add smoke coverage for command availability and labels.
2. `astria-native-diagnostics-crash-reports`
   - Add local diagnostics export/failure report boundary.
   - Add redaction tests or smoke coverage.
3. `astria-signing-notarization-updater-boundary`
   - Update release validation scripts/docs for signing/notarization/updater
     metadata boundaries.
   - Keep local development usable without credentials.
4. Final review.
   - Update Kocoro parity estimate and remaining gaps.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-16`
- `scripts/smoke_macos_astria_shell.sh`
- `scripts/validate_release_artifacts.sh --npm-only --astria-local`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not add private signing credentials or updater secrets.
- Do not route diagnostics off-machine.
- Do not make native commands depend on a healthy daemon unless the command
  explicitly performs daemon retry or diagnostics.
