# Terminal workspace tool depth implementation plan

## Steps

1. Read task artifacts and applicable Trellis backend specs.
2. Add `terminal_workspace` common tool implementation.
3. Add darwin and non-darwin Ghostty helper implementations.
4. Register the tool in `RegisterLocalTools`.
5. Add focused unit tests for:
   - tool info and registration
   - approval/read-only behavior
   - invalid JSON and unknown action
   - status and list fallback without Ghostty
   - missing fields for `send_input`
   - invalid split direction
   - successful new tab/split tracking via injected fake functions
   - version comparison
6. Run:
   - `gofmt` on touched Go files
   - `go test ./internal/tools`
   - broader `go test ./internal/daemon ./internal/tools` if registry wiring affects daemon usage
7. Run Trellis validation and archive the child task.
8. Commit the completed child task independently.

## Review gates

- Do not call real Ghostty or `osascript` from tests.
- Do not add real remote terminal or cloud behavior.
- Keep all user-facing runtime text StarClaw-branded.
