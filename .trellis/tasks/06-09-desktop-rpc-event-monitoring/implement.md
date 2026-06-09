# Desktop RPC event monitoring implementation plan

## Checklist

1. Add `DesktopEventMonitor` under `internal/daemon`.
2. Add safe event metadata to `desktop_rpc.Status`.
3. Wire `Server` to own the monitor and expose it through `desktopRPCStatus`.
4. Wire daemon listener startup to pass `EventSink: server.RecordDesktopEvent`.
5. Add tests:
   - monitor bounded retention and redacted status;
   - listener event sink dispatch;
   - daemon `/status` includes safe event metadata.
6. Run validation:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-event-monitoring`
   - `go test ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
   - `scripts/smoke_macos_astria_shell.sh`
   - `go test ./...`
   - `git diff --check`

## Risk Points

- Do not expose raw event payloads through `/status`.
- Do not make event monitoring required for browser/CLI flows.
- Keep retention bounded and protected by a mutex.
- Keep the Desktop RPC wire format unchanged.
