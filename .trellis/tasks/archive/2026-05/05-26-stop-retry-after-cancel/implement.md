# Stop Retry After Cancel Implementation Plan

## Checklist

- [x] Add a regression test for cancellation during retry backoff.
- [x] Change `retryWait` to return an error and use a stoppable timer.
- [x] Stop `chatWithRetry` when `retryWait` returns context cancellation.
- [x] Run `gofmt`.
- [x] Run `go test ./internal/agent -run Retry`.
- [x] Run `go test ./...`.
- [x] Run `git diff --check`.
