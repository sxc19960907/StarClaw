# Astria Kocoro parity phase 3

## Goal

Close the next major Kocoro/Shannon parity gap by moving Astria from primarily local workbench/UI parity toward platform-level runtime capabilities: hard budget enforcement, API compatibility, runtime routing/fallback, structured observability, workflow control, and only then a dedicated Astria stellar UI language pass.

## Requirements

- Preserve StarClaw CLI/module/package/release names; Astria remains the product-facing Web UI name.
- Keep UI polish explicitly last. Do not spend this phase on visual styling until hard runtime/API/observability gaps have dedicated slices.
- Prioritize Kocoro/Shannon-like platform capabilities over more planning-only panels:
  - token budget tracking and hard-stop enforcement
  - OpenAI-compatible API gateway surface
  - runtime complexity routing and model fallback
  - structured events, metrics, and tracing
  - workflow control API for pause/resume/cancel/replay
  - Astria stellar workbench UI language after the above hard slices
- Each child must be independently testable and must state whether it is a backend/runtime slice or UI polish.
- Maintain the embedded daemon Web UI architecture unless a child explicitly scopes a new API endpoint.

## Acceptance Criteria

- [ ] Parent task lists hard-function child tasks before UI style work.
- [ ] Each child task has PRD acceptance criteria and clear non-goals.
- [ ] UI style task is last in the child map and marked as deferred polish.
- [ ] Each implemented child passes targeted tests and relevant smoke coverage.
- [ ] The phase produces measurable backend/API/runtime parity improvements, not only new UI panels.

## Notes

- Kocoro/Shannon parity remains incomplete until hard runtime/API/observability gaps are implemented and verified.
