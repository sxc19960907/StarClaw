# Implementation Plan

## Parent Task Checklist

- [x] Create parent Trellis task.
- [x] Create child Trellis tasks.
- [x] Write parent PRD, design, and implementation plan.
- [x] Write child PRDs.
- [ ] Review child priorities after current Astria Web UI Home task is finished.
- [ ] Start the first child task only after its planning artifacts are approved.

## Recommended Execution Order

1. `06-06-astria-webui-home` - finish current UI foundation.
2. `06-06-astria-home-actions` - make the home launcher operational.
3. `06-06-document-archive-tools` - add immediate practical local-file value.
4. `06-06-mcp-starport-ui` - make integration setup visible and manageable.
5. `06-06-memory-map-mvp` - add durable project/user context.
6. `06-06-agent-council-swarm` - add multi-agent planning/review mode.
7. `06-06-channel-inbox-mvp` - add external channel entry point.

## Validation Gates

- Run existing Go tests for touched packages.
- Run Web UI smoke tests for any Web UI change.
- Add tool-specific tests for any new tool category.
- Confirm disabled/future UI states do not break navigation or smoke selectors.

## Rollback Points

- Each child task should remain revertible independently.
- Avoid schema migrations in early UI-only children.
- Feature-flag or hide incomplete surfaces rather than shipping partially wired controls.
