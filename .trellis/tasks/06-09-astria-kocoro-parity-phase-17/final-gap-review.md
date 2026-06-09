# Phase17 final gap review

## Scope closed

Phase17 closed the next native OS tool-depth slice for Astria:

- Astria now has local native clipboard/file affordances for Copy Current Route,
  Copy Support Summary, and Reveal Diagnostics Folder.
- Clipboard support output reuses the diagnostics redaction boundary and copies
  only safe relative `/app` routes or local-only redacted summary text.
- Astria now has a Permission Help command and smoke-testable helper model for
  Calendar, Contacts, Reminders, file access, and notifications without
  silently requesting broad macOS privacy/TCC permissions.
- Astria now restores routes per native window using safe relative `/app`
  values, while preserving a shared safe route fallback for new windows.
- The macOS shell smoke covers native command metadata, diagnostics/export
  redaction, permission helper guidance, and multi-window route isolation /
  fallback behavior.

## Evidence

- `a022266 feat: add Astria clipboard file affordances`
- `424d663 feat: add Astria permission helper guidance`
- `575010d feat: restore Astria routes per window`
- Validation performed during Phase17 children:
  - `python3 ./.trellis/scripts/task.py validate <child-task>`
  - `scripts/smoke_macos_astria_shell.sh`
  - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./...`
  - `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 92-95% aligned with Kocoro for local-first desktop
platform behavior. The major Kocoro-style architecture boundaries are present:
daemon-hosted local app, OpenAI-compatible gateway, runtime controls, Desktop
RPC lifecycle, native commands, local diagnostics, local support affordances,
permission guidance, and multi-window safe route restoration.

The remaining gap is no longer the core local platform shape. It is production
native polish and release execution depth.

## Remaining Kocoro gaps

- Native notification depth: real notification permission request flow,
  notification routing, and user-visible notification preferences.
- Crash reporter depth: local crash capture, structured crash summaries, and
  optional user-approved export/upload workflows.
- Actual signed/notarized release production with external Apple credentials,
  Hardened Runtime, notarization, stapling, and artifact publication.
- Signed updater implementation with checksum/signature verification and
  app/daemon compatibility enforcement.
- Broader Desktop RPC native tool coverage beyond Calendar, including real
  Contacts/Reminders/file-provider tool surfaces once permission request flows
  are explicitly scoped.

## Recommended next phase

Phase18 should focus on production-native reliability and release readiness:

1. local crash reporter and crash-summary export;
2. notification permission/request/preferences and native notification smoke;
3. signed updater/release execution design with checksum/signature enforcement.
