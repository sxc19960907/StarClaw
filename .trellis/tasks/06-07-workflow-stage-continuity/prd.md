# Workflow stage continuity

## Goal

Add an explicit Home workflow stage rail so Astria shows whether the current workspace is in Draft, Running, Review, or Memory stage, tying recipes, launched runs, Mission Control, and Memory Map into a continuous workflow.

## Requirements

- Reuse existing Home, run, approval, and memory state; do not add backend endpoints.
- Show a compact stage rail on Home near the mission composer.
- Stages must cover Draft, Running, Review, and Memory.
- Selecting a recipe should put the rail into Draft with the selected workflow label.
- Launching from Home should move the rail toward Running.
- Loaded run state should update the rail to Running when work is active, Review when work is failed/completed/unknown, and Memory when memory warnings or facts are present.
- Keep Mission Control filters semantically correct; do not let Home run-health grouping conflict with Mission Control grouping.
- Keep the UI dense and operational, consistent with existing Astria styling.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Home renders a workflow stage rail with Draft, Running, Review, and Memory stages.
- [x] Selecting a workflow recipe marks Draft as active and shows the recipe title.
- [x] Home run state updates the rail using existing loaded runs.
- [x] Mission Control active/attention/completed filters remain correct after status grouping cleanup.
- [x] Core smoke verifies the stage rail and recipe Draft state.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No persisted workflow state.
- No backend workflow model.
- No replacement of the Runs or Memory panels.
