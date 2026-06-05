# Design

## Architecture

This task stays inside the Web UI static frontend:

- `internal/daemon/webui/assets/app.js` owns rendering, state, and delegated click handling.
- `internal/daemon/webui/assets/styles.css` owns small layout adjustments for new actions.
- `scripts/lib/webui_smoke_common.sh` owns browser smoke assertions.

No daemon API shape changes are required. Existing run payloads already include `response`, prompt metadata, and timeline events.

## Data Flow

### Run Detail Result Copy

`state.currentRunDetail` -> `formatRunResponse(run.response)` -> clipboard via `copyText`.

The copied text must match the displayed Result block, so there is one formatting source of truth.

### Run Detail Tool Result Copy

`run.events` -> `groupRunTimelineEvents(events)` -> grouped tool entry -> `formatToolPayload(entry.result)` -> escaped into a `data-run-tool-copy-result` button attribute -> clipboard via delegated click handler.

Only render the button when `entry.result` is non-empty.

### Agent Test Live Timeline

`streamMessage(... renderer.appendEvent(eventType, data))` -> live event list in memory -> `groupRunTimelineEvents(events)` -> `renderRunTimelineEntry(entry)` -> live timeline container.

This keeps live Agent Test display aligned with persisted Run detail display.

## Compatibility

- Existing DOM selectors for current smoke tests should remain valid.
- Existing `run-tool-event` class remains the stable selector for grouped tool cards.
- New buttons are additive and should not change backend behavior.

## Trade-Offs

- Rendering live Agent Test events by rebuilding the timeline is simpler and consistent; event volume in smoke/user agent tests is small, so this is acceptable.
- Copying through `data-*` attributes follows existing UI patterns but requires escaping text carefully, using existing `escapeHTML`.

## Rollback

Revert changes in `app.js`, `styles.css`, and smoke assertions. Backend state and stored run/session data are unaffected.
