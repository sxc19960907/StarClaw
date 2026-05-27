# Implementation

## Steps

- [x] Replace the initial alignment `time.After` call with `time.NewTimer`.
- [x] Stop and non-blocking drain the alignment timer on context cancellation.
- [x] Review existing scheduler cancellation tests and add or adjust coverage if useful.
- [x] Run daemon scheduler tests.
- [x] Run full repository tests.
- [x] Run whitespace check.

## Validation

```bash
go test ./internal/daemon -run Scheduler
go test ./...
git diff --check
```
