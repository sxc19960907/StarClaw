# Implement npm Installer

## Problem

The repository has npm packaging scaffolding, but installing it fails intentionally. This blocks npm from being a useful cross-platform install path even though README and install docs mention npm as a future distribution option.

## User Value

Users should be able to install the StarClaw CLI from an npm package that downloads the correct GitHub Release binary for their platform and exposes a working `starclaw` command.

## Confirmed Facts

- `npm/package.json` has `postinstall` and `preuninstall` scripts.
- `npm/scripts/install.js` currently exits with "npm distribution is not published yet."
- `npm/bin/starclaw` currently exits with "Please build from source."
- GoReleaser publishes release archive names:
  - `starclaw_Darwin_arm64.tar.gz`
  - `starclaw_Darwin_x86_64.tar.gz`
  - `starclaw_Linux_arm64.tar.gz`
  - `starclaw_Linux_x86_64.tar.gz`
  - Windows zip archives via `format_overrides`
- Existing release install smoke can validate an extracted StarClaw binary.

## Requirements

- Replace the npm placeholder installer with a Node.js installer that downloads and extracts the platform release archive.
- Support macOS, Linux, and Windows on x64/arm64 where release assets exist.
- Write the installed executable under the npm package `bin/` path used by `package.json`.
- Make `npm/bin/starclaw` execute the installed binary and fail with a clear message when the binary is missing.
- Provide deterministic local smoke coverage using a locally built release archive, without requiring GitHub network access.
- Update README and install docs to present npm as a supported path with the release-backed installer.

## Out of Scope

- Publishing the package to npm.
- Changing release asset naming.
- Adding Homebrew.
- Verifying npm registry authentication or provenance.

## Acceptance Criteria

- [x] `npm/scripts/install.js` installs from `STARCLAW_NPM_ARCHIVE` when provided.
- [x] `npm/scripts/install.js` can derive the correct GitHub Release URL for the current platform when no local archive is provided.
- [x] `npm/bin/starclaw` runs the installed binary.
- [x] Unsupported platforms fail with an actionable message.
- [x] A new npm smoke script builds a local release archive, installs the npm package into a temp project, and verifies `starclaw version` and `starclaw app --check`.
- [x] README and docs/INSTALL describe npm as release-backed and supported.
- [x] `scripts/smoke_npm_install.sh`, `go test ./...`, `go vet ./...`, and `git diff --check` pass locally.
