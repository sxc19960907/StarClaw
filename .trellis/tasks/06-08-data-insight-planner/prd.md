# Data insight planner

## Goal

Add an Astria Data Insight Planner that narrows the Kocoro data-analysis gap by turning local files, tables, metrics, and research exports into reviewed analysis mission starters.

## Requirements

- Preserve StarClaw CLI/module/package/release names while keeping product-facing UI copy as Astria.
- Keep the embedded static daemon Web UI architecture; do not add a frontend build step, external visual assets, or backend data parsing in this slice.
- Add a dedicated Data Insight Planner surface that helps the user describe a source, analysis question, and output format.
- Provide reusable insight mission cards for profiling, trend analysis, anomaly review, visual summary, and knowledge capture.
- Generated drafts must route to existing Astria surfaces such as Chat, Comparison, Memory, and Reuse Gallery without pretending to execute data analysis directly in the browser.
- Preserve the dense operational UI direction with subtle celestial naming and compact controls.

## Acceptance Criteria

- [x] Manage and sidebar navigation expose a Data Planner entry with a live count.
- [x] The planner accepts a source descriptor and analysis question, renders five mission cards, and updates detail text from the current inputs.
- [x] Each mission card can draft a concrete Chat prompt that includes the source descriptor, question, output expectation, review constraints, and non-destructive behavior.
- [x] At least one planner action routes knowledge capture to Memory and one routes reusable findings to Reuse Gallery.
- [x] Web UI smoke coverage opens Data Planner, fills inputs, verifies mission cards/detail content, drafts a Chat prompt, and routes a card to Memory or Reuse Gallery.
- [x] Targeted and full tests pass for the edited UI and daemon smoke path.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
