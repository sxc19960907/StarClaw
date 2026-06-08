# Browser handoff lease depth implementation plan

## Checklist

- [x] Read Kocoro browser lease/handoff behavior and existing StarClaw browser tests.
- [x] Add browser lease tracker and context helpers.
- [x] Add BrowserTool owner/deprecated/cleanup test hooks.
- [x] Mark browser use from `BrowserTool.Run`.
- [x] Install/release browser leases in daemon `RunAgentWithApproval`.
- [x] Add focused lease/handoff/browser marking tests.
- [x] Run focused and full tests.
- [x] Validate Trellis artifacts.
- [x] Archive and commit the child task.

## Validation Commands

```bash
go test ./internal/tools ./internal/daemon
go test ./...
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-browser-handoff-lease-depth
git diff --check
```

## Risk Points

- Do not alter browser action semantics.
- Avoid data races in owner counters and deprecated state.
- Release must be idempotent and must not decrement counters below zero.
