# Stop SSE Reconnect Timer Implementation Plan

## Checklist

- [x] Add a test for cancellation during a long reconnect delay.
- [x] Replace `time.After(delay)` with stoppable timer logic.
- [x] Run `gofmt`.
- [x] Run `go test ./internal/client`.
- [x] Run `go test ./...`.
- [x] Run `git diff --check`.
