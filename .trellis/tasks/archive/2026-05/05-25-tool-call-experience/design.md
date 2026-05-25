# Design

## Event Flow

Agent callbacks can occur while `loop.Run` is executing inside a Tea command. The safe UI path is:

`AgentLoop callback -> TUIEventHandler -> tea.Program.Send -> Model.Update`

When no program is attached, the handler should still update the model directly to preserve simple tests and non-program contexts.

## Messages

Extend `toolResultMsg` to include tool name, args, content, error flag, and elapsed duration. This gives `Model.Update` enough context to:

- update the visible `pendingTool`
- append a `toolResultEntry`
- reset the tool expansion level
- return the state to thinking so the UI no longer appears blocked on approval after a tool finishes

Add an `agentDoneMsg` for streamed runs. `sendMessage` returns this when streaming already emitted text; `Model.Update` sets state back to idle without appending duplicate content.

## Compatibility

CLI output stays unchanged. The existing `formatCompactToolResult` and `formatExpandedToolResult` helpers remain the rendering source.
