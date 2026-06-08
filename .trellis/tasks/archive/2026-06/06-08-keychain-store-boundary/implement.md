# Keychain store boundary implementation plan

## Steps

1. Add `internal/keychain/keychain.go` with errors, constants, backend interface, and store methods.
2. Add `internal/keychain/backend_mem.go` with concurrency-safe in-memory backend.
3. Add `internal/keychain/backend_darwin.go` using an OS keychain dependency.
4. Add `internal/keychain/backend_other.go` for unsupported platforms.
5. Add `internal/keychain/keychain_test.go` covering store behavior and memory backend concurrency.
6. Run `gofmt`.
7. Run validation:

```bash
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-keychain-store-boundary
go test ./internal/keychain ./internal/config ./internal/daemon
```

8. Confirm no config/daemon automatic keychain wiring was added.

## Risk Controls

- Keep edits scoped to `internal/keychain` and Trellis artifacts, except `go.mod`/`go.sum` if the OS backend dependency is needed.
- Do not call `NewOSStore` from existing startup/config paths.
- Do not write tests that touch the real OS Keychain.
