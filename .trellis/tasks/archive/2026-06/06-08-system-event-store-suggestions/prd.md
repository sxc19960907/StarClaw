# System event store and suggestions

## Goal

Add Kocoro-style route-scoped system event and suggestion foundations to StarClaw's daemon so channel/runtime signals can be queued for the correct next turn and UI clients can retrieve/accept session suggestions.

This is Phase 7 child 3 under `Astria Kocoro parity phase 7: channel and cloud delivery parity`.

## Confirmed Facts

- Local Kocoro baseline is `/Users/timmy/PycharmProjects/Kocoro` at commit `74cdb3c`.
- Kocoro has route-scoped next-turn event storage in `internal/daemon/system_event_store.go`.
- Kocoro has suggestion APIs in `internal/daemon/suggestion_handler.go`.
- Kocoro has agent-side system event and suggestion types in:
  - `internal/agent/system_event.go`
  - `internal/agent/suggestion.go`
  - `internal/agent/suggestion_state.go`
- StarClaw already has structured trace/metric events, but those are observability records, not model-facing next-turn system events.
- StarClaw Phase 7 has already added cloudflow dispatch, route index, and connection state foundations.

## Requirements

- Add a route-scoped `SystemEventStore` in the daemon.
- Store only daemon-authored system event metadata/text needed for next-turn injection.
- Bound events per route with deterministic FIFO eviction.
- Collapse consecutive events with the same non-empty context key by replacing the older event.
- Support `Enqueue`, `Drain`, and `Forget` semantics.
- Add agent-facing `SystemEvent` type and minimal formatting/injection helpers.
- Add a suggestion state store keyed by session id.
- Add read and accept endpoints for suggestions on both default and named-agent session routes if the existing router/session model supports both.
- Accepting a suggestion must consume it and return the accepted text.
- Keep system events route-scoped; events for one route must not surface on another.
- Do not use broad observability/metrics stores as the source of model-facing system events.
- Do not enable real external channel transport in this task.

## Acceptance Criteria

- [ ] Unit tests cover enqueue, drain, forget, context-key collapse, nil/empty no-op behavior, and FIFO cap eviction.
- [ ] Tests prove route isolation: draining route A does not return route B events.
- [ ] Suggestion state tests cover set/get/clear/accept behavior.
- [ ] Handler tests cover no-suggestion 404, get suggestion, accept suggestion, and consumed-after-accept behavior.
- [ ] Existing daemon tests still pass.
- [ ] `go test ./internal/agent ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Out of Scope

- Real Slack, Feishu/Lark, Telegram, LINE, or webhook transport.
- WebSocket cloud-controller lifecycle.
- Delivery failure injection; that is the next Phase 7 child.
- UI styling or Astria workbench polish.
- OpenAI-compatible streaming/tool-call delta work.

## Evidence

- `.trellis/research/kocoro-local-comparison-phase7-plan.md`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/system_event_store.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/suggestion_handler.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/system_event.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/suggestion.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/agent/suggestion_state.go`
