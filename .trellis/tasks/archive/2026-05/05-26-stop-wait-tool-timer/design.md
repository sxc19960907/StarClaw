# Stop Wait Tool Timer Design

## Boundary

The runtime change is limited to `internal/tools/wait.go`. Existing tests in `internal/tools/wait_test.go` cover cancellation and duration behavior.

## Behavior

`WaitTool.Run` should create a `time.NewTimer(duration)` and defer timer cleanup. The select should keep the same two outcomes:

- timer fires: return the existing success message
- context cancels first: return `wait cancelled` as an error tool result

Timer cleanup should use `Stop` and a non-blocking drain to avoid blocking if the timer already fired.
