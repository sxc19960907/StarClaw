# Keychain store boundary design

## Architecture

Add a new `internal/keychain` package with two layers:

- `Backend`: low-level credential store contract with `Read`, `Write`, and `Delete`.
- `Store`: StarClaw domain wrapper that knows the service/account constants and active-user behavior.

Production code may call `NewOSStore` in a future task, but this task does not wire it into config, daemon, cloud, or UI paths.

## Constants

Use StarClaw names, not Kocoro names:

- `ServiceDaemonAPIKey = "ai.starclaw.daemon.api_key"`
- `ServiceDaemonState = "ai.starclaw.daemon.state"`
- `AccountCurrentUser = "current_user_id"`
- `AccountLegacy = "legacy"`

## Platform Behavior

- darwin: `NewOSStore` returns a `Store` backed by OS keychain.
- non-darwin: `NewOSStore` returns `ErrUnsupportedPlatform`.
- tests use `NewMemBackend`.

The darwin backend should not be exercised in unit tests because it may prompt the OS Keychain. Unit tests should target the memory backend and compile the OS backend.

## Store Contract

- `Read(service, account)` returns empty string and nil error when the backend reports `ErrNotFound`.
- `Delete(service, account)` is idempotent.
- `SetAPIKey(userID, key)` requires both values and writes key before active-user pointer.
- `DeleteAPIKey()` removes the current user's API key but preserves `current_user_id`.
- `ClearActiveUser()` removes only the active-user pointer.
- `RenameLegacy(realUserID)` moves `AccountLegacy` to the real user and removes the legacy key when present.

## Privacy

The package stores and returns raw secret values to direct callers, but it must not log them. This task adds no API surface that serializes those values.

## Compatibility

No existing config behavior changes. API keys continue to load from existing config/env sources until a future explicit opt-in task wires keychain into credential resolution.
