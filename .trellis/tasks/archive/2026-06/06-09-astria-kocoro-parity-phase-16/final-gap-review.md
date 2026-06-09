# Phase16 final gap review

## Scope closed

Phase16 moved Astria further from a thin WebView shell toward native macOS app
product readiness:

- Astria now has native command/window behavior for New Window, Reload Astria,
  Open Diagnostics, Export Diagnostics, and Retry Daemon.
- The command model is smoke-testable and wired to SwiftUI commands without
  changing daemon/browser fallback paths.
- Astria can export local-only diagnostics reports with redaction for API keys,
  bearer tokens, Desktop RPC socket paths, and pidfile paths.
- Release validation now includes a credential-free Astria distribution
  boundary check for local shell artifacts, absence of committed signing /
  notarization private material, and unavailable-safe updater metadata.
- Install/release/Astria shell docs and backend specs now capture signing,
  notarization, updater, diagnostics export, and native command boundaries.

## Evidence

- `fc18ae0 feat: add Astria native app commands`
- `3814bb9 feat: export Astria diagnostics reports`
- `0ad3d36 feat: harden Astria distribution boundary`
- Validation performed during Phase16 children:
  - `scripts/smoke_macos_astria_shell.sh`
  - `scripts/validate_release_artifacts.sh --npm-only --astria-local`
  - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./...`
  - `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 90-93% aligned with Kocoro for local-first desktop
lifecycle and baseline native app shell behavior. The remaining gap is mostly
the depth of native OS capabilities and production release execution, not the
architecture of the local daemon/app boundary.

## Remaining Kocoro gaps

- Deeper native OS integrations: notifications, clipboard/file affordances,
  TCC permission helper flows, and richer Dock/window state restoration.
- Real crash reporter integration: local crash capture, structured crash
  summaries, and optional user-approved export/upload workflows.
- Actual signed/notarized release production with external Apple credentials,
  stapling, and release artifact publication.
- A signed updater implementation with checksum/signature verification and
  app/daemon compatibility enforcement.

## Recommended next phase

Phase17 should focus on native OS tool depth:

1. notification and clipboard/file native affordances;
2. TCC-style permission helper flows for desktop tools;
3. richer multi-window state restoration and app lifecycle polish.
