# Astria UI Polish Design

## Architecture Boundary

This child changes only the embedded Web UI assets under `internal/daemon/webui/` and smoke coverage under `scripts/lib/` unless a small test adjustment is required. The UI remains static HTML/CSS/JS served by the daemon.

## Design Direction

- Keep the sidebar and central mission composer as the product anchor.
- Add Inbox to Home docked tools so the channel workflow is visible from the first screen.
- Improve repeated panel empty states with concise action buttons that route to the relevant flow.
- Prefer small, useful visual refinements over broad restyling.
- Keep celestial identity subtle: rings, star marks, light gradients, restrained color.

## Data Flow

- Existing `refreshAll()` remains the source of counts.
- Existing state arrays (`runs`, `agents`, `skills`, `schedules`, `councilRuns`, `inboxItems`, `memory`) drive Home/Manage counts.
- No new daemon routes are required.

## Compatibility

- Existing smoke selectors should continue working.
- Existing panel IDs, form IDs, and API paths remain unchanged.
- No new dependency or build step.

## Risk

- Highest risk is breaking a broad static JS file with many listeners. Mitigate with `node --check` and Web UI smoke.
- CSS changes can affect many panels. Keep edits scoped and visually inspect smoke screenshots.
