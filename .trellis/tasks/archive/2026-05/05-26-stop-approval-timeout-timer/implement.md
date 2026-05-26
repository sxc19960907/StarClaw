# Stop Approval Timeout Timer Implementation Plan

## Checklist

- [x] Replace `time.After` in `ApprovalBroker.WaitForApproval` with a stopped/drained timer.
- [x] Run `gofmt`.
- [x] Run `go test ./internal/daemon -run ApprovalBroker`.
- [x] Run `go test ./...`.
- [x] Run `git diff --check`.
