# Results

Date: 2026-05-25

## Completed

- Applied `agent.thinking`, `agent.thinking_mode`, `agent.thinking_budget`, `agent.reasoning_effort`, and `agent.model` to normal CLI and TUI agent loops.
- Enabled streaming at loop construction for normal runs; unsupported clients continue through non-streaming `Chat`.
- Prevented duplicate CLI final text when streaming already printed deltas.
- Wired TUI streaming deltas into the existing `streamingMsg` update path.
- Added named agent config override support for thinking, reasoning effort, thinking budget, and specific model.
- Added regression tests for `AgentLoop` chat option propagation and named agent advanced overrides.

## Verification

- `go test ./internal/agent ./internal/config ./internal/client ./internal/tui ./cmd`
- `go test -count=1 ./internal/client`
- `go test -count=1 ./cmd`
- `go test ./...`
- `git diff --check`

## Notes

- Anthropic native SSE streaming remains out of scope for this task.
- `.agents/skills/obsidian-cli/` remains intentionally untracked and unrelated to this change.
