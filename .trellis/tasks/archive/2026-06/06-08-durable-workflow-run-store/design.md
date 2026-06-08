# Durable workflow run store design

## Scope

This slice makes `RunStore` optionally durable. It does not change how runs execute; it only makes run metadata recoverable after daemon restart.

## Storage

- Add an optional persistence file to `RunStore`, e.g. `runs.json`.
- Persist a compact envelope:
  ```json
  {
    "schema_version": "2026-06-08",
    "records": [ { "id": "...", "status": "completed" } ]
  }
  ```
- Keep using JSON rather than adding SQLite or another dependency.
- Write through a temp file and rename into place.

## Data Flow

1. `NewRunStore(limit)` remains available and returns the existing in-memory behavior.
2. Add `NewPersistentRunStore(limit, path)` or equivalent constructor.
3. Persistent constructor loads existing records, clamps to store limit, rebuilds order and event sequence counters.
4. Mutations (`Start`, `Complete`, `AddEvent`, `AddControlDecision`) call a best-effort persistence hook after in-memory state updates.
5. Daemon server construction can keep in-memory store for tests unless a data path is available; this child can prove durability at `RunStore` level first.

## Error Handling

- Load errors are returned from the persistent constructor for tests/callers that want to surface them.
- Corrupt JSON should not panic. A helper can return an empty store plus error.
- Persist write failures should be returned from explicit persistence helpers where possible; mutation APIs remain compatible and should not change signatures in this slice.

## Compatibility

- Existing JSON fields on `RunRecord` remain unchanged.
- `/runs`, `/runs/{id}`, `/metrics`, `/cancel`, `/runs/{id}/control` behavior remains compatible.
- Existing `NewRunStore` tests keep using in-memory mode.

## Rollback

Remove the persistent constructor and persistence hook; in-memory behavior remains unchanged.
