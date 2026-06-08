# Sync marker dry-run foundation design

## Architecture

Add a new `internal/sync` package with the local-only primitives needed before any session sync runtime exists:

- `config.go`: sync configuration defaults and `LoadConfig`.
- `marker.go`: marker schema, read reset/sidecar behavior, and atomic writes.
- `lock.go`: lock acquisition for `.starclaw/sync.lock` with timeout.
- `dryrun.go`: dry-run outbox batch writer.

No daemon, CLI, cloud client, or session scanner will call this package in this task.

## Config

Add `Sync SyncConfig` to `internal/config.Config` and default values to both `Load()` and `defaultConfig()` paths.

Default values:

- `enabled=false`
- `dry_run=false`
- `endpoint=""`
- `batch_max_sessions=25`
- `batch_max_bytes=5242880`
- `single_session_max_bytes=4194304`
- `daemon_interval=24h`
- `daemon_startup_delay=60s`
- `failed_max_attempts_transient=5`
- `lock_timeout=30s`

This makes the config shape explicit while preserving local-first behavior.

## Marker

Marker path convention for future callers:

- `.starclaw/sync_marker.json`

Marker fields:

- version
- last sync time/count/outcome
- failed entries keyed by session id

Read behavior:

- missing file: return empty marker, nil error;
- corrupt JSON or missing/zero version: sidecar `.corrupt.bak`, return empty marker;
- unknown version: sidecar `.unknown-v<N>.bak`, return empty marker;
- valid marker with nil failed map: normalize to empty map.

Write behavior:

- fill default version and failed map;
- create parent dirs;
- write temp file in same dir;
- chmod/write as `0600`;
- rename atomically.

## Lock

Use an exclusive lock file at `.starclaw/sync.lock`.

Expose:

- `WithLock(ctx, lockPath, timeout, fn)` for future run orchestration.
- `ErrLockContention` sentinel when timeout/context prevents acquiring the lock.

Implementation should poll `internal/filelock.Exclusive` via a lock file. On platforms where filelock is a no-op, tests still exercise the API shape; Unix tests cover real contention.

## Dry-Run Outbox

Define a StarClaw-local batch shape instead of importing a future cloud client protocol:

- `BatchRequest`
- `SessionEnvelope`
- `BatchResponse`

`DryRunUploader.Send(ctx, batch)` writes a JSON outbox file under a configured directory and returns all session ids in `Accepted`. It must not make network calls.

## Privacy

This task does not inspect session bodies or strip thinking yet. The dry-run writer persists exactly the batch passed to it. The next task, `session-sync-batcher-privacy`, will own transformation and redaction before batches reach this writer.
