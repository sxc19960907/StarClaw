# Add Agent Thinking And Streaming Support

## Goal

Make StarClaw's existing thinking/reasoning/streaming configuration actually affect normal agent runs so CLI and TUI users get provider request options and incremental output without duplicate final text.

## Confirmed Facts

- `internal/config.AgentConfig` already has `thinking`, `thinking_mode`, `thinking_budget`, `reasoning_effort`, and `model`.
- `internal/client.ChatOptions` already carries `Thinking`, `ReasoningEffort`, and `SpecificModel`.
- Anthropic, OpenAI-compatible, and Ollama clients already accept these options in request construction, with streaming implemented for OpenAI-compatible clients.
- `internal/agent.AgentLoop` already has setters for thinking, reasoning effort, specific model, and streaming.
- CLI and TUI construction paths configure max iterations, max tokens, context window, permissions, and hooks, but do not currently apply thinking/reasoning/model/streaming options.
- CLI `OnStreamDelta` prints deltas inline, while `runChat` also prints the final response, so enabling streaming directly would duplicate text.
- TUI has a `streamingMsg` update path but `TUIEventHandler.OnStreamDelta` is currently a no-op.

## Requirements

- Normal CLI chat and interactive TUI runs must apply agent thinking config from `config.Config`.
- Normal CLI chat and interactive TUI runs must apply `agent.reasoning_effort` and `agent.model`.
- Streaming should be enabled for normal runs when the selected client supports it.
- CLI streaming output must not be printed twice.
- TUI streaming deltas must reach the UI using the existing `streamingMsg` path.
- Named agent config overrides should support thinking, reasoning effort, and model overrides.
- Existing non-streaming behavior must remain compatible for clients that do not implement streaming.

## Acceptance Criteria

- Unit tests prove loop options are passed to the LLM request.
- Unit tests prove named agent overrides include thinking/reasoning/model settings.
- Unit tests or package tests prove CLI/TUI construction compiles with streaming enabled.
- `go test ./internal/agent ./internal/config ./internal/client ./internal/tui ./cmd` passes.
- `go test ./...` passes or any unrelated failure is documented.

## Out Of Scope

- Implementing Anthropic native SSE streaming.
- Replacing OpenAI Chat Completions with Responses API.
- GUI/browser automation.
- MCP server expansion.
