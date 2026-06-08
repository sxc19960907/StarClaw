# OpenAI gateway streaming parity

## Goal

Make StarClaw's local `POST /v1/chat/completions` streaming path explicit, documented, and strongly tested so local OpenAI-compatible clients can rely on `stream:true` behavior instead of falling back to blocking responses.

## Confirmed Facts

- StarClaw already has `openAIChatCompletionRequest.Stream` and `handleOpenAIChatCompletionsStream` in `internal/daemon/openai_api.go`.
- StarClaw already emits OpenAI-style `chat.completion.chunk` frames and a final `data: [DONE]` on success.
- Existing tests cover basic streaming, duplicate final text suppression, and error frames.
- `README.md` currently says the OpenAI-compatible gateway is non-streaming and that streaming is unsupported. This contradicts the current code.
- Kocoro does not expose this exact local OpenAI-compatible gateway as its daemon API, but it does have provider-level streaming and SSE-first daemon behavior. The relevant parity target for this child is reliable local streaming UX and clear client contract.

## Requirements

- Update docs to describe `stream:true` support for `POST /v1/chat/completions`.
- Add a documented curl example for OpenAI-compatible streaming.
- Strengthen tests so the streaming contract proves:
  - response status is `200`;
  - `Content-Type` is `text/event-stream`;
  - first chunk contains `delta.role = "assistant"`;
  - content deltas are emitted incrementally;
  - final success chunk has `finish_reason = "stop"`;
  - success streams end with exactly one `data: [DONE]`;
  - error streams emit an error frame and do not emit `[DONE]` or success stop chunks;
  - run store records `Source = "openai-compatible"` and `EnableStreaming = true`.
- Keep unsupported OpenAI fields rejected with clear errors:
  - tool/function calling fields;
  - `response_format`;
  - `metadata`;
  - `n > 1`;
  - unknown fields.
- Preserve local-only boundaries and existing daemon execution path.

## Acceptance Criteria

- [ ] README no longer claims `/v1/chat/completions` is non-streaming-only.
- [ ] `docs/EXAMPLES.md` includes a streaming curl example for `/v1/chat/completions`.
- [ ] `internal/daemon/openai_api_test.go` has regression coverage for exactly-one `[DONE]`, role chunk ordering, incremental content chunks, error stream behavior, and run-store streaming metadata.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] No real network/cloud uploader/auth dependency is introduced.

## Out of Scope

- Full OpenAI tool/function-call streaming.
- Multiple choices (`n > 1`).
- Provider stream idle timeout watchdog. That is a separate Phase11 child.
- Changing `/message` SSE event vocabulary. That is a separate Phase11 child.
