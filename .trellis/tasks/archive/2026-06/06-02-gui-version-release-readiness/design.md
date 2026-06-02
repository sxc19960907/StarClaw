# Design

## UI

- Extend `renderVersion()` to build a `Release readiness` card before the existing version metadata card.
- Use existing `row-item`, `run-meta-grid`, and tag styling.
- Use current `/version` fields only; no backend API change.

## Smoke

- Extend core smoke to assert the readiness card and development-build update support text.
