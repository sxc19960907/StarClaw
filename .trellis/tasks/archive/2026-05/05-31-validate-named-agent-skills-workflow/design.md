# Design

## Boundaries

- `internal/agents` remains responsible for reading named agent files and parsing agent-local config.
- `internal/config.MergeAgentConfig` is the boundary where parsed agent-local config becomes runtime config.
- `internal/session` already supports arbitrary session roots; tests will validate the agent-scoped path by using a manager rooted at `<starclaw>/agents/<agent>/sessions`.
- `internal/tools` remains responsible for skill listing and activation. Tests will isolate it by setting `HOME` before constructing tools.

## Data Flow

1. Named agent files are created under a temporary StarClaw home.
2. `agents.LoadAgent` reads prompt, memory, custom commands, model overrides, tool filters, and auto-approve.
3. `config.MergeAgentConfig` overlays model and tool-filter values onto a global config copy.
4. A local tool registry can be filtered from the merged allow/deny values.
5. Session manager persists under the named agent directory.
6. Skill tools load `SKILL.md` from the temporary `~/.starclaw/skills` directory.

## Compatibility

- No new config keys are introduced.
- Existing nil/empty merge behavior remains unchanged unless an agent explicitly sets tool allow or deny lists.
- Tests avoid live LLM providers and do not use the developer's real home directory.

## Trade-Offs

- This task adds focused regression coverage rather than a broad CLI black-box test because the CLI path requires provider setup. The covered components are the same boundaries used by `cmd/root.go`.
