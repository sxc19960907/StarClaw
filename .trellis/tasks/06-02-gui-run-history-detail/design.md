# Design

## Backend Boundary

Add an in-memory run store to `internal/daemon`:

- `RunRecord`: summary/detail fields.
- `RunEvent`: type, timestamp, and JSON-compatible data.
- Store keeps recent runs only, bounded to a fixed size.

Routes:

- `GET /runs` returns summaries ordered newest first.
- `GET /runs/{id}` returns full detail.

`handleMessage` will create a run record after request normalization. It wraps the current event handler with a recorder handler so synchronous and SSE paths both capture the same event stream. Completion records status, response, session id, usage, error, and end time.

Approval events are already published via `EventBus`, but they are not passed through `agent.EventHandler`. In this iteration, run detail will capture tool events from the run handler and approval status through existing GUI live approval cards; if approval request correlation is available through the request id, add it to run events opportunistically. Full persisted approval audit is out of scope.

## Frontend Boundary

Add a `Runs` nav item and panel:

- list pane: recent runs
- detail pane: selected run metadata and timeline

Existing run summary rendering will include `Open run` when `result.request_id` or payload request id is available. The action switches to the Runs panel and loads `/runs/{request_id}`.

## API Contract

Run summary shape:

```json
{
  "id": "request-id",
  "status": "completed",
  "agent": "helper",
  "session_id": "session-id",
  "started_at": "RFC3339",
  "ended_at": "RFC3339",
  "prompt": "preview"
}
```

Run detail adds:

```json
{
  "request": { "...": "..." },
  "response": { "...": "..." },
  "usage": { "...": 1 },
  "events": [
    { "type": "tool_call", "at": "RFC3339", "data": { "...": "..." } }
  ]
}
```

## Test Strategy

- Unit/API test for `/runs` and `/runs/{id}` after a successful `/message`.
- Web UI smoke:
  - mocked summary path can verify `Open run` button visibility if route response includes request id,
  - real direct message path can verify Runs list/detail after refresh.

## Rollout

In-memory history avoids schema/file migration. If later persistence is needed, this store can be backed by disk without changing the frontend contract.
