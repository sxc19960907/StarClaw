# Stop Wait Tool Timer Implementation Plan

## Checklist

- [x] Replace `time.After` in `WaitTool.Run` with a stopped/drained timer.
- [x] Run `gofmt`.
- [x] Run `go test ./internal/tools -run TestWaitTool`.
- [x] Run `go test ./...`.
- [x] Run `git diff --check`.
