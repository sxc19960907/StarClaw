# Validate named agent and skills workflow

## Goal

Validate named agent loading, agent-scoped sessions, memory, and skills workflow with deterministic tests.

## Requirements

- Verify named agent definitions load as a complete unit: `AGENT.md`, `MEMORY.md`, `config.yaml`, and `commands/*.md`.
- Verify agent model overrides and tool allow/deny filters from `config.yaml` affect the runtime config path used by chat execution.
- Verify agent sessions can be persisted and resumed from the named agent's own `sessions/` directory without touching global sessions.
- Verify skill activation works against an isolated StarClaw home directory for both `use_skill` and the `skill` management tool.
- Keep tests deterministic and offline; no provider credentials, network calls, or real user config may be required.
- Leave unrelated untracked workspace files untouched.

## Acceptance Criteria

- [x] Unit or integration tests cover named agent load + memory + command + config behavior.
- [x] Tests cover agent-scoped session save/resume under `<starclaw>/agents/<name>/sessions`.
- [x] Tests cover `use_skill` activation and `skill` list/load/unload against a temporary `~/.starclaw/skills`.
- [x] Agent tool allow/deny config is propagated into merged runtime config and covered by regression tests.
- [x] `go test ./...` and `go vet ./...` pass.

## Notes

- Current scope is validation plus small fixes required to make the validated workflow behave as documented by existing structs.
