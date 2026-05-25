# Add Anthropic Native Streaming

## Goal

Make the default Anthropic provider implement the existing agent streaming interface so CLI/TUI streaming works on StarClaw's main provider path.

## Confirmed Facts

- `AgentLoop` already detects `StreamingLLMClient` and calls `StreamChat` when streaming is enabled.
- OpenAI-compatible and Ollama clients already implement `StreamChat`.
- `AnthropicClient` currently implements only non-streaming `Chat`.
- Anthropic Messages streaming uses SSE events including `message_start`, `content_block_start`, `content_block_delta`, `message_delta`, and `message_stop`.
- Text deltas arrive as `content_block_delta` with `delta.type == "text_delta"`.
- Tool use starts as a `content_block_start` block with type `tool_use`, then input arrives as `input_json_delta` chunks.

## Requirements

- `AnthropicClient` must implement `StreamChat` with the same signature as other streaming clients.
- Streaming requests must use the existing Anthropic Messages endpoint with `"stream": true`.
- Request construction must preserve model override, tools, system prompt, thinking config, and reasoning effort behavior from non-streaming `Chat`.
- Text deltas must call `onDelta` immediately and assemble into the final `Response.Content`.
- Tool use streams must assemble final `Response.ToolUses` with ID, name, and JSON input.
- Usage and stop reason from Anthropic stream events must be reflected in the final response when present.
- Non-OK stream responses must return useful API errors.

## Acceptance Criteria

- Unit tests cover Anthropic text streaming with delta callbacks.
- Unit tests cover Anthropic tool-use streaming with input JSON delta accumulation.
- Unit tests cover Anthropic usage and stop reason parsing.
- Unit tests cover `AnthropicClient.StreamChat` request body and headers.
- `go test ./internal/client ./internal/agent ./cmd ./internal/tui` passes.
- `go test ./...` passes or unrelated failures are documented.

## Out Of Scope

- Displaying thinking deltas in CLI/TUI.
- Changing OpenAI/Ollama streaming behavior.
- Retrying partially interrupted streams beyond existing agent retry handling.
