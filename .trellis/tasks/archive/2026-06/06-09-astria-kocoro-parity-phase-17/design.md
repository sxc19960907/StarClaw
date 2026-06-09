# Astria Kocoro parity phase 17 design

## Architecture Boundary

Phase17 keeps Astria as a local macOS shell around the daemon Web UI. Native
affordances must be explicit user actions and should operate on already-redacted
local artifacts or safe route URLs.

## Native Areas

- Clipboard/file affordances: copy current route, copy diagnostics summary,
  reveal exported diagnostics, and redacted report handoff.
- Permission helper flows: explain local desktop permissions and current
  availability without requiring signed entitlements.
- Multi-window restoration: keep independent windows usable while preserving
  same-origin `/app` route safety.

## Compatibility

- Existing CLI/browser launch remains unchanged.
- Existing Desktop RPC session/event plumbing remains unchanged.
- Native actions must not require cloud services.

## Rollout

1. Clipboard/file affordances.
2. Permission helper boundaries.
3. Multi-window restoration.
