# Daemon Event Contracts

StarClaw's daemon exposes two local SSE event surfaces:

- `GET /events` is the daemon-wide EventBus stream used by Astria and local
  integrations that need replayable lifecycle events.
- `POST /message` with streaming enabled is a per-request run stream used for
  one active run response.

Both surfaces are local-first. They do not enable Shannon Cloud transport,
off-machine telemetry, or Kocoro IM `MESSAGE_LIFECYCLE` events by default.

## `GET /events`

`/events` returns `text/event-stream` and sends EventBus events as SSE blocks:

```text
id: 42
event: run_completed
data: {"run_id":"req-123","status":"completed"}
```

### Replay Cursor

Clients can request missed events with either cursor form:

- Query parameter: `GET /events?last_event_id=41`
- Header: `Last-Event-ID: 41`

The query parameter wins when both are present. Event IDs are monotonically
assigned by the in-memory EventBus. The replay buffer is bounded, so very old
events may no longer be available after the ring wraps.

`SubscribeWithReplay` registers the live subscriber and computes missed events
under one lock. Replayed events are written first, then live events are
streamed. Invalid cursors replay the currently buffered history.

The stream sends keepalive comments periodically:

```text
: keepalive
```

Slow subscribers may miss live events when their channel buffer is full. Those
events are still eligible for bounded replay if they remain in EventBus
history.

### EventBus Event Names

| Event | Purpose |
|---|---|
| `approval_needed` | A tool or replay action needs local operator approval. |
| `approval_resolved` | A pending approval was allowed or denied. |
| `run_started` | A run record entered `running`. |
| `run_completed` | A run finished successfully. |
| `run_error` | A run finished with an error. |
| `run_status` | Runtime status such as watchdog, retry, or budget stop detail. |
| `tool_status` | Tool progress summarized for daemon-wide clients. |
| `preamble` | Assistant preamble/thinking summary where available. |
| `stream_delta` | Assistant streaming delta for daemon-wide consumers. |
| `usage` | Token usage counts. |
| `budget_status` | Runtime token budget status. |
| `cloud_delegate_start` | Local cloud-delegation boundary marker. |
| `cloud_delegate_progress` | Local cloud-delegation progress marker. |
| `cloud_delegate_complete` | Local cloud-delegation completion marker. |
| `error` | Generic daemon runtime error event. |

Run lifecycle payloads are recovery-oriented summaries. They may include safe
fields such as `run_id`, `status`, `agent`, `channel`, `source`,
`session_id`, `started_at`, `ended_at`, `usage`, `budget_status`, `routing`,
`fallback`, and a redacted `error`.

## `POST /message` Streaming SSE

The per-request `/message` stream is scoped to one run. It uses the normal
daemon path: permissions, approvals, sessions, run store, budget/routing,
fallback, and structured events.

The stream can emit both StarClaw legacy names and Kocoro-compatible aliases so
older local clients keep working while newer clients can consume the common
vocabulary.

| Event | Status |
|---|---|
| `session_started` | Kocoro-compatible session metadata; includes `session_id`. |
| `tool_call` | Legacy tool start event with tool name and args. |
| `tool_result` | Legacy tool result event with content/preview and error state. |
| `tool` | Kocoro-compatible tool progress alias. |
| `text` | Legacy assistant text or streaming text chunk. |
| `delta` | Kocoro-compatible streaming text alias. |
| `preamble` | Legacy assistant preamble. |
| `assistant_text` | Kocoro-compatible assistant preamble/text alias. |
| `usage` | Per-run token usage counts. |
| `done` | Final run response payload. |
| `error` | Terminal stream error payload. |

Clients should de-duplicate alias pairs when they consume both legacy and
compatible names. Astria suppresses duplicate text when `text`/`delta` or
`preamble`/`assistant_text` carry the same content.

## Privacy And Redaction

EventBus lifecycle payloads and structured observability events are redacted
before publication or export. Fields named `args`, `content`, `text`, `delta`,
`preamble`, `prompt`, `request`, or `response` are not treated as shareable
event attributes. Values or keys that look like API keys, bearer tokens,
secrets, or passwords are replaced with `[REDACTED]`.

Run detail pages may still expose prompt/result panels for local operator
review. Do not treat those local detail views as support bundles or telemetry
payloads.

Metrics are aggregate-only. Trace export is explicit, caller-directed, and
local JSONL only.

## Kocoro Compatibility Boundary

StarClaw intentionally aligns with Kocoro where it improves local recovery:

- SSE aliases such as `session_started`, `delta`, `assistant_text`, `tool`,
  and `usage`.
- Replayable daemon events with `Last-Event-ID` continuity.
- Run lifecycle vocabulary for `run_started`, `run_completed`, and `run_error`.

StarClaw intentionally does not enable Kocoro/Shannon Cloud behavior by
default:

- No Shannon Cloud auth.
- No off-machine event transport or telemetry.
- No IM `MESSAGE_LIFECYCLE` protocol unless a future explicit local/client
  boundary task adds it.
