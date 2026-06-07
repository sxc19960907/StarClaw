# Command center palette

## Goal

Add an Astria Command Center palette so users can quickly launch workflows, jump to panels, and open workspace actions from one unified app-level entry point.

## Requirements

- Reuse existing static embedded Web UI architecture.
- Add a visible topbar Command Center button.
- Support `Ctrl+K` / `Cmd+K` to open the palette.
- Support text filtering over commands.
- Commands should cover workflow recipes, major panels, and high-value Home actions.
- Executing a command should reuse existing handlers such as `selectWorkflowRecipe`, `switchPanel`, and `runHomeAction`.
- Escape or backdrop click should close the palette.
- Keep the UI dense and operational, consistent with Astria styling.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Topbar exposes a Command Center entry.
- [x] `Ctrl+K` / `Cmd+K` opens the palette.
- [x] Palette filters commands by typed text.
- [x] Selecting a recipe command preloads the recipe and closes the palette.
- [x] Selecting a panel command switches to that panel and closes the palette.
- [x] Core smoke verifies opening, filtering, and executing a command.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No persisted command history.
- No fuzzy ranking beyond simple text matching.
- No backend command endpoint.
