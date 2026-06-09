# Native desktop diagnostics and recovery UX design

## Current Shape

Astria already has:

- `DaemonState.bannerMessage` for daemon health/failure banners.
- `StatusBanner` with Diagnostics and conditional Retry actions.
- `LaunchStateView` for non-attached startup/failure states.
- `DesktopRPCSessionState` from the Phase15 lifecycle child, but it is not yet
  surfaced directly in the native UI.

## Proposed Shape

Add a small `DesktopRPCDiagnostics` helper that maps session state into:

- optional banner message;
- diagnostic severity label for future UI expansion;
- retry eligibility through the existing daemon retry action.

`AstriaRootView` should prefer WebView load state, then daemon banner, then
Desktop RPC diagnostics. That keeps critical daemon failures more visible while
still showing RPC degradation when the WebView is otherwise healthy.

## Redaction

Diagnostic copy must use high-level state and error messages already produced
by validation helpers. It must not include socket/pidfile paths or event
payloads.

## Test Strategy

Add Swift smoke coverage under a new or existing Desktop RPC smoke path for:

- connected state has no warning banner;
- reconnecting state produces a retrying banner;
- degraded state produces fallback copy;
- mismatch state produces compatibility copy.
