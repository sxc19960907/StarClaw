# Native desktop diagnostics and recovery UX implementation plan

## Checklist

1. Add a `DesktopRPCDiagnostics` helper in `AstriaApp.swift`.
2. Feed `supervisor.desktopRPCSessionState` into `AstriaRootView` banner
   selection.
3. Add smoke assertions for diagnostic message mapping.
4. Update macOS shell spec with diagnostic redaction and session-state UX
   requirements.
5. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-native-desktop-diagnostics-recovery`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Do not add a separate native frontend.
- Do not show Desktop RPC paths or raw event payloads.
- Do not hide daemon crash/failure banners behind lower-priority RPC state.
