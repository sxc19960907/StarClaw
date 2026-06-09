# Packaging signing update boundary design

## Decision

Define Astria as a macOS-only local-first desktop artifact boundary for this
phase:

1. Keep the existing unsigned development `.app` build as the always-available
   local artifact.
2. Add deterministic unsigned packaging validation for CI/local smoke, without
   requiring Apple credentials.
3. Document the release artifact shape for future signed builds:
   `Astria.app` plus a bundled `starclaw` daemon binary under app resources.
4. Define signing, notarization, and updater boundaries as opt-in release
   responsibilities that never require committed private credentials.
5. Preserve existing CLI, npm, archive, and GoReleaser release checks.

This mirrors the useful Kocoro boundary in the public repo: daemon/archive/npm
release automation is public and checksum-backed, while Desktop signing,
notarization, and app-store-grade packaging remain a separate credentialed
release concern.

## Scope

In scope:

- Document development build and smoke commands.
- Add or tighten local packaging smoke checks for `Astria.app` bundle metadata,
  embedded daemon compatibility, route recovery, and daemon supervision.
- Define where a bundled daemon binary may live in the app.
- Define app/daemon version compatibility rules.
- Define signing/notarization requirements and what can be checked without
  private Apple credentials.
- Define updater behavior and failure modes at the boundary level.
- Update install/release docs and code-spec.

Out of scope:

- Committing signing identities, Apple ID credentials, notary profiles, or
  update private keys.
- Shipping a signed/notarized DMG in this repository's Linux release workflow.
- Adding remote telemetry, cloud sync, or off-machine update checks by default.
- Implementing Sparkle or another auto-updater before update metadata and
  signing policy are settled.
- Deep Kocoro-style Desktop RPC pidfile/socket reconciliation.

## Artifact Contract

Development artifact:

- Path: `build/desktop/macos/Astria.app`
- Builder: `scripts/build_macos_astria_shell.sh`
- Signing: unsigned
- Daemon source: external `starclaw` selected by:
  - `ASTRIA_STARCLAW_BIN`
  - bundled resource path when present
  - `PATH` fallback

Future release artifact:

- `Astria.app`
- `Contents/MacOS/Astria`
- `Contents/Resources/starclaw`
- `Contents/Info.plist`
- Optional archive/DMG wrapping the `.app`

The bundled daemon must be executable and should report the same StarClaw
semantic version as the app release. Development builds may use `dev` or
`0.0.0`, but release packaging should not mix an app version and daemon version
that point to different release tags.

## Signing and Notarization Boundary

Local and CI checks may validate:

- App bundle structure.
- `CFBundleIdentifier`.
- `CFBundleShortVersionString` / `CFBundleVersion` presence.
- local networking ATS allowance.
- executable bits for app binary and bundled daemon.
- unsigned smoke execution.

Credentialed release checks belong outside the default CI job and require:

- Apple Developer Team ID.
- Developer ID Application signing identity.
- Hardened Runtime.
- `notarytool` keychain profile or equivalent secure secret store.
- stapling validation for distributed artifacts.

No signing identity, notarization password, API key, app-specific password, or
update signing private key may be committed.

## Updater Boundary

For this phase, the desktop app does not auto-update itself. User-facing update
behavior remains CLI-driven through existing StarClaw release/update paths.

Documented future updater rules:

- Update metadata must be unavailable-safe: missing metadata leaves the current
  app running and surfaces a non-fatal status.
- Updates must verify checksums/signatures before replacing an app or daemon.
- App and bundled daemon versions must stay compatible; incompatible metadata
  should be rejected rather than launching a mismatched daemon.
- No background update check should send off-machine telemetry by default in
  local-first mode.

## Compatibility

- `starclaw app` browser launch remains supported.
- Existing GoReleaser archives, Linux packages, and npm package validation
  remain unchanged.
- `scripts/smoke_macos_astria_shell.sh` remains the local macOS authority for
  unsigned desktop shell validation.
- Release workflow remains Linux-based unless a separate macOS signing job is
  explicitly introduced with secrets.

## Kocoro Parity

Kocoro's public repo does not contain signed Desktop packaging, DMG creation,
or Sparkle metadata. Its concrete public packaging advantages are:

- release archives include sidecars;
- npm installer downloads matching GitHub release assets;
- update checks use GitHub releases plus `checksums.txt`;
- Desktop/daemon version and capability boundaries are explicit.

This task closes StarClaw's comparable gap by making Astria's app artifact,
bundled daemon boundary, unsigned smoke validation, and signing/update limits
explicit and testable. It does not claim parity with any closed-source Kocoro
Desktop signing pipeline.
