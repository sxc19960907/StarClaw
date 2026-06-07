# Workflow Recipes Launcher Design

## UI

Add a `workflow-recipes` section under the Home composer and above the activity board. Use compact repeated buttons, not a full dashboard.

Each recipe includes:

- title
- short label/category
- prompt
- optional panel route

Selecting a recipe:

- sets `state.homeMode`
- updates the mode bar
- fills `#home-task-input`
- if a route exists, exposes the existing `home-mode-route` button

## JavaScript

Add a `workflowRecipes` constant and `selectWorkflowRecipe(id)` function. Reuse `seedMissionPrompt`, `renderHomeMode`, and `switchPanel`.

Do not add backend endpoints.

## Testing

Extend `scripts/lib/webui_smoke_common.sh` core flow to:

- click a code review recipe and assert the Home composer is filled.
- click a file intake recipe, assert the route button opens File Intake or that the route-aware UI is set.
