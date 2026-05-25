# Design

## Parser

Add an Anthropic-specific stream parser in `internal/client/stream.go` next to the OpenAI-compatible parser.

State tracked while parsing:

- text content builder
- active content blocks by Anthropic block index
- tool-use blocks keyed by index, preserving final order
- usage
- stop reason

Handled events:

- `message_start`: read initial usage if present.
- `content_block_start`: if the block type is `text`, optionally capture initial text; if `tool_use`, capture id/name/input.
- `content_block_delta`: for `text_delta`, append and call `onDelta`; for `input_json_delta`, append to the indexed tool call arguments.
- `message_delta`: read `delta.stop_reason` and incremental/final usage.
- `message_stop`: finish normally.

Unknown event types are ignored for forward compatibility.

## Client

Add `StreamChat` to `AnthropicClient`:

- Build the same request shape as `Chat`, with `"stream": true`.
- Reuse a helper for request body construction to avoid drift between streaming and non-streaming calls.
- Set `Accept: text/event-stream`.
- On non-200, read the body and return the same style of API error as `Chat`.
- On success, pass the response body to `ParseAnthropicStream`.

## Compatibility

The agent loop already falls back to `Chat` if a client does not implement `StreamingLLMClient`. After this change, Anthropic will take the streaming path when enabled.
