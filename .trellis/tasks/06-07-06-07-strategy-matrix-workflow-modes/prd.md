# Strategy matrix workflow modes

## Goal

Add a Home strategy matrix inspired by Shannon/Kocoro execution strategies for choosing Astria workflow modes.

## Requirements

- Reuse the embedded static Web UI architecture; do not add frontend dependencies or backend endpoints.
- Add a Home "Strategy Matrix" that lets users choose an execution mode before launching a task.
- Reflect Kocoro/Shannon-style modes as Astria-native options: quick execution, research brief, council/swarm, guarded approval, memory capture, and tool/MCP setup.
- Selecting a strategy should update Home mode, mission prompt text, workflow stage label, and Focus Brief context.
- Strategies should route to relevant existing panels when appropriate: Runs, Council, Inbox/permissions, Memory, or MCP.
- The matrix should feel like an operational control surface, not a marketing/dashboard card.
- Keep product-facing copy as Astria while preserving StarClaw internals.

## Acceptance Criteria

- [x] Home renders a Strategy Matrix section.
- [x] At least six strategies are available and visually distinct.
- [x] Selecting a strategy updates the mission prompt and Focus Brief.
- [x] Strategy route actions navigate to existing panels.
- [x] Core smoke verifies rendering, selection, and one route action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No backend strategy execution engine.
- No persistent strategy preferences.
- No rename of CLI/module/config paths from StarClaw to Astria.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
