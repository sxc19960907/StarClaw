# Runtime recovery UI implementation plan

## Checklist

1. Read Trellis backend/Web UI specs before editing.
2. Inspect Mission Control list/detail rendering and existing run control/replay/trace APIs.
3. Extend Web UI state with selected run trace/error.
4. Load trace data when selecting a run, with graceful fallback.
5. Add derived helpers for recovered/replay/pause/trace badges.
6. Update Mission Control cards and filters.
7. Update run rows with runtime badges.
8. Update run detail with recovery, workflow steps, control history, and trace sections.
9. Add CSS for badges, recovery cards, step/control rows, and trace summaries.
10. Add or update tests for embedded UI asset content and trace/recovery API compatibility.
11. Run validation:
    - `gofmt -w internal/daemon/*.go`
    - `go test ./internal/daemon -run 'TestWeb|TestHandleRunTrace|TestHandleRunControl|TestPersistentRunStore' -count=1`
    - `go test ./internal/daemon ./cmd`
    - `go test ./...`
    - `git diff --check`
12. Update acceptance criteria, commit, archive child task, and record journal.

## Risk Files

- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/webui_test.go`
- `.trellis/spec/backend/quality-guidelines.md`

## Non-Goals

- No runtime semantic changes.
- No backend trace schema changes unless tests reveal a compatibility gap.
- No prompt archival/export feature.
