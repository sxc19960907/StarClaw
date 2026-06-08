# SSE reconnect idle watchdog

## Goal

Bring StarClaw's local SSE client/server behavior closer to Kocoro's reconnectable stream contract by adding explicit idle timeout, bounded reconnect, and `Last-Event-ID` replay support.

## Requirements

- Keep behavior local-first. Do not add cloud credentials, external channel delivery, or telemetry.
- Preserve existing `SSEClient.Connect(ctx, url)` behavior for existing callers.
- Add an explicit options-based SSE connection API for:
  - idle timeout detection,
  - bounded reconnect attempts,
  - configurable reconnect backoff for tests and callers,
  - `Last-Event-ID` on reconnect after an event with an ID was observed.
- Distinguish terminal cases:
  - context cancellation closes the channel promptly,
  - clean EOF closes normally,
  - idle timeout reconnects only while reconnect budget remains,
  - exhausted reconnect budget closes the channel.
- Ensure `/events` can replay missed events from a provided event ID.

## Non-Goals

- Do not change OpenAI-compatible streaming chunk shapes.
- Do not add real cloud Gateway or external transport behavior.
- Do not redesign daemon event persistence beyond the existing in-memory event bus.

## Acceptance Criteria

- [x] SSE client tests cover `Last-Event-ID` on reconnect.
- [x] SSE client tests cover idle timeout reconnect and exhausted reconnect budget.
- [x] SSE client tests cover cancellation during reconnect delay.
- [x] Daemon `/events` tests cover replay from query or `Last-Event-ID` header if existing behavior is incomplete.
- [x] Existing `SSEClient.Connect` tests continue to pass.
- [x] `go test ./internal/client` passes.
- [x] `go test ./internal/daemon` passes.
- [x] `go test ./...` passes or any unrelated failure is documented.
