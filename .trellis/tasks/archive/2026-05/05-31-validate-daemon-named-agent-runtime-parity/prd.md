# Validate daemon named-agent runtime parity

## Goal

Align daemon RunAgent named-agent behavior with CLI runtime: config merge, tool filters, memory injection, and deterministic tests.

## Requirements

- Daemon `RunAgent` must apply named agent `config.yaml` overrides consistently with CLI chat where the same runtime settings exist.
- Named agent `MEMORY.md` must be injected into daemon agent runs, not only CLI runs.
- Named agent `tools.allow` / `tools.deny` must filter the daemon tool registry for that run without mutating the shared base registry.
- Daemon startup must pass loaded global config into daemon dependencies so scheduler and HTTP runs can merge named-agent config.
- Default daemon agent behavior must remain compatible when no config is supplied in tests or legacy callers.
- Tests must be deterministic and offline.
- Leave unrelated untracked workspace files untouched.

## Acceptance Criteria

- [x] `RunAgent` named-agent tests verify prompt and `MEMORY.md` appear in the effective system prompt.
- [x] `RunAgent` named-agent tests verify agent `model`, `max_tokens`, `thinking`, `thinking_mode`, `thinking_budget`, and `reasoning_effort` reach the LLM request options; `max_iterations` is applied through loop configuration.
- [x] `RunAgent` named-agent tests verify allow/deny tool filters are applied per run and do not mutate the base registry.
- [x] Daemon startup wires loaded config into `ServerDeps`.
- [x] Existing default-agent and named-agent session tests continue to pass.
- [x] `go test ./internal/daemon ./cmd` passes.
- [x] `go test ./...` and `go vet ./...` pass.

## Notes

- Evidence from code inspection: CLI chat loads named agents, merges config, filters tools, applies agent options, and injects memory. Daemon `RunAgent` currently loads the named agent prompt and session only.
