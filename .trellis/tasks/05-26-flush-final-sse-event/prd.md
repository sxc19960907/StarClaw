# Flush final SSE event

## Goal

Prevent StarClaw's daemon SSE client from silently dropping the final server-sent event when the stream ends without a trailing blank line.

This improves streaming/event reliability for agent-facing daemon integrations where the producer may close the connection immediately after writing the last event.

## Requirements

- `internal/client.SSEClient.readEvents` must emit a pending event after scanner EOF if it has any event type or data.
- Existing behavior for normal blank-line-delimited events must remain unchanged.
- Heartbeat/comment lines must still be ignored.
- Multi-line `data:` payloads must still be joined with newlines.
- Scanner errors must still be returned as wrapped errors.
- Add a regression test for a final event with no trailing blank line.

## Acceptance Criteria

- [ ] An SSE stream ending with `event: done\ndata: ok` produces one `SSEEvent`.
- [ ] Existing SSE client tests continue passing.
- [ ] `go test ./internal/client` passes.
- [ ] `go test ./...` passes.

## Notes

Out of scope:

- Last-Event-ID reconnect support.
- HTTP error body reporting.
- OpenAI/Anthropic stream parser changes.
