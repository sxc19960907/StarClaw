# Phase 5 runtime E2E smoke design

## Scope

This child validates the main runtime product path with deterministic local fixtures. It should extend existing daemon/Web UI smoke coverage rather than introduce a second browser test framework.

## Existing Assets

- `scripts/smoke_webui_runs.sh` runs the `runs` mode in `scripts/lib/webui_smoke_common.sh`.
- `runRuns(page)` already covers run creation, run history, Mission Control filters, run detail timeline, session actions, copy actions, rerun, and error run display.
- Recent Phase 4 UI added runtime recovery, workflow steps, control history, and trace sections.

## Implementation Shape

1. Extend the mocked run detail in `runRuns(page)` to include:
   - `budget_status`
   - `routing`
   - `fallback`
   - `control` entries for replay approval and pause/resume
   - `steps` entries for replay approval and runtime pause
   - `structured_events`/`trace_events` summary count where needed
2. Add a route mock for `GET /runs/{id}/trace` for the mocked run.
3. Add browser assertions for:
   - Mission Control `Recovered` card/filter availability.
   - row badges for replay approval, pause/resume, and trace count.
   - Run detail `Runtime Recovery`, `Workflow Steps`, `Control History`, and `Trace` sections.
   - budget/routing/fallback JSON remains renderable in detail/timeline surfaces when present.
4. Add a targeted daemon test if an API summary field is missing or ambiguous.

## Compatibility

The smoke script remains optional/manual; normal `go test ./...` should still pass without Playwright. Static asset route tests should continue to cover hook presence.

## Rollback

Revert the smoke assertions and any narrow API/test additions. Runtime behavior should remain unchanged.
