# Results

Date: 2026-05-25

## Completed

- Routed TUI tool-call events through Bubble Tea messages when a program is available.
- Routed TUI tool-result events through Bubble Tea messages when a program is available.
- Preserved direct update fallback for tests and non-program contexts.
- Moved tool-call and tool-result state changes into model methods.
- Tool results now update the visible tool row and populate expandable result history.
- Streamed responses now finish with an idle transition instead of leaving the TUI in receiving state.
- Removed the misleading approval-required state for ordinary tool progress in the current non-interactive TUI flow.
- Added focused TUI tests for tool-call messages, tool-result messages, tracked result history, and streaming completion.

## Verification

- `go test ./internal/tui`
- `go test ./internal/tui ./internal/agent ./cmd`
- `git diff --check`
- `go test ./...`

## Notes

- Real interactive tool approval continuation remains out of scope.
- `.agents/skills/obsidian-cli/` remains intentionally untracked and unrelated.
