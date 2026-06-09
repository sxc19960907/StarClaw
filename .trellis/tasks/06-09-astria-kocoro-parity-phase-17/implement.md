# Astria Kocoro parity phase 17 implementation plan

## Checklist

1. `astria-native-clipboard-file-affordances`
   - Add native command/action model for safe clipboard/file actions.
   - Add redacted support text generation.
   - Add smoke coverage.
2. `astria-native-permission-helper-flows`
   - Add permission helper status/copy guidance.
   - Keep unsigned local builds functional.
3. `astria-multi-window-state-restoration`
   - Improve per-window route state behavior.
   - Add smoke coverage for route safety.
4. Final review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-17`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not copy raw secrets or private paths to the clipboard.
- Do not reveal or upload files automatically.
- Do not loosen route safety.
