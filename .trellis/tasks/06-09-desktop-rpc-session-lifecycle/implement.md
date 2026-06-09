# Desktop RPC session lifecycle implementation plan

## Checklist

1. Add `DesktopRPCSessionState` and monitor task ownership to
   `DaemonSupervisor`.
2. Start the monitor after successful `reconcileDesktopRPC()`.
3. Stop/restart the monitor when daemon supervision restarts.
4. Add bounded retry options to `LaunchConfig` or internal defaults.
5. Add reusable session probe logic around `DesktopRPCClient.systemCapabilities`.
6. Add `--desktop-rpc-session-smoke` to validate connected/retry/mismatch
   behavior.
7. Wire the new smoke into `scripts/smoke_macos_astria_shell.sh`.
8. Run validation:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-session-lifecycle`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Avoid leaking Swift tasks when daemon supervision restarts.
- Keep retry intervals short in smoke tests and conservative in runtime.
- Do not change Desktop RPC framing in this child task.
- Do not make missing Desktop RPC fatal when HTTP fallback is healthy.
