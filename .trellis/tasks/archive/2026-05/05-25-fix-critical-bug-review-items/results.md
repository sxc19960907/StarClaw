# Results

## Fixed / Covered

- `internal/client/mock.go`
  - Preserved existing mutex protection.
  - Added defensive copies for stored and returned messages/tools.
  - Added concurrent-use and defensive-copy regression tests.

- `internal/agent/loop.go`
  - Current implementation already returns immediately after successful `StreamChat`.
  - Added regression coverage proving streaming success does not call fallback `Chat`.

- `internal/context/sanitize.go`
  - Current implementation already merges consecutive `assistant` / `user` content.
  - Strengthened tests to assert earlier content is preserved for both roles.

- `internal/context/window.go`
  - Current implementation already counts only `assistant -> user` turn pairs.
  - Added regression coverage to keep the two most recent turn pairs untruncated.

- `internal/daemon/checkpoint.go`
  - Strengthened ID sanitization for both `/` and `\` separators.
  - Added backslash traversal regression coverage.

- `internal/agent/watchdog.go`
  - While running race verification, `TestWatchdog_Reset_PreventsFire` exposed a timer-generation issue.
  - Added generation checks so old callbacks from prior starts/resets/stops cannot fire after state changes.

## Verification

- `go test ./internal/client ./internal/agent ./internal/context ./internal/daemon` passed.
- `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon` passed.
- `go test ./...` passed.
