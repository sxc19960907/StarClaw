# Astria Kocoro parity phase 16 design

## Architecture Boundary

Phase16 stays within the optional macOS Astria shell and release tooling. The
daemon remains the runtime owner for the Web UI, agent execution, approvals,
permissions, sessions, Desktop RPC broker, and local status APIs.

## Native Integration Areas

- App/window commands: native menu entries for creating windows, reloading the
  WebView, opening diagnostics, and retrying daemon supervision.
- Diagnostics/crash reports: local export bundles and failure summaries that
  redact unsafe paths/secrets/raw payloads.
- Distribution boundary: scripts/docs/validators for signed/notarized builds
  and update metadata, without embedding private credentials.

## Compatibility

- Unsigned local builds remain supported.
- Existing smoke scripts stay the primary local verification path.
- Native commands should be testable with smoke-only models where direct UI
  automation is unnecessary.

## Rollout

1. Land native app commands/window shell behavior.
2. Land local diagnostics/crash report export.
3. Land distribution/updater boundary validation.
