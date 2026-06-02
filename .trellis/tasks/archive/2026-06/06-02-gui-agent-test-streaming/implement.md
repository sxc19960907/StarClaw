# Implementation Plan

## Checklist

- [x] Read relevant Trellis specs before code edits.
- [x] Add Agent Test Runner streaming state and Stop control.
- [x] Reuse `/message` SSE flow for agent tests.
- [x] Render streaming progress and final result in the test output.
- [x] Refresh and auto-open matching Run detail after completion.
- [x] Extend Web UI smoke coverage for streaming and cancellation UX.
- [x] Run targeted Go tests and Web UI smoke.

## Validation Commands

```bash
go test ./internal/daemon ./cmd
scripts/smoke_webui_agents.sh
scripts/smoke_webui_runs.sh
git diff --check -- . ':(exclude).agents/skills/obsidian-cli' ':(exclude)output/playwright/*.png'
```
