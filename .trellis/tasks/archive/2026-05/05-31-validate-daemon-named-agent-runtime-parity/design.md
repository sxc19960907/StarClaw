# Design

## Scope

Bring daemon named-agent execution closer to CLI execution for runtime behavior that exists in shared internal packages:

- config merge
- tool allow/deny filtering
- model and thinking request options
- max iteration/token/result settings
- named agent prompt and memory injection

## Data Flow

1. `cmd daemon start` loads global config and passes it into `daemon.ServerDeps`.
2. `RunAgent` loads a named agent when `RunAgentRequest.Agent` is set.
3. `RunAgent` starts from `deps.Config` when present, otherwise a conservative default config.
4. `config.MergeAgentConfig` overlays agent config.
5. A per-run registry is derived from `deps.Registry` using merged tool allow/deny lists.
6. The agent loop is configured from merged config.
7. The named agent prompt and memory are injected into the loop before running.

## Boundaries

- Keep CLI-only concerns out of `internal/daemon`; do not import `cmd` helpers.
- Add small internal daemon helpers near runner code instead of exporting CLI helpers.
- `ServerDeps.Config` is optional for compatibility with existing tests and callers.
- Do not mutate `deps.Registry`; filtering returns derived registries.

## Compatibility

- Default daemon runs still work when `deps.Config == nil`.
- Existing test mocks continue to work.
- No provider/network calls are introduced.
