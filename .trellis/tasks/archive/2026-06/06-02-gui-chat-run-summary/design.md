# Design

## Frontend

Add a `renderRunSummary(result, payload)` helper in `app.js`.

The helper appends a compact transcript item after `renderDoneResult` for successful responses. It reads:

- `result.session_id`
- `payload.agent`
- `payload.request_id`
- `result.usage`

The summary is informational only. It does not change session state or message content.

## Backend

No backend changes. Existing `RunAgentResponse` already exposes `session_id`, `messages`, and `usage`.

## Compatibility

Existing chat behavior remains unchanged when `usage` is absent.
