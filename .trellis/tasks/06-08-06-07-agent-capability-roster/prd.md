# Agent capability roster

## Goal

Make named agents feel like Kocoro-style first-class collaborators by showing their capability posture before the operator opens the editor.

## Requirements

- Add an Agents panel roster that summarizes each named agent's model, reasoning effort, memory, tool allow/deny counts, auto-approve posture, heartbeat status, and command count.
- Reuse existing `/agents` data; do not add a new daemon endpoint for this slice.
- Keep StarClaw internal names unchanged and use Astria only as product-facing UI language where needed.
- Preserve the embedded static Web UI architecture with no frontend build pipeline or new dependencies.
- Keep the roster dense and operational, with subtle celestial styling that fits the current Astria interface.

## Acceptance Criteria

- [x] Agents panel renders an explicit capability roster above the existing agent list.
- [x] Empty agent state is clear in both the roster and the list.
- [x] Each roster row/card shows model/reasoning, memory, tools allowed/denied, auto-approve, heartbeat, and commands.
- [x] Roster actions reuse the existing agent edit flow.
- [x] Web UI smoke coverage verifies the roster after creating and updating `smoke-agent`.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
