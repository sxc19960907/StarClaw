# Improve Tool Call Experience

## Goal

Make tool call progress in the TUI reliable and easier to read during streaming and tool-use turns.

## Confirmed Facts

- TUI already has `toolCallMsg`, `toolResultMsg`, `streamingMsg`, compact tool result formatting, and expandable result formatting.
- `TUIEventHandler.OnStreamDelta` already sends `streamingMsg` through Bubble Tea's program message queue.
- `TUIEventHandler.OnToolCall` and `OnToolResult` currently mutate the model directly, bypassing the Bubble Tea message queue.
- Direct mutation from agent callbacks can fail to repaint promptly and risks inconsistent state.
- `toolResultMsg` currently updates only `pendingTool` and does not record expandable result history.
- A streaming response that returns nil from `sendMessage` can leave the TUI in `StateStreaming`.

## Requirements

- TUI tool-call start events must be delivered through the Bubble Tea message queue when a program is available.
- TUI tool-result events must be delivered through the Bubble Tea message queue when a program is available.
- The existing direct update behavior must remain usable in tests or contexts without a Bubble Tea program.
- Tool results must update the visible tool-call row and populate expandable result history.
- After a streamed agent response completes, the TUI must return to idle without appending duplicate final text.
- Existing CLI tool display must remain unchanged.

## Acceptance Criteria

- Unit tests cover tool call event handling via model update.
- Unit tests cover tool result update and expandable result history.
- Unit tests cover streaming completion transitioning back to idle.
- `go test ./internal/tui ./internal/agent ./cmd` passes.
- `go test ./...` passes or unrelated failures are documented.

## Out Of Scope

- Redesigning the whole TUI layout.
- Implementing real interactive tool approval continuation.
- Changing CLI output formatting.
