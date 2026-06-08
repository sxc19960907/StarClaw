# OpenAI compatible streaming deltas

## Goal

Support `stream=true` on StarClaw's OpenAI-compatible `POST /v1/chat/completions` endpoint by returning OpenAI-style Server-Sent Events for assistant text deltas and a final `[DONE]` marker.

## Confirmed Facts

- `internal/daemon/openai_api.go` accepts a `stream` field but currently rejects `stream=true`.
- StarClaw's agent loop already emits `OnStreamDelta` events when the configured LLM client supports streaming.
- Existing daemon `POST /message` SSE handler streams internal `text` events.
- This task is local API compatibility; OpenAI tool/function streaming deltas remain out of scope.

## Requirements

- Stop rejecting `stream=true`.
- When `stream=false` or omitted, preserve the existing JSON response behavior.
- When `stream=true`, respond with `Content-Type: text/event-stream`.
- Emit OpenAI-compatible `chat.completion.chunk` SSE frames.
- Emit an initial role delta with `delta.role="assistant"`.
- Emit text deltas as `choices[0].delta.content`.
- Emit a terminal chunk with `finish_reason="stop"`.
- Emit `data: [DONE]`.
- Preserve run store lifecycle, run ids, usage recording, routing/fallback metadata, and error handling.
- Keep OpenAI tool/function request fields unsupported in this slice.

## Acceptance Criteria

- [ ] Streaming request returns `text/event-stream`.
- [ ] Streaming response includes initial assistant role chunk.
- [ ] Streaming response includes content delta chunk(s).
- [ ] Streaming response includes terminal stop chunk and `[DONE]`.
- [ ] Streaming run is recorded and can be fetched via `/runs/{id}`.
- [ ] Non-streaming tests remain unchanged.
- [ ] Validation still rejects tool/function request fields.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Out of Scope

- OpenAI tool-call delta streaming.
- Multiple choices (`n > 1`).
- Real external channel streaming.
- UI work.
