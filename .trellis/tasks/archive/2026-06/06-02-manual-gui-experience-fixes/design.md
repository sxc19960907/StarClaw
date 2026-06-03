# Design

## Frontend Only

- Use existing diagnostics state in `loadDiagnostics()` to update the chat empty-state copy.
- Update `renderVersion()` and `checkForUpdates()` to respect unsupported update checks.
- Adjust existing CSS grid sizing for `#panel-agents`.
- Reorder Agent editor HTML so Test Runner is above command/import controls.
- Add a small `hideToast()` helper and call it on panel changes.

## Validation

- Run JS syntax check.
- Run core smoke for diagnostics/version/chat effects.
- Run agents and permissions smoke for layout-adjacent workflows.
- Run full smoke once after fixes.
