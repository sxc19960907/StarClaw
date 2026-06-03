# Audit and fix release install loop

## Goal

Ensure StarClaw's documented install and release paths match real, usable release artifacts and do not promise unavailable installation methods.

## Requirements

- Audit release workflow and GoReleaser configuration for published artifact coverage.
- Audit install documentation for claims about installer scripts, release binaries, npm, and Homebrew.
- Verify whether documented bootstrap URLs such as `https://get.starclaw.dev` are backed by real installation scripts.
- Fix documentation or repository scripts so documented installation paths are accurate.
- Avoid adding a package manager claim unless the repository already supports it.

## Acceptance Criteria

- [x] Release workflow artifact coverage is reviewed.
- [x] Installation docs only describe currently supported install paths.
- [x] Unsupported bootstrap script URLs are removed or replaced with accurate release download guidance.
- [x] NPM/Homebrew claims match repository reality.
- [x] Diff check passes.

## Audit Notes

- GoReleaser builds Linux, macOS, and Windows archives for `amd64`/`arm64`.
- GoReleaser also defines `.deb`, `.rpm`, and `.apk` Linux packages.
- No Homebrew tap is configured; docs must continue to say Homebrew is unavailable.
- `get.starclaw.dev` and `get.starclaw.dev/windows` do not resolve, so docs must not recommend those bootstrap URLs.
- npm packaging exists, but `npm/scripts/install.js` was only a placeholder. It now fails explicitly until a real npm installer is implemented.
