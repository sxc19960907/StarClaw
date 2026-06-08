# Design

## Handler Flow

`handleOpenAIChatCompletions` keeps the existing validation and run request construction.

If `req.Stream` is false:

- keep current synchronous behavior.

If `req.Stream` is true:

- require `http.Flusher`.
- set SSE headers.
- start the run record.
- create runtime cancel/pause handle consistently with the normal daemon run path.
- run the agent with a custom OpenAI streaming handler.
- complete run store.
- emit terminal stop chunk and `[DONE]`.

## OpenAI Streaming Handler

Add an internal handler implementing `agent.EventHandler`:

- `OnStreamDelta(delta)` writes a content chunk.
- `OnText(text)` writes content only if no stream deltas were already emitted, matching existing `sseEventHandler` behavior.
- `OnToolCall` / `OnToolResult` can be no-op because OpenAI tool streaming is out of scope and tool request fields remain rejected.
- `OnUsage` stores usage for the final run result through normal `RunAgentResponse`, not per-chunk.

The chunk shape:

```json
{
  "id": "chatcmpl-<request_id>",
  "object": "chat.completion.chunk",
  "created": 123,
  "model": "request-model",
  "choices": [
    {"index": 0, "delta": {"content": "hi"}, "finish_reason": null}
  ],
  "starclaw_run_id": "<request_id>"
}
```

Initial role chunk:

```json
{"choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}
```

Final chunk:

```json
{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}
```

Then:

`data: [DONE]`

## Error Behavior

If validation fails before headers, return OpenAI JSON error as today.

If the run fails after SSE headers are sent, emit an SSE chunk with `error` would diverge from OpenAI. For this first slice, emit a terminal error-shaped SSE data object only if needed and do not send `[DONE]`. Tests focus on success path.

## Compatibility

Non-streaming response structures and tests should remain unchanged.
