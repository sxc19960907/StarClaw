# Calendar native tool boundary implementation plan

## Checklist

- [x] Read backend specs and Kocoro calendar tool behavior.
- [x] Inspect StarClaw tool registration and daemon local tool wiring.
- [x] Add calendar Desktop RPC helper and tool files.
- [x] Add optional Desktop RPC broker registration wiring.
- [x] Add focused calendar tool tests.
- [x] Run `go test ./internal/tools ./internal/daemon`.
- [x] Run `go test ./...`.
- [x] Validate Trellis artifacts.
- [x] Archive and commit this child task.

## Validation Commands

```bash
go test ./internal/tools ./internal/daemon
go test ./...
python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-calendar-native-tool-boundary
git diff --check
```

## Risk Points

- Do not expose calendar tools from no-broker registries.
- Do not leak approval `description` into Desktop RPC params.
- Preserve existing `RegisterLocalTools` API compatibility.
- Do not directly access EventKit.
