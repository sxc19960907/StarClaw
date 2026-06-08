# Share pack builder

## Goal

Add an Astria Share Pack Builder that narrows the Kocoro team-reuse gap by packaging successful local work into a reviewed, copyable handoff starter without adding cloud accounts or sync.

## Requirements

- Preserve StarClaw CLI/module/package/release names while keeping product-facing UI copy as Astria.
- Keep the embedded static daemon Web UI architecture; do not add a frontend build step, external assets, account system, or cloud sharing in this slice.
- Add a dedicated Share Pack Builder surface that lets the user name a package, choose an audience, and describe the handoff intent.
- Generate local share pack cards for mission brief, evidence bundle, reusable prompt, knowledge handoff, and review checklist.
- Each card must draft a concrete Chat prompt that includes package name, audience, handoff intent, included artifacts, review constraints, and local-only sharing boundaries.
- Pack routes should reuse existing Astria surfaces such as Reuse Gallery, Runs, Memory, Comparison, and Data Planner.
- Preserve the dense operational UI direction with compact controls and subtle celestial styling.

## Acceptance Criteria

- [x] Manage and sidebar navigation expose a Share Pack entry with a live count.
- [x] The builder accepts package name, audience, and handoff intent, renders five share pack cards, and updates detail text from the current inputs.
- [x] Each share pack card can draft a Chat prompt with local-only boundaries, evidence expectations, and review steps.
- [x] At least one card routes reusable assets to Reuse Gallery and one routes durable facts to Memory.
- [x] Web UI smoke coverage opens Share Pack, fills inputs, verifies five cards/detail content, drafts a Chat prompt, and routes to Reuse Gallery and Memory.
- [x] Targeted and full checks pass for the edited UI and daemon smoke path.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
