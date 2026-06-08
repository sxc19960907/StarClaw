# Observability trace export design

## Scope

This child exports already-redacted `StructuredRunEvent` data to local JSONL. It does not add an external OpenTelemetry SDK, collector, cloud upload, or UI panel.

## Trace Record Shape

Each JSONL line is one event:

```json
{
  "schema_version": "2026-06-08",
  "trace_id": "run-id",
  "span_id": "run-id-000001",
  "parent_span_id": "",
  "run_id": "run-id",
  "event_id": "run-id-000001",
  "name": "run_started",
  "phase": "start",
  "timestamp": "2026-06-08T00:00:00Z",
  "attributes": {}
}
```

`attributes` comes from `StructuredRunEvent.Data`, which is already redacted. Export code should still sanitize recursively before writing.

## Store API

- `ExportTracesJSONL(path string) error`
  - writes all stored runs in store order.
- `ExportRunTraceJSONL(runID, path string) error`
  - writes one run.
- `TraceEvents(runID string) ([]TraceExportRecord, bool)`
  - returns one run's export records for API read/testing.

Write through a temp file in the destination directory and rename into place.

## HTTP API

- `GET /runs/{id}/trace`
  - Returns JSON `{ "trace": [...] }` for a single run.
- `GET /traces/export?path=/local/file.jsonl`
  - Writes all stored trace events to the requested local path.
  - Returns JSON `{ "path": "...", "events": n }`.

This keeps export explicit and local. If `path` is missing, return HTTP 400.

## Redaction

Never export:

- prompt/user text
- tool arguments
- provider request/response bodies
- raw run request/response
- API keys, auth tokens, secret-looking values

Export should rely on `StructuredRunEvent` plus an extra recursive sanitization pass.

## Compatibility

Existing `/metrics`, `/runs`, `/runs/{id}`, persistence, replay, pause/resume, and Web UI behavior remain unchanged.

## Rollback

Remove trace export record/types, store export methods, routes, and tests.
