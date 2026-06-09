# Astria Kocoro parity phase 16: native OS integration and distribution hardening

## Goal

Close the next Kocoro parity gap after Phase15 by moving Astria from a thin
single-window shell toward a product-grade macOS app boundary: native
menu/Dock/window behavior, diagnostics/crash report export, and signed release
pipeline guardrails.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase15 completed long-lived Desktop RPC session monitoring, bounded
  reconnect/degraded fallback, daemon-side `desktop_event` metadata, and native
  Desktop RPC diagnostic banners.
- Phase15 final gap review estimates Astria is roughly 88-92% aligned for
  local-first desktop lifecycle behavior.
- Remaining gaps are broader native OS integration and release/distribution
  maturity, not basic daemon/RPC lifecycle.
- Existing CLI/browser fallback paths remain part of the product contract.

## Child Plan

1. `astria-native-menu-dock-window`: add native app commands and window/Dock
   behavior that make Astria feel like an app rather than only a WebView.
2. `astria-native-diagnostics-crash-reports`: add local diagnostics export and
   crash/failure report boundaries without remote telemetry.
3. `astria-signing-notarization-updater-boundary`: harden signing,
   notarization, and updater pipeline scripts/docs/validation without requiring
   private credentials in local development.

## Requirements

- Preserve the daemon-served Web UI as the primary surface.
- Preserve `starclaw app`, `starclaw daemon start`, and browser fallback.
- Keep all diagnostics local; do not add Shannon Cloud auth, remote lifecycle,
  or off-machine telemetry.
- Do not commit Apple credentials, signing identities, update private keys, or
  notarization secrets.
- Add smoke/test coverage for each native/productization boundary.

## Acceptance Criteria

- [ ] Each child task has independent PRD/design/implementation artifacts and
      testable acceptance criteria before implementation starts.
- [ ] Astria exposes native menu/window commands for common app actions.
- [ ] Astria provides local diagnostics/crash/failure export boundaries without
      leaking secrets or raw user content.
- [ ] Release/distribution validation documents signing/notarization/updater
      boundaries and remains usable without private credentials.
- [ ] Final gap review updates Kocoro parity and remaining native/product gaps.

## Out of Scope

- Cloud account lifecycle, Shannon Cloud auth, or remote telemetry.
- Full native rewrite of the Web UI.
- Shipping signed public artifacts from this local session.

## Notes

Parent task only. Start child tasks for implementation.
