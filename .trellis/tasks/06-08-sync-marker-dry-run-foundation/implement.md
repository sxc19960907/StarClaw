# Sync marker dry-run foundation implementation plan

## Steps

1. Add `internal/sync/config.go`.
2. Add `SyncConfig` to `internal/config/config.go` and `internal/config/multilevel.go`.
3. Add config tests for defaults and overlay behavior.
4. Add `internal/sync/marker.go` and marker tests.
5. Add `internal/sync/lock.go` and lock tests.
6. Add `internal/sync/dryrun.go` and dry-run outbox tests.
7. Run `gofmt`.
8. Run:

```bash
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-sync-marker-dry-run-foundation
go test ./internal/sync ./internal/config
```

9. Search to confirm no cloud uploader/network client was added under `internal/sync`.

## Risk Controls

- Keep sync disabled by default.
- Do not add daemon startup calls or background workers.
- Do not import `internal/client` from `internal/sync` in this task.
- Keep dry-run output local and file-per-batch.
