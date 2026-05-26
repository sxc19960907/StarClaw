# Stop SSE reconnect timer

## Goal

Make SSE reconnect backoff cancellation release timer resources promptly instead of leaving a `time.After` timer alive until the full delay expires.

This improves daemon streaming reliability when an agent client is cancelled during retry backoff.

## Requirements

- Replace `time.After(delay)` in `SSEClient.run` retry backoff with a stoppable `time.Timer`.
- Preserve existing reconnect behavior and exponential backoff.
- Preserve immediate return when context is cancelled.
- Add test coverage proving cancellation during a long reconnect delay closes the event channel promptly.

## Acceptance Criteria

- [ ] `SSEClient.run` stops the retry timer when cancellation wins the select.
- [ ] Existing SSE behavior tests continue passing.
- [ ] New cancellation/backoff test passes without waiting for the full reconnect delay.
- [ ] `go test ./internal/client` passes.
- [ ] `go test ./...` passes.

## Notes

Out of scope:

- Changing reconnect retry policy.
- Adding Last-Event-ID support.
- Changing HTTP error reporting.
