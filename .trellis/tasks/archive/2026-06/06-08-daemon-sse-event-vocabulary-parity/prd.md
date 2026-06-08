# Daemon SSE event vocabulary parity

## Goal

Align StarClaw's `POST /message` SSE event vocabulary with the local Kocoro baseline so local clients can consume live daemon runs using Kocoro-style `delta`, `usage`, `tool`, `assistant_text`, and `session_started` events while existing StarClaw clients remain compatible.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- StarClaw `POST /message` supports SSE when `Accept: text/event-stream` is present.
- StarClaw currently emits:
  - `tool_call` with running tool metadata;
  - `tool_result` with completed/error tool metadata;
  - `text` for both stream deltas and final fallback text;
  - `preamble` for preamble narration;
  - `done` and `error`.
- StarClaw `OnUsage` currently does not emit a separate SSE event; usage is available in the final `done` payload and run store.
- Kocoro emits:
  - `session_started` once the session id is known;
  - unified `tool` events with `status=running|completed`;
  - `delta` events for streamed text;
  - `usage` events for live token/cost meter updates;
  - `assistant_text` for preamble/mid-turn narration;
  - `done` and `error`.
- StarClaw's current `agent.EventHandler` does not carry Kocoro's `tool_use_id`, elapsed duration, or cost fields, so exact payload parity is out of scope for this child.

## Requirements

- Preserve existing StarClaw SSE event names so current clients do not break.
- Add Kocoro-compatible event aliases on the per-request `/message` SSE path:
  - `tool` for running tool calls;
  - `tool` for completed/error tool results;
  - `delta` for `OnStreamDelta`;
  - `assistant_text` for `OnPreamble`;
  - `usage` for `OnUsage`;
  - `session_started` when the session id is known.
- Keep `text` delta events for existing StarClaw clients, but avoid duplicate final text after streamed deltas.
- Usage events must include `input_tokens`, `output_tokens`, and `total_tokens` from `client.Usage`.
- Tool aliases must include enough data to render a Kocoro-like progress row:
  - `tool`;
  - `status`;
  - redacted/truncated `args` for running tools;
  - `is_error`, `preview`, and `error_category` for results when available.
- Do not introduce real cloud transport, auth, or off-machine telemetry.

## Acceptance Criteria

- [ ] `sseEventHandler.SetSessionID` emits `session_started` without breaking empty-session behavior.
- [ ] `OnStreamDelta` emits both legacy `text` and Kocoro-compatible `delta`.
- [ ] `OnPreamble` emits both legacy `preamble` and Kocoro-compatible `assistant_text`.
- [ ] `OnUsage` emits a `usage` event with `input_tokens`, `output_tokens`, and `total_tokens`.
- [ ] `OnToolCall` emits legacy `tool_call` and Kocoro-compatible `tool status=running`.
- [ ] `OnToolResult` emits legacy `tool_result` and Kocoro-compatible `tool status=completed|error`.
- [ ] Handler tests cover all new event aliases and legacy compatibility.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.

## Out of Scope

- Changing the global `/events` ring.
- Adding `tool_use_id`, elapsed time, cost, or model to StarClaw's base `agent.EventHandler`.
- Changing OpenAI-compatible `/v1/chat/completions` streaming.
- Provider stream idle watchdog behavior.
