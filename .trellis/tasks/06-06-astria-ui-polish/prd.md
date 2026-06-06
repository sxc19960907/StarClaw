# Astria UI Polish

## Goal

Polish the existing Astria Web UI so the MVP surfaces feel like one coherent desktop application: consistent navigation, panel density, copy, activity state, empty states, and responsive behavior.

## Requirements

- Keep the embedded static Web UI architecture.
- Improve visible product cohesion without adding new product scope or external dependencies.
- Preserve the Kocoro-inspired calm native layout and Astria's subtle star/constellation identity.
- Make Home expose the current major product surfaces consistently, including Inbox.
- Improve panel empty states and action affordances where current surfaces still read as raw forms.
- Keep text inside controls stable on desktop and narrow widths.
- Maintain existing API contracts.

## Acceptance Criteria

- [x] Home, Manage, and sidebar counts are consistent for major surfaces.
- [x] Inbox appears as a first-class docked tool from Home, not only sidebar/manage.
- [x] Core panels have polished empty states and concise action affordances.
- [x] Narrow viewport layout avoids obvious overlap or clipped controls.
- [x] Web UI smoke passes.
- [x] `node --check internal/daemon/webui/assets/app.js`, `git diff --check`, and `go test ./internal/daemon` pass.

## Non-Goals

- No new backend feature behavior.
- No React/Vite/build pipeline.
- No redesign away from the current light desktop shell.
- No new external assets.

## Dependencies

- Builds on the committed Astria roadmap surfaces in `3142d9c`.
