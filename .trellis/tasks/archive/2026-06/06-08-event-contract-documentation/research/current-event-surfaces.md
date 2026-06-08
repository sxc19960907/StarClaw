# Current event surface evidence

## Backend

- `internal/daemon/events.go` defines EventBus event names and replay history.
- `internal/daemon/server.go` implements `/events` with `last_event_id`,
  `Last-Event-ID`, replay, live stream, and keepalive comments.
- `internal/daemon/server.go` implements per-request `/message` SSE through
  `sseEventHandler`, including Kocoro-compatible aliases.
- `internal/daemon/run_store.go` publishes run lifecycle EventBus events:
  `run_started`, `run_completed`, and `run_error`.

## Web UI

- `internal/daemon/webui/assets/app.js` consumes approval lifecycle events and
  run lifecycle events from `/events`.
- `internal/daemon/webui/assets/app.js` consumes `delta`, `assistant_text`,
  `session_started`, `tool`, and legacy SSE names from `/message`.

## Decision

Create one canonical documentation page under `docs/DAEMON_EVENTS.md`, linked
from README. This should document current local behavior and explicitly avoid
claiming Kocoro/Shannon Cloud IM lifecycle support.
