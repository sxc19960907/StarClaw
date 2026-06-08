# Session sync batcher privacy implementation plan

## Steps

1. Add `internal/sync/scanner.go` and tests for StarClaw session layouts, watermark filtering, excludes, transient retry, permanent no-churn, and skipped invalid files.
2. Add `internal/sync/strip_thinking.go` and tests for assistant thinking block stripping.
3. Add `internal/sync/batcher.go` and tests for load errors, size failures, privacy-before-size, batch limits, and context cancellation.
4. Run `gofmt`.
5. Run:

```bash
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-session-sync-batcher-privacy
go test ./internal/sync ./internal/session ./internal/daemon
go test ./...
```

6. Search `internal/sync` for cloud/network imports.

## Risk Controls

- Do not add a cloud uploader or call `DryRunUploader`.
- Do not mutate marker watermarks in this task.
- Keep candidate discovery tolerant of unreadable/corrupt session files.
- Keep privacy transformation structural and covered by tests.
