# Astria stellar workbench UI language

## Goal

Polish the Astria Web UI into a stronger project-specific stellar workbench language after the hard Kocoro parity functionality slices have landed.

## Requirements

- Defer this task until token budget enforcement, API gateway, routing/fallback, observability, and workflow control have dedicated implementation slices.
- Establish a reusable Astria visual language around orbit rails, constellation maps, score rings, route glyphs, and consistent celestial color grammar.
- Apply the style to high-impact surfaces first: Home, Run Quality, Budget Guard, Workspace Snapshot, and Result/Playbook reuse panels.
- Keep the UI operational and dense; do not turn the app into a marketing landing page.
- Do not add external assets or a frontend build pipeline.

## Acceptance Criteria

- [ ] UI language spec documents Astria-specific visual patterns and color grammar.
- [ ] Home, Run Quality, Budget Guard, and Workspace Snapshot share the same stellar workbench design system.
- [ ] Existing smoke tests pass and screenshots show no text overlap on desktop smoke viewport.
- [ ] No external asset or build pipeline is introduced.
- [ ] This task starts only after the preceding hard parity tasks have been addressed or explicitly deferred.

## Notes

- This is intentionally last in the phase 3 child order.
