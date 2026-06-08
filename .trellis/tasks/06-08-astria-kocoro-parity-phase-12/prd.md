# Astria Kocoro parity phase 12: event replay and lifecycle resilience

## Goal

Close the next Kocoro parity gap after streaming hardening: local daemon event continuity, replay, lifecycle vocabulary, and Astria client recovery after reconnect or refresh.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase11 closed basic streaming parity for `/message`, provider streaming, OpenAI-compatible streaming, and Web UI delta consumption.
- Phase11 final review identified the remaining Kocoro gap as lifecycle resilience and Desktop-style event continuity, not basic streaming support.
- StarClaw already has a local `EventBus`, `/events` SSE endpoint, event IDs, replay from `last_event_id` and `Last-Event-ID`, and Web UI `EventSource("/events")` for approval events.
- Kocoro's event bus has a larger replay ring and an atomic subscribe-with-replay path for reconnecting SSE clients.
- StarClaw must remain local-first. Phase12 must not add Shannon Cloud auth, off-machine telemetry, or remote Desktop transport.

## Child Plan

1. `eventbus-replay-sse-resilience`: harden `/events` replay and reconnect semantics, including atomic replay+subscribe behavior and Astria client reconnect state.
2. `run-session-lifecycle-events`: normalize local lifecycle events needed for reconnect/recovery, such as run started/completed/error and session/run status transitions.
3. `webui-live-recovery`: recover live run timelines and pending state after `/events` reconnect or page refresh.
4. `event-contract-documentation`: document canonical event names, legacy aliases, replay behavior, and client expectations.

## Requirements

- Preserve existing `/events` clients and current event payload shapes unless a compatibility alias is added.
- Keep event replay local and bounded; do not persist all daemon events to cloud or disk unless a child explicitly scopes local-only persistence.
- Make reconnect behavior deterministic enough to test with `httptest` and static Web UI contract checks.
- Record any intentional divergence from Kocoro in task artifacts.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and testable acceptance criteria.
- [ ] `/events` reconnect behavior is stronger or explicitly documented as intentionally different from Kocoro.
- [ ] Astria UI can surface reconnect/recovered state without dropping critical approval or lifecycle events.
- [ ] No Phase12 work introduces real cloud transport, Shannon Cloud auth, or off-machine telemetry by default.
- [ ] Phase can close only after all children are archived and a final Kocoro gap review is recorded.
