# Keychain store boundary

## Goal

Add StarClaw's local OS keychain storage boundary so future credential flows can use an explicit, testable abstraction without silently migrating existing config secrets or enabling cloud behavior.

## Requirements

- Add `internal/keychain` with:
  - a low-level `Backend` interface;
  - high-level `Store` methods for raw read/write/delete, active user lookup, API key set/get/delete, active user clear, and legacy rename;
  - StarClaw-specific service/account constants;
  - `ErrUnsupportedPlatform` and `ErrNotFound` sentinel errors;
  - a concurrency-safe memory backend for tests;
  - a darwin OS backend;
  - a non-darwin unsupported backend.
- Keep the package independent from daemon/config by default.
- Do not automatically migrate `config.yaml`, env vars, or cloud credentials into keychain.
- Do not add daemon endpoints or UI controls in this task.
- Do not log or expose stored secret values.
- Preserve local-first defaults.

## Acceptance Criteria

- [ ] Missing entries read as empty through `Store.Read`.
- [ ] Non-darwin `NewOSStore` returns `ErrUnsupportedPlatform`.
- [ ] Memory backend covers set/get/delete, active user, delete-preserves-user, clear-active-user, legacy rename, nil store/backend behavior, and concurrent access.
- [ ] Darwin OS backend compiles and maps missing entries to `ErrNotFound`.
- [ ] No config or daemon endpoint stores to keychain automatically.
- [ ] `go test ./internal/keychain ./internal/config ./internal/daemon` passes.

## Out of Scope

- Config-to-keychain migration.
- Cloud authentication changes.
- Sync or migration implementation.
- UI or daemon keychain management endpoints.
