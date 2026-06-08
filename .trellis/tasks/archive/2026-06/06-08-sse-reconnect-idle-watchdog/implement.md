# SSE reconnect idle watchdog implementation plan

## Steps

1. Read backend specs and existing client/server SSE tests.
2. Add `SSEConnectOptions` and `ConnectWithOptions` in `internal/client/sse.go`.
3. Extend reconnect loop to track last event ID, send `Last-Event-ID`, support idle timeout, bounded reconnect, and context-aware backoff.
4. Add client tests mirroring Kocoro scenarios:
   - idle timeout reconnects and sends `Last-Event-ID`,
   - reconnect budget exhaustion closes after `MaxReconnects + 1` attempts,
   - cancellation during reconnect delay closes promptly.
5. Inspect `/events` replay behavior and add daemon tests for query/header cursors or implement replay if missing.
6. Validate:
   - `go test ./internal/client`
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-sse-reconnect-idle-watchdog`

## Rollback

Revert this task's edits to `internal/client/sse.go`, `internal/client/sse_test.go`, daemon SSE files/tests if touched, and this task directory.
