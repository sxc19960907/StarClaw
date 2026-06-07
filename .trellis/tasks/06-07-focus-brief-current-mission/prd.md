# Focus brief current mission

## Goal

Add a Home Focus Brief that summarizes the current mission context from selected workflow, active stage, latest run, latest session, and suggested next action.

## Requirements

- Reuse existing Home, workflow, session, run, and memory state; do not add backend endpoints.
- Render a compact Focus Brief on Home near the Workspace Hub.
- The brief should show current stage, mission title, latest session/run context, and next suggested action.
- If a workflow recipe is selected, the brief should reflect that recipe as the current mission.
- If recent run/session data exists, the brief should offer direct resume/open actions.
- Keep the styling dense, operational, and consistent with Astria workspace UI.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Home renders a Focus Brief.
- [x] Selecting a workflow recipe updates the Focus Brief title/stage.
- [x] Loaded recent session/run data updates the Focus Brief context.
- [x] Focus Brief actions can open the latest session or run when available.
- [x] Core smoke verifies Focus Brief rendering and recipe update.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No persisted focus state.
- No backend current-mission endpoint.
- No replacement of Workspace Hub or Workflow Brief.
