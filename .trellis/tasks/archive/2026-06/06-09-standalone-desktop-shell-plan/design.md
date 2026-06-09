# Standalone desktop shell plan design

## Decision

Use a macOS-native SwiftUI app shell as the first standalone app route.

The shell should be thin:

- Bundle or resolve a `starclaw` daemon binary.
- Start or attach to the local daemon.
- Host `http://127.0.0.1:7533/app/` in a web view.
- Surface daemon readiness, diagnostics, and recovery states.

This keeps the Web UI as the product surface and aligns with Kocoro's Desktop
parent-process model without forcing a new frontend stack.

## Alternatives

- Tauri: useful later if cross-platform desktop packaging becomes the priority,
  but it adds Rust and packaging surface before the launcher contract is
  stable.
- Electron: natural Web UI host but too heavy for a local-first shell whose
  main missing capability is daemon supervision.
- CLI-only browser launch: already implemented; not enough for Kocoro
  standalone app parity.

## Repository Shape

Recommended initial layout:

- `desktop/macos/Astria/` for the app project and Swift sources.
- `desktop/macos/Astria/Resources/` for app assets and local development
  metadata.
- Keep Go daemon and Web UI under existing paths.
- Add scripts only when they can run without private signing credentials.

## Launch Contract

Stage 1 should use the existing HTTP readiness contract:

- `/health` for daemon readiness.
- `/status` for version and runtime summary.
- `/diagnostics` for user-visible troubleshooting context.
- `/app/` as the hosted UI route.

Stage 2 can add explicit `--rpc-socket` and `--rpc-pidfile` flags plus
Desktop RPC reconciliation if the launcher child determines that HTTP health
alone is insufficient for crash/version handling.

## Compatibility

- Existing `starclaw app` remains the browser fallback.
- Existing daemon-only `starclaw daemon start` remains valid.
- The app shell must be optional and must not become a prerequisite for local
  API usage.
- All cloud behavior remains opt-in and outside this task.
