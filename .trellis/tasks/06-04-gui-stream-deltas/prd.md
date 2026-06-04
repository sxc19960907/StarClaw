# Stream Model Deltas to GUI

## Problem

The daemon can call streaming LLM clients and records `stream_delta` events, but the HTTP SSE handler does not forward `OnStreamDelta` to the browser. The GUI only shows assistant text after the full model response completes.

## Scope

- Forward model stream deltas from daemon SSE `/message` responses to the browser.
- Keep final `done` response and run history behavior intact.
- Extend the existing fake OpenAI-compatible streaming smoke to assert partial text appears before the final response completes.

## Acceptance Criteria

- [x] `sseEventHandler.OnStreamDelta` emits browser-consumable text events.
- [x] GUI streaming smoke verifies partial output is visible while the provider stream is still open.
- [x] Final chat output does not duplicate streamed text.
- [x] Run history/detail still includes final response and usage.
- [x] `scripts/smoke_webui_streaming.sh` passes locally.
- [x] `scripts/smoke_webui_core.sh` passes locally.
- [x] `go test ./...` and `go vet ./...` pass locally.

## Notes

- `OnStreamDelta` now emits the same browser `text` event shape as completed text chunks.
- The SSE handler suppresses the final `OnText` event after streaming deltas have already been emitted to avoid duplicate GUI output.
- Validation completed with `go test ./internal/daemon`, `scripts/smoke_webui_streaming.sh`, `scripts/smoke_webui_core.sh`, `go test ./...`, `go vet ./...`, and `git diff --check`.
