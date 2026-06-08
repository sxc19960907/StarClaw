# Citation grounding planner

## Goal

Add an Astria Citation Grounding Planner that narrows the Kocoro knowledge reliability gap by planning source coverage, citation needs, and evidence gaps before an answer, data brief, browser review, or handoff is drafted.

## Requirements

- Preserve StarClaw CLI/module/package/release names while keeping product-facing UI copy as Astria.
- Keep the embedded static daemon Web UI architecture; do not add a frontend build step, external assets, citation database, account system, or backend evidence store in this slice.
- Add a dedicated Citation Grounding Planner surface that lets the user describe a claim or answer, choose a source posture, and state the required evidence level.
- Generate grounding cards for source coverage, claim mapping, quote/evidence capture, freshness risk, and gap escalation.
- Each card must draft a concrete Chat prompt that includes claim scope, source posture, evidence level, citation rules, gaps to report, and local-only review boundaries.
- Planner routes should reuse existing Astria surfaces such as Source Registry, Memory, Browser Planner, Data Planner, Share Pack, and Reuse Gallery.
- Preserve the dense operational UI direction with compact controls and subtle celestial styling.

## Acceptance Criteria

- [x] Manage and sidebar navigation expose a Citation Planner entry with a live count.
- [x] The planner accepts claim scope, source posture, and evidence level, renders five grounding cards, and updates detail text from the current inputs.
- [x] Each grounding card can draft a Chat prompt with citation rules, evidence expectations, uncertainty handling, and local-only boundaries.
- [x] At least one card routes to Source Registry, one to Browser Planner, one to Data Planner, and one to Memory or Share Pack.
- [x] Web UI smoke coverage opens Citation Planner, fills inputs, verifies five cards/detail content, drafts a Chat prompt, and checks routes to Source/Browser/Data/Memory-or-Share.
- [x] Targeted and full checks pass for the edited UI and daemon smoke path.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
