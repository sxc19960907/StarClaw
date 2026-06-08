# Daemon SSE event vocabulary parity implementation plan

## Checklist

1. Read task artifacts and backend specs.
2. Add a small `writeEvent` helper on `sseEventHandler` to reduce repeated `fmt.Fprintf` calls.
3. Add `SetSessionID` on `sseEventHandler` and wire it through the existing optional session-id callback path if present.
4. Update `OnToolCall`, `OnToolResult`, `OnStreamDelta`, `OnPreamble`, and `OnUsage` to dual-emit legacy and Kocoro-compatible events.
5. Add tests for:
   - legacy `tool_call`/`tool_result` still present;
   - new `tool` running/completed/error aliases;
   - `delta` plus legacy `text` on stream deltas;
   - `assistant_text` plus legacy `preamble`;
   - `usage`;
   - `session_started`.
6. Run validation:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-daemon-sse-event-vocabulary-parity`
   - `go test ./internal/daemon`
   - `go test ./...`
   - `rg -n "event: (delta|usage|tool|assistant_text|session_started)|tool_call|tool_result|preamble" internal/daemon`

## Risk Points

- Dual-emission increases event volume. Keep payloads compact and reuse redaction helpers.
- `OnText` must still avoid duplicate final answer when deltas were streamed.
- `SetSessionID` must be a no-op for empty ids and nil writers.

## Rollback

All code changes should be isolated to `internal/daemon/server.go` and `internal/daemon/server_test.go`. Reverting this task should not affect OpenAI gateway streaming.
