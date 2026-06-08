# Phase 11 final gap review

## Baseline

- Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.
- StarClaw comparison state: after `c09daa0 feat: consume Kocoro SSE aliases in Web UI`.

## Completed streaming parity scope

- `openai-gateway-streaming-parity`: StarClaw's local OpenAI-compatible `/v1/chat/completions` endpoint supports `stream:true`, emits OpenAI-style chunks, suppresses duplicate final text after deltas, stores runs, and rejects unsupported OpenAI fields explicitly.
- `daemon-sse-event-vocabulary-parity`: `POST /message` emits Kocoro-compatible `delta`, `assistant_text`, `usage`, `tool`, and `session_started` aliases while preserving legacy StarClaw `text`, `preamble`, `tool_call`, and `tool_result` events.
- `provider-stream-watchdog-parity`: provider streaming paths expose a configurable stream idle timeout, return `ErrStreamIdleTimeout`, avoid retry/fallback rehangs, and preserve partial streamed text.
- `webui-delta-consumption-parity`: Astria Web UI consumes Kocoro-compatible stream aliases and suppresses duplicate rendering when legacy and alias events are dual-emitted.

## Evidence

- Kocoro daemon SSE handlers emit `session_started`, `assistant_text`, `delta`, `usage`, and `tool` events from `internal/daemon/server.go`.
- Kocoro provider stream watchdog behavior is implemented around `GatewayClient.CompleteStream`, `ErrStreamIdleTimeout`, and agent-loop no-retry handling.
- StarClaw now has equivalent local streaming contracts in:
  - `internal/daemon/openai_api.go`
  - `internal/daemon/server.go`
  - `internal/client/stream.go`
  - `internal/agent/loop.go`
  - `internal/daemon/webui/assets/app.js`

## Remaining differences

- StarClaw intentionally dual-emits legacy and Kocoro-compatible SSE names for local client compatibility. Kocoro's per-request SSE path is closer to a canonical Kocoro vocabulary.
- Kocoro's broader daemon event surface includes richer Desktop/cloud lifecycle concerns such as replayable event bus IDs, route/session lifecycle events, channel queue events, and Desktop-focused reconnect behavior. StarClaw has local-first structured run events and SSE support, but the client-resilience story is not yet as mature.
- Kocoro has a deeper Desktop-oriented transport/lifecycle stack. StarClaw should not add Shannon Cloud auth or off-machine telemetry by default, but can continue aligning local event replay, lifecycle, and recovery contracts.

## Next recommended phase

Phase 12 should focus on daemon lifecycle and client-resilience parity:

1. EventBus replay contract: align `/events` replay, `Last-Event-ID`, keepalive, ring retention, and tests against Kocoro behavior.
2. Run/session lifecycle vocabulary: add or normalize local lifecycle events needed by Astria for reconnect and recovery without adding cloud transport.
3. UI reconnect recovery: make Astria recover live run timelines and in-flight status after SSE reconnect or page refresh.
4. Event contract documentation: document canonical vs legacy event names and expected client behavior.

## Closeout

Phase 11 closes the explicit streaming hardening scope. The remaining Kocoro gap is no longer basic streaming support; it is lifecycle resilience and Desktop-style event continuity.
