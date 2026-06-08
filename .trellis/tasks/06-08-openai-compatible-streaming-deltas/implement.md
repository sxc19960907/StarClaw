# Implementation Plan

## Steps

1. Update OpenAI validation so `stream=true` is allowed while tool/function fields remain rejected.
2. Add OpenAI streaming chunk types/helpers.
3. Add streaming event handler and `handleOpenAIChatCompletionsStream`.
4. Route `req.Stream` to the streaming path.
5. Add tests for streaming success response, chunks, `[DONE]`, and run recording.
6. Update validation test expected message for stream.
7. Run:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-openai-compatible-streaming-deltas`
8. Commit and archive the child task.

## Review Gates

- Do not add tool-call streaming in this slice.
- Do not alter non-streaming JSON shape.
- Do not leak prompt text into metrics/traces.
- Ensure final `[DONE]` is sent on success.

## Completion Notes

TBD.
