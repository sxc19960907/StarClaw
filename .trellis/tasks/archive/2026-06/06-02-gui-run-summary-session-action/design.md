# Design

## Boundary

This task changes the daemon Web UI static assets and smoke test only:

- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

No daemon API contract changes are planned. The action consumes the existing `/message` response field `session_id` and existing `/sessions/{id}` selection path.

## UI Behavior

`renderRunSummary(result, payload)` will append an action row when `result.session_id` is non-empty. The row contains a single `Open session` button carrying the session id in a `data-run-summary-session` attribute.

The global click handler will handle that attribute and call the existing `selectSession(sessionID)` flow. This keeps active-state, panel switching, chat state labels, and transcript rendering consistent with sidebar session selection.

## Data Flow

`/message` response -> `renderRunSummary` -> `data-run-summary-session` -> click handler -> `selectSession` -> `/sessions/{id}` -> `renderSessionThread`

If `/sessions/{id}` fails, `selectSession` already renders an error fallback that keeps the selected session id for the next message. This is acceptable for mocked or stale session ids.

## Compatibility

Runs without `session_id` will continue to render the summary grid but omit the action row. Existing summary text and layout remain intact.

## Test Strategy

The Web UI smoke test will assert the `Open session` action appears for a mocked completed run with a session id. Existing real-session smoke coverage already verifies sidebar session selection and session management, so this task does not need a new backend round trip.
