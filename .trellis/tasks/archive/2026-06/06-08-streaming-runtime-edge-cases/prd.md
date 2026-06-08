# Streaming runtime edge cases

## Goal

Harden StarClaw's local OpenAI-compatible streaming runtime so the happy-path `stream=true` implementation behaves predictably under fallback, duplicate-final-text, and run-error scenarios.

## Requirements

- Keep the endpoint local-first: no real cloud credentials, external channel transport, or off-machine telemetry.
- Preserve the existing non-streaming `POST /v1/chat/completions` JSON response contract.
- Preserve successful streaming output shape:
  - `Content-Type: text/event-stream`
  - first assistant role chunk
  - zero or more content chunks
  - one terminal `finish_reason:"stop"` chunk
  - final `data: [DONE]`
- Define and test stream error behavior after SSE headers have been written.
- Ensure fallback final text is streamed once when the agent loop produces `OnText` but no streaming deltas.
- Ensure final `OnText` does not duplicate content when streaming deltas already emitted text.
- Ensure run records still complete with the correct error/success state and source metadata.

## Non-Goals

- Do not add OpenAI tool/function-call streaming deltas in this slice.
- Do not change `/message` SSE event names or the broader `/events` bus.
- Do not add Kocoro-style SSE reconnect/idle watchdog behavior here; that is the next Phase 8 child.

## Acceptance Criteria

- [x] Tests cover successful OpenAI streaming terminal behavior.
- [x] Tests cover fallback `OnText` content when no deltas are emitted.
- [x] Tests cover duplicate suppression when both deltas and final text are observed.
- [x] Tests cover run error/result error after streaming starts, including the emitted SSE error frame and run-store completion.
- [x] `go test ./internal/daemon` passes.
- [x] `go test ./...` passes or any unrelated failure is documented.
