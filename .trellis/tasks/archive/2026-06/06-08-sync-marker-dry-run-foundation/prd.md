# Sync marker dry-run foundation

## Goal

Add StarClaw's local-only sync foundation: disabled-by-default config, marker persistence, lock serialization, and dry-run outbox writing. This creates the safe base for later session discovery/batching without adding any cloud uploader or off-machine data movement.

## Requirements

- Add `internal/sync` with:
  - `Config` and default loader for `sync.*` settings;
  - `Marker` schema with atomic read/write;
  - corrupt and unknown-version sidecar behavior;
  - lock helper around `.starclaw/sync.lock` with timeout and contention detection;
  - dry-run batch/outbox writer that writes local JSON files with `0600` permissions and returns accepted ids.
- Add `config.SyncConfig` to StarClaw config loading with local-first defaults:
  - `sync.enabled=false`
  - `sync.dry_run=false`
  - no endpoint by default.
- Keep this task local-only:
  - no cloud uploader;
  - no session scanner;
  - no daemon/CLI job runner;
  - no automatic background sync.
- Reuse existing file locking style where possible.

## Acceptance Criteria

- [ ] Missing marker file returns an empty current-version marker.
- [ ] Corrupt marker file sidecars to `.corrupt.bak` and resets safely.
- [ ] Unknown marker version sidecars to `.unknown-v<N>.bak` and resets safely.
- [ ] Marker writes are atomic and create parent directories.
- [ ] Lock contention returns a typed contention result/error and does not look like a fatal data failure.
- [ ] Dry-run outbox writes valid JSON with `0600` permissions and accepts all ids present in the batch.
- [ ] Config defaults keep sync disabled.
- [ ] No network/client/cloud uploader code is added.
- [ ] `go test ./internal/sync ./internal/config` passes.

## Out of Scope

- Session candidate scanning.
- Thinking/redacted-thinking stripping.
- Cloud upload protocol.
- Daemon scheduler or CLI command.
- UI controls.
