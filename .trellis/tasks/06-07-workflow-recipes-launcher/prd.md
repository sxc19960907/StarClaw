# Workflow Recipes Launcher

## Goal

Add guided workflow recipes to Astria Home so users can start common agent workflows without manually composing prompts or hunting through panels.

## Requirements

- Add a compact recipe launcher to the Home surface.
- Recipes should connect existing capabilities rather than introduce backend complexity.
- Each recipe should prefill the Home composer, update the mission mode, and optionally route to the relevant panel.
- Include recipes for code review, feature planning, file intake, research brief, MCP setup, inbox triage, and memory update.
- Preserve existing Home shortcuts and cards.
- Keep the UI calm, dense, and desktop-app-like.

## Acceptance Criteria

- [x] Home shows workflow recipes as a first-class section.
- [x] Selecting a recipe preloads a concrete prompt into the Home composer.
- [x] Recipes can route to relevant panels where useful.
- [x] Smoke coverage verifies at least two recipes: one prompt-only and one route-aware recipe.
- [x] JS syntax check and core Web UI smoke pass.

## Non-Goals

- No saved user-custom recipes.
- No backend persistence.
- No model-side execution until the user launches the task.

## Dependencies

- Depends on Astria Home launcher and second-phase workflow surfaces.
