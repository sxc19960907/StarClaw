# Desktop RPC capabilities reconciliation implementation plan

## Checklist

1. Add Astria runtime path configuration.
   - Add `desktopRPCSocketPath` and `desktopRPCPidfilePath` to `LaunchConfig`.
   - Resolve from `ASTRIA_RUNTIME_DIR` or Application Support.
2. Add Swift Desktop RPC client.
   - Define frame/request/result/capability structs.
   - Implement Unix socket connect/read/write helpers.
   - Implement `system.capabilities` request.
   - Implement validation for protocol and required methods.
3. Wire daemon launch.
   - Append `--rpc-socket` and `--rpc-pidfile` to `starclaw daemon start`.
   - Preserve `ASTRIA_STARCLAW_BIN`, bundled binary, and PATH resolution.
4. Wire supervisor readiness.
   - For a daemon launched by Astria, wait for health then reconcile Desktop
     RPC before setting `.attached`.
   - Keep existing HTTP attach behavior for already-running daemons.
   - Surface Desktop RPC failure as `.failed(...)`.
5. Add smoke coverage.
   - Extend `--supervision-smoke` to verify capabilities after launch.
   - Add a validation smoke for protocol/method mismatch helpers.
   - Keep missing binary failure coverage.
6. Update docs/spec.
   - macOS shell README.
   - backend Desktop RPC spec.
7. Validate.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-capabilities-reconciliation`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not block attach to a healthy daemon that was not launched by the shell in
  this child; fallback/recovery hardening is next.
- Do not delete stale sockets or pidfiles here.
- Do not expose socket/pidfile paths in daemon HTTP status.
- Keep Desktop RPC frame limits and request field names aligned with
  `internal/daemon/desktop_rpc`.
- Avoid Swift network abstractions that do not support Unix domain sockets
  cleanly in a `swiftc` single-file app build.
