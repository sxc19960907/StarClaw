# Implementation Plan

## Checklist

1. Replace `npm/scripts/install.js` with platform detection, download/local archive, extraction, and binary copy logic.
2. Replace `npm/bin/starclaw` stub with a spawn wrapper.
3. Update `npm/scripts/uninstall.js` to remove binary artifacts while preserving the shim.
4. Add `scripts/smoke_npm_install.sh`:
   - build local release archive for host platform
   - `npm pack` the npm package
   - install into a temp project with `STARCLAW_NPM_ARCHIVE`
   - run `npx starclaw version`
   - run `npx starclaw app --check`
5. Update README and docs/INSTALL npm sections.
6. Run validation commands.

## Validation Commands

- `scripts/smoke_npm_install.sh`
- `go test ./...`
- `go vet ./...`
- `git diff --check`

## Risk Points

- npm lifecycle scripts run from the package directory; use `__dirname` relative paths.
- Windows extraction requires zip handling through PowerShell or built-in tooling; smoke may only run current host platform.
- Do not delete the shim during uninstall; only remove installed binary artifacts.
- Avoid adding npm dependencies; installer should use Node built-ins.
