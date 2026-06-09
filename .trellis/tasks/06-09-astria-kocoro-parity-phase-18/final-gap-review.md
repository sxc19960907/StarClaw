# Phase18 final gap review

## Scope closed

Phase18 moved Astria's native shell closer to production reliability while
keeping the app local-first and credential-free:

- Astria now has an Export Crash Summary command and local crash summary
  exporter. Reports are written under Astria diagnostics storage, redacted, and
  explicitly `localOnly=true` / `uploadReady=false`.
- Crash summary redaction now covers raw prompt-style fields in addition to API
  keys, bearer tokens, Desktop RPC socket/pidfile paths, and private local
  paths.
- Astria Permission Help now includes passive notification readiness from
  `UNUserNotificationCenter.getNotificationSettings`, with ready, blocked,
  requires-explicit-request, and unavailable-safe states.
- Notification readiness smoke proves passive checks do not promise
  `requestAuthorization`, do not send test notifications, and remain local.
- Release validation now has a signed updater metadata boundary: missing
  metadata is unavailable-safe, unsafe metadata fails, private updater fields
  fail, and present metadata must include checksum/signature/public-key and
  app/daemon compatibility fields without enabling app replacement.

## Evidence

- `0506632 feat: export Astria crash summaries`
- `5007b3a feat: add Astria notification readiness`
- `7f53d53 feat: validate Astria updater metadata boundary`
- Validation performed during Phase18 children:
  - `python3 ./.trellis/scripts/task.py validate <child-task>`
  - `scripts/smoke_macos_astria_shell.sh`
  - `scripts/validate_release_artifacts.sh --updater-boundary-smoke`
  - `scripts/validate_release_artifacts.sh --npm-only --astria-local`
  - `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
  - `go test ./...`
  - `git diff --check`

## Updated Kocoro parity estimate

Astria is now roughly 94-96% aligned with Kocoro for local-first desktop
platform behavior. The remaining gaps are mostly production execution work that
requires real distribution credentials, release infrastructure, or deliberately
scoped remote endpoints.

## Remaining Kocoro gaps

- True OS crash reporter ingestion: reading actual macOS crash/fault logs,
  classifying crash signatures, and optional user-approved export/upload.
- Real signed/notarized Astria release production with external Apple
  credentials, Hardened Runtime, notarization, stapling, and artifact
  publication.
- Verified updater implementation: signed metadata verification, app/daemon
  compatibility enforcement, replacement transaction safety, rollback, and
  user-facing updater controls.
- Broader native Desktop RPC tools beyond Calendar, especially Contacts,
  Reminders, file provider flows, and explicit permission request actions.
- Polished production notification behavior: actual user-triggered permission
  request flow, notification preferences, routing, and delivery telemetry that
  remains local or explicitly user-approved.

## Recommended next phase

Phase19 should focus on verified updater implementation and release execution:

1. signed updater verification and no-replacement dry-run flow;
2. release artifact compatibility manifest for app plus bundled daemon;
3. optional user-approved crash artifact collection from local OS crash logs.
