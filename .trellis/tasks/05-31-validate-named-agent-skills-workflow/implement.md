# Implementation Plan

## Checklist

- [x] Start Trellis task after planning artifacts are complete.
- [x] Add regression coverage for agent tool allow/deny merge behavior.
- [x] Fix `config.MergeAgentConfig` if tool filters are not propagated.
- [x] Add an integration test for named agent load, memory injection, agent-scoped session persistence, registry filtering, and skill activation.
- [x] Replace placeholder `SkillTool_LoadUnloadCycle` with a real temp-home list/load/unload test.
- [x] Run targeted tests while iterating.
- [x] Run `go test ./...` and `go vet ./...`.

## Validation Commands

```bash
go test ./internal/config ./internal/tools ./tests
go test ./...
go vet ./...
```

## Rollback Points

- Revert only the changed tests and `internal/config/merge.go` if the merge behavior conflicts with a documented runtime decision.
