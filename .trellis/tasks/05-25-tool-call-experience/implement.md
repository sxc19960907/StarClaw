# Implementation Plan

1. Update TUI message types for full tool result context and streamed completion.
2. Move tool-call/tool-result state mutations into reusable model methods.
3. Make `TUIEventHandler` send Tea messages when a program is available, with direct fallback.
4. Make streamed `sendMessage` return an idle transition message instead of nil.
5. Add focused TUI tests for event update paths.
6. Run focused and full tests.

## Validation Commands

```bash
go test ./internal/tui ./internal/agent ./cmd
go test ./...
```
