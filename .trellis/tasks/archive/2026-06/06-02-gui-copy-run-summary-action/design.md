# Design

## Boundary

This task changes only Web UI static assets and smoke coverage:

- `internal/daemon/webui/assets/app.js`
- `internal/daemon/webui/assets/styles.css`
- `scripts/smoke_webui.sh`

No daemon API change is required. The copied text is derived from the existing `/message` result and submitted payload.

## UI Behavior

`renderRunSummary(result, payload)` will render a `Copy summary` button for every successful summary. If `session_id` exists, the existing `Open session` button remains beside it.

The button stores a plain-text summary in a `data-run-summary-copy` attribute. The delegated click handler copies that value through `navigator.clipboard.writeText` and shows a toast.

## Copied Format

The copied text uses one field per line:

```text
Session: <session id or ->
Agent: <agent or default>
Usage: <usage values or ->
Request: <request id or ->
```

This format is intentionally plain text so it works in terminals, docs, and issue trackers.

## Compatibility

Clipboard failures are surfaced through the existing toast path. The task does not add a fallback textarea copy path because the Web UI smoke environment and modern browsers support `navigator.clipboard` in this context.

## Test Strategy

The Web UI smoke test will click `Copy summary`, verify the toast, and read clipboard text to confirm key fields are copied.
