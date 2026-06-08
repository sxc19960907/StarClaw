# Browser handoff lease depth

## Goal

Add Kocoro-style browser use lease and handoff foundations to StarClaw so browser tool usage can be tracked per run and future persistent browser cleanup/reload behavior has a safe ownership boundary.

## Confirmed Facts

- Kocoro implements `internal/tools/browser_lease.go` and `browser_handoff.go` for per-run browser ownership and reload cleanup.
- StarClaw's current `BrowserTool` does not own a persistent Chrome process or a cleanup method; it opens URLs and inspects browser state through OS commands/AppleScript.
- StarClaw's agent loop executes tools with a shared run context, and daemon `RunAgentWithApproval` owns the per-run setup/teardown boundary.

## Requirements

- Add a browser lease tracker with idempotent acquire/release behavior and per-owner counts.
- Install a fresh browser lease in daemon agent runs.
- Mark the lease whenever `BrowserTool.Run` is invoked.
- Add a handoff helper that can mark a browser owner deprecated and run cleanup only when no leases reference it.
- Keep cleanup callback-based because StarClaw currently has no persistent browser process to destroy.
- Add tests for idempotency, race-safe release, per-owner cleanup, context installation, and `BrowserTool` usage marking.
- Do not change user-visible browser behavior or introduce a new browser runtime.

## Acceptance Criteria

- [x] Browser lease tracker and context helpers are implemented with focused unit tests.
- [x] Daemon runs install and release browser leases.
- [x] `BrowserTool.Run` marks the current run lease when invoked.
- [x] Handoff helper defers cleanup while an owner has active leases.
- [x] Existing browser tool behavior remains unchanged.
- [x] `go test ./internal/tools ./internal/daemon` and `go test ./...` pass.

## Out of Scope

- Adding chromedp or Playwright runtime ownership.
- Starting or killing browser processes.
- Daemon hot-reload implementation beyond the handoff helper.
