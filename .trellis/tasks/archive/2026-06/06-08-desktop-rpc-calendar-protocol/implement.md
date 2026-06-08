# Desktop RPC calendar protocol implementation plan

## Checklist

- [x] Read backend specs and existing Desktop RPC tests.
- [x] Compare Kocoro Desktop RPC type constants for the calendar v1 protocol.
- [x] Update `internal/daemon/desktop_rpc/types.go` with calendar methods, error/enum constants, and typed payload structs.
- [x] Add tests for method order, constants, JSON tags, and representative payload round trips.
- [x] Run focused Desktop RPC tests.
- [x] Run broader daemon tests if touched behavior warrants it.
- [x] Validate Trellis task artifacts.
- [x] Archive and commit this child task.

## Validation Commands

```bash
go test ./internal/daemon/desktop_rpc
go test ./internal/daemon
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-desktop-rpc-calendar-protocol
git diff --check
```

## Risk Points

- Do not rename existing structs or JSON fields consumed by current Desktop RPC tests.
- Do not add calendar tools in this task.
- Do not add any system calendar or cloud behavior.
