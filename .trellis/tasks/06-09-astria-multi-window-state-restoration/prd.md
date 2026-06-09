# Astria multi-window state restoration

## Goal

Improve Astria's multi-window route restoration and lifecycle behavior while
preserving same-origin `/app` route safety.

## Requirements

- Keep each window usable around the local daemon Web UI.
- Preserve route safety for restored and copied routes.
- Avoid storing full external origins.
- Keep daemon supervision and Desktop RPC lifecycle shared and local.

## Acceptance Criteria

- [ ] Multi-window route/state behavior is documented and smoke-tested.
- [ ] Unsafe routes still fall back to `/app/`.

## Notes

- Deep per-window session identity can be expanded later.
