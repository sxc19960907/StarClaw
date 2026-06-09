# Astria native clipboard file affordances implementation plan

## Checklist

1. Add command specs for Copy Current Route, Copy Support Summary, and Reveal
   Diagnostics Folder.
2. Add action closures and root-view wiring.
3. Add support summary builder with redaction.
4. Extend native command/diagnostics smoke coverage.
5. Update macOS shell spec.
6. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-native-clipboard-file-affordances`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Do not copy full external URLs or unsafe stored routes.
- Do not copy raw secret-like values.
- Do not auto-reveal files without user action.
