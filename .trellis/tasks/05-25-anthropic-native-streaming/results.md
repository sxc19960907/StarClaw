# Results

Date: 2026-05-25

## Completed

- Added `AnthropicClient.StreamChat` so the default provider implements the agent streaming interface.
- Added shared Anthropic Messages request-body construction for streaming and non-streaming calls.
- Added `ParseAnthropicStream` for Anthropic SSE events.
- Supported streamed text deltas, final text assembly, tool-use input JSON deltas, stop reason, and usage.
- Preserved system prompt, tools, thinking config, reasoning effort, and specific model override in streaming requests.
- Added HTTP and parser regression tests for Anthropic streaming.
- Made stream tool-call finalization robust to non-contiguous content block indexes.

## Verification

- `go test ./internal/client`
- `go test ./internal/client ./internal/agent ./cmd ./internal/tui`
- `git diff --check`
- `go test ./...`

## Notes

- Thinking deltas are parsed as ignored events for now; displaying them is still out of scope.
- `.agents/skills/obsidian-cli/` remains intentionally untracked and unrelated.
