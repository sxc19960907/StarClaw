# Terminal workspace tool depth design

## Reference

Kocoro exposes a macOS-only `ghostty` tool implemented through:

- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty_darwin.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty_stub.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/ghostty_test.go`

StarClaw will implement the same local desktop workflow capability under the StarClaw-owned tool name `terminal_workspace`.

## Tool contract

Tool name: `terminal_workspace`

Actions:

- `status`: report platform, Ghostty availability, minimum version, and fallback guidance.
- `list_tabs`: list tabs tracked during this process.
- `new_tab`: open a visible Ghostty tab, optionally run a command, and track the title.
- `new_split`: open a visible Ghostty split, defaulting to `right`, optionally run a command, and track the title.
- `send_input`: send text to a tracked tab title.

Arguments:

- `action` is required.
- `description` is required by schema guidance for model-facing calls.
- `command`, `title`, `direction`, `target`, and `text` are action-specific optional fields.

Safety:

- `RequiresApproval()` returns true because the tool opens terminals and sends input.
- `IsReadOnlyCall()` returns true only for `status` and `list_tabs` with valid JSON.
- The tool never shells out directly for user commands; commands are sent to a visible terminal after approval.
- No remote terminal, SSH, cloud shell, sync, credential, or daemon transport is added.

## Implementation boundary

Common file:

- `internal/tools/terminal_workspace.go`
- Owns schema, argument validation, tab registry, title resolution, result formatting, and tool interface behavior.
- Calls package-level Ghostty helper functions behind a small runtime boundary.

Platform files:

- `internal/tools/terminal_workspace_darwin.go`
  - Checks Ghostty installation through Spotlight metadata and `defaults read`.
  - Executes Ghostty AppleScript through `osascript`.
  - Opens new tabs/splits and sends input using the same visible terminal model as Kocoro.

- `internal/tools/terminal_workspace_stub.go`
  - Builds on non-darwin.
  - Reports unsupported availability and returns explicit errors for mutating operations.

Registration:

- `internal/tools/register.go` registers `NewTerminalWorkspaceTool()`.

Tests:

- `internal/tools/terminal_workspace_test.go` uses injected package-level hooks to avoid opening real Ghostty.
- `internal/tools/terminal_workspace_version_test.go` covers version comparison.

## Compatibility

This is additive. Existing `bash`, `applescript`, `process`, and browser tools keep their contracts. The fallback text intentionally points model callers to those tools when a visible Ghostty workspace is unavailable.

## Rollback

Rollback is deleting the new terminal workspace files and removing the single registry line. No persisted state or external data migration is introduced.
