# Stop retry after cancel

## Goal

Make LLM retry backoff stop promptly and abort the retry loop when the run context is cancelled.

This prevents StarClaw from issuing another LLM call with an already-cancelled context after a transient error enters retry backoff.

## Requirements

- `AgentLoop.retryWait` must use a stoppable timer instead of `time.After`.
- `AgentLoop.retryWait` must report context cancellation to its caller.
- `chatWithRetry` must return promptly when retry backoff is cancelled instead of continuing to the next attempt.
- Preserve successful retry behavior for transient errors.
- Preserve non-retry behavior for context cancellation/deadline errors.

## Acceptance Criteria

- [ ] A test demonstrates cancellation during retry backoff stops after the first failed attempt.
- [ ] Existing retry success/exhaustion tests continue passing.
- [ ] `go test ./internal/agent -run Retry` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Notes

Out of scope:

- Changing retry counts or jitter policy.
- Adding user-configurable retry settings.
