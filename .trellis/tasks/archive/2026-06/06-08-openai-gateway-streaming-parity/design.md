# OpenAI gateway streaming parity design

## Boundary

This task is limited to `POST /v1/chat/completions` in the local daemon. It does not change provider clients, `/message` SSE, Web UI rendering, or cloud integrations.

## Current Flow

1. `handleOpenAIChatCompletions` decodes and validates the OpenAI-compatible request.
2. Non-streaming requests run `RunAgentRequest` through `s.runAgent` and return one completion envelope.
3. Streaming requests call `handleOpenAIChatCompletionsStream`.
4. The streaming handler sets SSE headers, sets `RunAgentRequest.EnableStreaming = true`, starts the run store record, and uses `openAIStreamingHandler`.
5. `openAIStreamingHandler` emits:
   - initial role chunk;
   - content chunks from `OnStreamDelta`;
   - fallback final text from `OnText` only when no deltas were seen;
   - stop chunk and `[DONE]` on success;
   - error frame only on failure.

## Contract

Success stream:

```text
Content-Type: text/event-stream

data: {"object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant"},"finish_reason":null}],...}

data: {"object":"chat.completion.chunk","choices":[{"delta":{"content":"..."}}],...}

data: {"object":"chat.completion.chunk","choices":[{"delta":{},"finish_reason":"stop"}],...}

data: [DONE]
```

Error stream:

```text
Content-Type: text/event-stream

data: {"error":{"message":"...","type":"server_error"}}
```

Error streams must not emit `[DONE]` or a success stop chunk.

## Compatibility

- Keep the existing local extension fields: `request_id`, `session_id`, `agent`, and `user`.
- Continue rejecting unsupported fields rather than silently degrading.
- Preserve `starclaw_run_id` on chunks for local clients that need run-store lookup.

## Privacy And Locality

The task updates only local HTTP docs/tests and the local daemon handler if needed. It must not add cloud uploader/auth paths or off-machine telemetry.
