# Results

## Fixed / Covered

- `internal/tools/process.go`
  - Replaced shared `bytes.Buffer` stdout/stderr capture with a mutex-protected process buffer.
  - `start` and `status` can now read process output while exec copy goroutines are still writing.
  - Added a regression test that polls status concurrently while a spawned shell process writes output.

- `internal/heartbeat/heartbeat.go`
  - `Close()` before `Start()` now returns immediately.
  - Duplicate `Start()` calls are no-ops while already running.
  - `Start()` / `Close()` lifecycle state is guarded with `running`, `done`, and `cancel` under the manager mutex.
  - Added lifecycle tests for Close-before-Start and concurrent Start/Close calls.

- `internal/agent/registry.go` and `internal/agent/readtracker.go`
  - Existing local code already had mutex protection for the reported races.
  - Added concurrent-access regression tests so the behavior is covered by race testing.

- `internal/agent/watchdog_test.go`
  - Widened a reset timing assertion that was flaky under `-race`; the watchdog generation fix remains covered.

## Verification

- `go test ./internal/tools ./internal/agent ./internal/heartbeat` passed.
- `go test -race ./internal/tools ./internal/agent ./internal/heartbeat` passed.
- `go test ./...` passed.
