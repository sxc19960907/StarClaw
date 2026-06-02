# Design

## Backend

Extend `providerConfigPatch` with:

```go
Permissions *permissions.Config `json:"permissions"`
```

`providerConfigPatch.apply` will copy and normalize permission lists:

- trim whitespace
- drop empty entries
- preserve order
- if all lists are empty, set `cfg.Permissions = nil`
- otherwise set `cfg.Permissions` to the cleaned config

The existing `handlePatchConfig` flow already reads YAML, applies the patch, marshals YAML, writes `0600`, and refreshes `s.deps.Config`; reuse it.

## Frontend

The current Permissions panel is read-only. Replace the list-only rendering with an editor form:

- one textarea per permission list
- newline-separated values
- save button sends `{ permissions: ... }` to `PATCH /config`
- clear button stages empty lists and saves
- overview/list remains visible as a preview of currently loaded/saved values

After save:

1. call `loadPermissions()`
2. call `loadDiagnostics()`
3. show toast feedback

## API Contract

Request:

```json
{
  "permissions": {
    "allowed_dirs": ["~", "."],
    "allowed_commands": ["go test"],
    "denied_commands": ["shutdown"],
    "network_allowlist": ["api.github.com"],
    "sensitive_patterns": ["*.secret"]
  }
}
```

Response uses the existing `/config` response shape.

## Test Strategy

- Go test:
  - patch permissions
  - assert YAML persisted
  - assert `deps.Config.Permissions` refreshed
  - assert `GET /permissions` returns updated lists
  - assert empty patch clears permissions to defaults/unconfigured
- Web UI smoke:
  - edit permissions fields
  - save
  - verify overview/list values reload
  - clear and verify default state

## Rollback

Revert config patch permissions support and Web UI permissions form changes if persistence or diagnostics behavior regresses.
