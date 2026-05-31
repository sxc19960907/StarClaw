# Implementation Plan

## Checklist

- [x] Start the Trellis task after planning artifacts are complete.
- [x] Add optional config field to `ServerDeps` and wire it from `cmd daemon start`.
- [x] Derive an effective config in `RunAgent`, including named-agent merge.
- [x] Derive per-run registry from `deps.Registry` using effective tool filters.
- [x] Configure loop max iterations, max tokens, result truncation, config dir, context window, thinking options, and specific model.
- [x] Inject named-agent prompt and memory using existing loop APIs.
- [x] Add capture-client tests for daemon named-agent runtime parity.
- [x] Run targeted tests.
- [x] Run full `go test ./...` and `go vet ./...`.

## Validation Commands

```bash
go test ./internal/daemon ./cmd
go test ./...
go vet ./...
```

## Rollback Points

- Revert `internal/daemon/runner.go`, `internal/daemon/types.go`, `cmd/daemon.go`, and related tests if the dependency shape conflicts with planned daemon config ownership.
