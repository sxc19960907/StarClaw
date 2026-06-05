# Design

## Architecture

The npm package remains a thin installer/wrapper:

- `npm/scripts/install.js` downloads or reads a release archive, extracts `starclaw` / `starclaw.exe`, and writes it to `npm/bin/`.
- `npm/bin/starclaw` is a Node.js shim that spawns the installed binary.
- `npm/scripts/uninstall.js` removes installed binary artifacts.
- `scripts/smoke_npm_install.sh` builds a local release archive and tests npm installation using `STARCLAW_NPM_ARCHIVE`.

No Go runtime code is changed for npm installation.

## Platform Mapping

Node platform/arch maps to GoReleaser asset names:

- `darwin` + `arm64` -> `starclaw_Darwin_arm64.tar.gz`
- `darwin` + `x64` -> `starclaw_Darwin_x86_64.tar.gz`
- `linux` + `arm64` -> `starclaw_Linux_arm64.tar.gz`
- `linux` + `x64` -> `starclaw_Linux_x86_64.tar.gz`
- `win32` + `arm64` -> `starclaw_Windows_arm64.zip`
- `win32` + `x64` -> `starclaw_Windows_x86_64.zip`

## Inputs

- `STARCLAW_NPM_ARCHIVE`: local archive path for tests or offline installs.
- `STARCLAW_NPM_VERSION`: optional version/tag override; defaults to `latest`.
- `STARCLAW_NPM_BASE_URL`: optional URL base override; defaults to GitHub releases.

## Install Flow

1. Resolve platform archive name.
2. If `STARCLAW_NPM_ARCHIVE` is set, copy/extract that archive.
3. Otherwise download from GitHub Release URL.
4. Extract archive to a temp directory.
5. Locate `starclaw` or `starclaw.exe`.
6. Copy it into `npm/bin/starclaw-bin` or `npm/bin/starclaw.exe`.
7. Set executable mode on non-Windows.

## Compatibility

- `package.json` keeps the public bin name as `starclaw`.
- Existing npm installs that only saw the placeholder now get a working shim.
- Local smoke does not require external network.

## Rollback

Revert npm scripts/shim/docs/smoke changes. No user data or StarClaw config is touched.
