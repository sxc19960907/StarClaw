# Starter kit launcher

## Goal

Add an Astria Starter Kit Launcher that narrows the Kocoro "prebuilt success" gap by offering curated local starter workflows for common agent tasks.

## Requirements

- Preserve StarClaw CLI/module/package/release names while keeping product-facing UI copy as Astria.
- Keep the embedded static daemon Web UI architecture; do not add a frontend build step, external assets, cloud marketplace, account system, or backend template store in this slice.
- Add a dedicated Starter Kit Launcher surface that exposes prebuilt kits for common local work patterns.
- Kits must combine prompt shape, suggested route, evidence expectations, reuse path, and safety/review boundaries.
- Starter kits must draft concrete Chat prompts and route to existing Astria surfaces such as Chat, Agents, Data Planner, Browser Planner, Share Pack, Reuse Gallery, and Memory.
- Preserve the dense operational UI direction with compact controls and subtle celestial styling.

## Acceptance Criteria

- [x] Manage and sidebar navigation expose a Starter Kits entry with a live count.
- [x] The launcher renders at least six curated kits and a detail pane that explains route, evidence, reuse path, and safety boundary.
- [x] Each kit can draft a concrete Chat prompt that includes objective, suggested agent posture, source/evidence plan, review gate, and reusable output.
- [x] At least one kit routes to Browser Planner, one to Data Planner, one to Share Pack, and one to Memory or Reuse Gallery.
- [x] Web UI smoke coverage opens Starter Kits, verifies kit cards/detail content, drafts a Chat prompt, and checks routes to Browser/Data/Share/Memory-or-Reuse.
- [x] Targeted and full checks pass for the edited UI and daemon smoke path.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
