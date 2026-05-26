# Flush Final SSE Event Implementation Plan

## Checklist

- [x] Add a focused regression test in `internal/client/sse_test.go`.
- [x] Flush pending `current` after EOF in `internal/client/sse.go`.
- [x] Run `gofmt`.
- [x] Run `go test ./internal/client`.
- [x] Run `go test ./...`.
- [x] Run `git diff --check`.

## Rollback

Revert the small `readEvents` EOF flush and the corresponding test.
