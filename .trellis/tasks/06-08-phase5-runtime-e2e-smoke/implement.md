# Phase 5 runtime E2E smoke implementation plan

## Checklist

1. Load backend/Web UI specs with `trellis-before-dev`.
2. Inspect existing Web UI smoke `runRuns(page)` and daemon run-control tests.
3. Extend `runRuns(page)` mocked run detail with Phase 3/4 runtime metadata.
4. Mock `GET /runs/{id}/trace` and assert trace/recovery sections render.
5. Assert Mission Control recovered filter/card and runtime badges are present.
6. Add/adjust daemon tests only if API summary coverage is insufficient.
7. Run validation:
   - `go test ./internal/daemon -run 'TestWeb|TestRunHistoryAPI|TestRunsSummaryIncludesRuntimeRecoveryMetadata|TestHandleRunControl|TestHandleRunTrace' -count=1`
   - `go test ./internal/daemon ./cmd`
   - `go test ./...`
   - `git diff --check`
   - `scripts/smoke_webui_runs.sh` if local Playwright dependencies are available.
8. Update PRD acceptance criteria, commit, archive child, and record journal.

## Risk Files

- `scripts/lib/webui_smoke_common.sh`
- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/server_test.go`
- `internal/daemon/webui_test.go`

## Non-Goals

- No new runtime control semantics.
- No paid provider smoke dependency.
- No new browser smoke framework.
