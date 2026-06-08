# Astria Kocoro parity phase 11: streaming hardening

## Goal

Close the remaining streaming reliability and contract gaps between StarClaw/Astria and the local Kocoro baseline, starting from local OpenAI-compatible gateway streaming and then moving toward daemon SSE/provider watchdog parity.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- StarClaw already has provider streaming for CLI/TUI and local daemon runs.
- StarClaw `POST /message` supports SSE when `Accept: text/event-stream` is present.
- StarClaw `POST /v1/chat/completions` supports `stream:true` in code, but `README.md` still describes this gateway as non-streaming and says streaming is unsupported.
- Kocoro has mature streaming contracts around provider `CompleteStream`, daemon SSE delta/usage events, and stream idle watchdog behavior.

## Child Plan

1. `openai-gateway-streaming-parity`: make StarClaw's OpenAI-compatible gateway streaming contract explicit, tested, and documented.
2. `daemon-sse-event-vocabulary-parity`: compare StarClaw `/message` SSE events with Kocoro's delta/usage/tool event vocabulary and align compatible local clients.
3. `provider-stream-watchdog-parity`: add or verify provider stream idle timeout behavior so silent stalled upstream streams fail deterministically.
4. `webui-delta-consumption-parity`: verify the embedded Astria UI consumes streamed deltas instead of only rendering final `done` responses.

## Constraints

- Keep StarClaw local-first. Do not add cloud upload, off-machine telemetry, or Shannon Cloud auth dependencies.
- Keep OpenAI-compatible behavior scoped to local daemon HTTP API.
- Do not silently accept unsupported OpenAI fields that StarClaw cannot implement correctly.
- Preserve normal daemon permission, approval, session, run-store, control, and structured-event paths.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and testable acceptance criteria.
- [ ] Streaming contract drift found during implementation is reflected in docs or follow-up tasks.
- [ ] No phase work introduces real cloud transport or off-machine telemetry by default.
- [ ] Phase can be closed only after all children are archived and a final gap review is recorded.
