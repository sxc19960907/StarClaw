# Implementation Plan

1. Add or update agent named config fields for thinking, reasoning effort, and model override.
2. Add a shared command helper that applies advanced agent options to an `AgentLoop`.
3. Use the helper in CLI chat and interactive TUI setup.
4. Avoid duplicate CLI final output when streaming printed deltas.
5. Wire TUI `OnStreamDelta` through the existing `streamingMsg` path.
6. Add tests for option propagation and named agent overrides.
7. Run focused package tests, then `go test ./...`.

## Validation Commands

```bash
go test ./internal/agent ./internal/config ./internal/client ./internal/tui ./cmd
go test ./...
```
