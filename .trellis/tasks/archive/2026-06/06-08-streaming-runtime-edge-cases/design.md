# Streaming runtime edge cases design

## Current Shape

`internal/daemon/openai_api.go` builds a `RunAgentRequest` from OpenAI-compatible chat messages. For `stream=true`, it:

1. Sets SSE headers.
2. Starts a run record.
3. Enables runtime streaming.
4. Writes an assistant role chunk immediately.
5. Runs the agent with `openAIStreamingHandler`.
6. Writes a stop chunk and `[DONE]` on success.
7. Writes an error payload on `err` or `result.Error`.

`openAIStreamingHandler` already suppresses final `OnText` when `OnStreamDelta` emitted text through the `streamedText` flag.

## Contract

### Success

The stream remains OpenAI-compatible for simple text:

- Role chunk is emitted before content.
- Content chunks are emitted as deltas arrive.
- If no deltas arrive but final text arrives, one content chunk is emitted from `OnText`.
- Stop chunk and `[DONE]` are emitted exactly once on success.

### Failure After Stream Start

Once SSE headers and any chunks have been sent, the HTTP status cannot be changed. Errors should therefore be represented as SSE data frames:

```json
{"error":{"message":"...","type":"server_error"}}
```

The stream should then end without a success stop chunk. `[DONE]` should not be emitted for failed streams in this slice, because clients commonly interpret `[DONE]` as a successful terminal marker. This preserves the existing implementation's broad shape while pinning it in tests.

### Duplicate Suppression

The handler should keep the current `streamedText` behavior:

- `OnStreamDelta("a")` then `OnText("ab")` emits only the delta chunk.
- `OnText("ab")` with no prior delta emits one content chunk.

This avoids duplicating the final full answer in clients that concatenate chunks.

## Test Strategy

Use focused daemon tests with custom runner behavior rather than real LLM calls:

- Streaming success shape remains covered by the existing smoke test.
- Handler-level tests cover fallback and duplicate suppression deterministically.
- Server-level tests cover run error/result error after stream start and inspect both SSE frames and run-store status.

## Risk

The main compatibility tradeoff is whether failed streams should emit `[DONE]`. This task explicitly chooses no `[DONE]` on error. The next SSE reconnect/idle-watchdog task can revisit terminal markers if client compatibility evidence requires it.
