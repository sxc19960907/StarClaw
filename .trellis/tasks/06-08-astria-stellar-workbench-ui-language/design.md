# Astria stellar workbench UI language

## Scope

This task polishes the embedded daemon Web UI language after the hard phase-3 backend slices have landed. It must keep the application dense and operational. It must not introduce external assets, a frontend build step, or a marketing landing page.

## Visual grammar

- **Orbit rail**: navigation, quick actions, and panel card lists use slim vertical or radial rail markers to show route and status.
- **Constellation map**: repeated planner cards use small star/line hints and restrained layered backgrounds, not decorative full-page artwork.
- **Score ring**: numeric quality or readiness scores use compact circular/ring treatment for fast scanning.
- **Route glyph**: cards carry a short glyph/kicker that identifies the workbench lane:
  - `Q` quality / evidence
  - `B` budget / stop rules
  - `S` snapshot / resume packs
  - `R` results
  - `P` playbooks
- **Color grammar**:
  - Cyan / teal: evidence, health, completion.
  - Amber: budget, review, caution.
  - Slate-blue: snapshot, route, continuity.
  - Pink used sparingly for launch/accent only.

## Implementation boundary

- Reuse the existing embedded files:
  - `internal/daemon/webui/index.html`
  - `internal/daemon/webui/assets/styles.css`
  - `internal/daemon/webui/assets/app.js`
- Add reusable CSS classes instead of copying more per-panel styling.
- Keep cards at 8px radius, preserve dense split panels, and avoid nested cards.
- Ensure text uses `overflow-wrap:anywhere` where dynamic run/session content can be long.

## Target surfaces

- Home: make the first viewport clearly Astria workbench oriented with orbit/system language.
- Run Quality: use score ring and evidence-lane card styling.
- Budget Guard: use amber lane styling and stop-rule route markers.
- Workspace Snapshot: use route/continuity lane styling.
- Result/Playbook panels: keep compatible with existing lists but share route glyph/kicker conventions where generated cards support it.

## Validation

- `go test ./internal/daemon -run 'TestWebApp|TestRouterRegistersRoutes' -count=1`
- `go test ./...`
- `git diff --check`
- Manual CSS scan for no external asset references or build pipeline additions.
