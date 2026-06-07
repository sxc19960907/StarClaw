# Workflow brief context

## Goal

Turn Home workflow recipes into visible Astria work briefs: when a recipe is selected, the user should see the workflow goal, required context, launch path, and next action instead of only receiving a prefilled prompt.

## Requirements

- Reuse the existing static embedded Web UI architecture.
- Keep existing recipe prompt prefill behavior.
- Add a compact work brief surface near the Home composer.
- Each recipe brief must show the workflow outcome, useful context/materials, the primary route, and a small next-step checklist.
- Route-aware recipes should expose a panel jump action from the brief.
- Keep the design dense and operational with subtle Astria/celestial language, not a marketing section.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Selecting a recipe renders a work brief with outcome, context, route, and checklist.
- [x] Selecting another recipe updates the brief without leaving Home.
- [x] Route-aware recipes provide a visible action that opens the associated panel.
- [x] Existing prompt prefill and start behavior continue to work.
- [x] Core smoke verifies the brief and route action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No backend API changes.
- No persisted workflow brief state.
- No redesign of Chat, Runs, or File Intake internals.
