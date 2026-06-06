# Design

## Architecture Boundary

This parent task is a planning and coordination artifact. It does not own direct implementation except for maintaining task metadata and planning documents. Feature work belongs to child tasks.

## Product Surface Principles

- Astria should feel like a focused desktop agent workspace: a central task composer, native sidebar navigation, clear activity state, and minimal modal friction.
- Kocoro is a reference for interaction density and shell style, not a source to clone exactly.
- Astria's differentiator is the mission/orbit/constellation language: tasks, agents, tools, and memories can be represented as connected points, but this must remain functional and restrained.

## Cross-Child Contracts

- Web UI panels should preserve existing IDs and aria labels used by smoke tests unless tests are intentionally updated in the same child task.
- Backend features should expose stable daemon APIs before the UI depends on them.
- Tool additions should register through existing tool patterns and include direct unit tests.
- Long-running actions need observable run state and permission/approval boundaries.
- New surfaces should degrade cleanly when a backend feature is unavailable.

## Sequencing

1. Finish and verify the current Astria Web UI Home task.
2. Implement `astria-home-actions` so Home becomes the product hub for later capabilities.
3. Implement `document-archive-tools` and `mcp-starport-ui` as high-value, concrete capability additions.
4. Implement `memory-map-mvp` once session extraction/indexing contracts are clear.
5. Implement `agent-council-swarm` after orchestration and UI state patterns are mature.
6. Implement `channel-inbox-mvp` after daemon/background reliability and approval handoff are sufficient.

## Rollout

Ship children incrementally. The parent can remain open while child tasks land. Incomplete child features should appear as disabled or labeled future surfaces only if that improves discoverability and does not create dead ends.
