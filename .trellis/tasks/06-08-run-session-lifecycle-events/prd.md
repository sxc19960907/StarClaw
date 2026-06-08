# Run session lifecycle events

## Goal

Publish local, replayable run/session lifecycle events through the daemon
`/events` EventBus so Astria clients can recover run state after reconnect or
refresh. This closes the Phase12 gap between StarClaw's persisted run
structured events and Kocoro's richer lifecycle/event-bus recovery model.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- StarClaw already persists redacted structured run events in `RunStore`:
  `run_started`, `run_completed`, `run_error`, `budget_status`,
  `routing_selected`, `fallback_decision`, `control_decision`, and
  `workflow_step`.
- StarClaw's daemon EventBus already supports IDs, bounded history, and
  atomic `SubscribeWithReplay` for `/events` reconnects.
- StarClaw's `/events` bus currently emits tool/runtime/approval events, but
  `RunStore.Start` and `RunStore.Complete` do not publish their lifecycle
  transitions to the EventBus.
- Kocoro exposes a broader daemon lifecycle model, including `run_status`
  bus events and Cloud IM `MESSAGE_LIFECYCLE` transitions. StarClaw remains
  local-first, so this task should not add Shannon Cloud auth, WebSocket cloud
  transport, or IM reaction lifecycle wire protocols.
- Existing StarClaw structured event redaction rules must be preserved:
  prompt/content/tool args/request/response values must not appear in
  observability surfaces.

## Requirements

- Publish canonical EventBus lifecycle events for run start, completion, and
  error using the existing event names `run_started`, `run_completed`, and
  `run_error`.
- Lifecycle EventBus payloads must be JSON, redacted, and recovery-oriented:
  include safe identifiers and summary fields such as `run_id`, `status`,
  `agent`, `channel`, `source`, `session_id`, and safe aggregate
  usage/budget/routing/fallback fields when present.
- The lifecycle bus path must preserve existing persisted `RunRecord.Events`
  and `RunRecord.StructuredEvents` behavior.
- `/events` subscribers must receive lifecycle events live, and reconnecting
  subscribers must receive missed lifecycle events via existing replay.
- Event payloads must not leak prompt text, assistant text, tool args, request
  bodies, response bodies, API keys, bearer tokens, secrets, or passwords.
- Existing clients and event names must remain compatible. Additive behavior
  is allowed; breaking changes are out of scope.

## Out of Scope

- Kocoro/Shannon Cloud IM `MESSAGE_LIFECYCLE` wire protocol.
- Real cloud transport, remote telemetry, or off-machine lifecycle export.
- Full Web UI live-run recovery. This task may leave detailed UI reconciliation
  to the Phase12 `webui-live-recovery` child.
- Persisting EventBus history beyond the current bounded in-memory replay ring.

## Acceptance Criteria

- [ ] `RunStore.Start` causes a live EventBus subscriber to observe a
      `run_started` event with safe run metadata.
- [ ] `RunStore.Complete(..., nil)` causes a live EventBus subscriber to
      observe a `run_completed` event with safe terminal metadata.
- [ ] `RunStore.Complete(..., err)` and `RunAgentResponse.Error` cause
      `run_error` EventBus events with safe error summaries.
- [ ] A `/events` subscriber reconnecting with `last_event_id` receives missed
      lifecycle events without duplicate live delivery.
- [ ] Lifecycle EventBus payload tests prove prompt/content/tool args/request
      and response bodies are not leaked.
- [ ] Existing daemon tests for run records, metrics, tracing, OpenAI API, and
      workflow control continue to pass.
