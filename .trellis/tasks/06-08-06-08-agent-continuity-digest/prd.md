# Agent continuity digest

## Goal

Make each named agent feel persistent by summarizing its recent run continuity, memory posture, command surface, and next useful action in the Agents panel.

## Requirements

- Add an Agent Continuity Digest to the Agents panel that derives from existing agents, runs, and memory state.
- Show each agent's recent run count, latest run status, profile memory state, command count, and a concise next-step hint.
- Provide direct actions to continue with the agent, draft memory for the agent, and open its latest run when available.
- Do not add a new daemon endpoint or frontend build pipeline.
- Preserve existing roster, editor, command, Chat, Runs, and Memory behavior.

## Acceptance Criteria

- [x] Agents panel renders an Agent Continuity Digest above the capability roster.
- [x] Empty agent state is explicit.
- [x] Each digest item shows run continuity, memory posture, command count, and next-step hint.
- [x] Continue action switches to Chat, selects the agent, and drafts a continuity prompt.
- [x] Draft memory action switches to Memory and drafts an agent memory candidate.
- [x] Open latest run action opens the latest run when the agent has one.
- [x] Agents smoke verifies digest render and actions for `smoke-agent`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
