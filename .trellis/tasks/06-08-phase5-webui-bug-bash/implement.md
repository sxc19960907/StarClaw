# Phase 5 Web UI bug bash implementation plan

## Checklist

1. Load backend specs and embedded UI context with `trellis-before-dev`.
2. Inspect Web UI runtime functions and current static smoke tests.
3. Identify scoped integration defects around runtime recovery, trace, budget, quality/reuse/share/memory navigation, and empty/error states.
4. Patch static assets conservatively.
5. Add/extend `internal/daemon/webui_test.go` assertions for changed hooks and render guards.
6. Run validation:
   - Targeted `go test ./internal/daemon -run 'TestWebUI' -count=1`
   - `go test ./internal/daemon ./cmd`
   - `go test ./...`
   - `git diff --check`
7. Update PRD acceptance criteria, commit, archive child, and record journal.

## Risk Files

- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `internal/daemon/webui/index.html`
- `internal/daemon/webui_test.go`

## Non-Goals

- No broad visual redesign.
- No frontend build pipeline.
- No new backend runtime behavior unless required to fix a real UI integration defect.
