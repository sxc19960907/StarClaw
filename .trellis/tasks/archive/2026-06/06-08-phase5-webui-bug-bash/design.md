# Phase 5 Web UI bug bash design

## Scope

Perform a focused bug bash on embedded Astria Web UI runtime validation surfaces. The target is not a redesign; it is to make Phase 3/4 platform capabilities inspectable and less fragile from the static app.

## Target Surfaces

- Mission Control run list and run detail.
- Runtime recovery, workflow steps, control decisions, and trace panels.
- Budget/routing/fallback/quality state summaries.
- Quality, reuse, share, memory, and related cards that draft prompts or navigate to run detail.
- Empty and error states for trace/API failure paths.

## Approach

1. Inspect current app.js rendering and event delegation for broken hooks, unsafe assumptions, and missing empty/error states.
2. Fix small integration defects with existing component patterns.
3. Add static smoke assertions in Go tests for any new hooks or rendering guards.
4. Use Playwright/browser checks only if a defect requires visual/runtime confirmation beyond static tests.

## Safety

The embedded architecture stays unchanged: static HTML/CSS/JS served by the daemon, no frontend build step, and no new backend runtime behavior unless a UI bug exposes a missing API contract.

## Rollback

Revert the changed static assets/tests and task artifacts. No persistent runtime migrations are introduced.
