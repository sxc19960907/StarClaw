# Flush Final SSE Event Design

## Boundary

The change is confined to `internal/client/sse.go` and its tests. The public API stays the same.

## Behavior

`readEvents` already accumulates fields into a `current SSEEvent` and sends it on blank-line boundaries. After `scanner.Scan()` ends and `scanner.Err()` is nil, the method should send `current` once more if it contains an event type or data.

This mirrors the existing boundary behavior without changing parsing rules.

## Validation

Add a regression test using an `httptest.Server` that writes an SSE event without a final blank line. The channel should close after the server returns, and the collected events should include the final event.
