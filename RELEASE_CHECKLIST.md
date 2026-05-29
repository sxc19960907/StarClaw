# Release Checklist

## Current Readiness Notes

- Hardening pass is documented in `CHANGELOG.md` under `Unreleased`.
- Current product-code changes should be committed separately from the pre-existing Trellis/agent infrastructure migration.
- Latest verification on 2026-05-29 passed:
  - `go test ./...`
  - `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon ./internal/tools ./internal/heartbeat`
  - `make build`
  - `make build-all`
  - CLI smoke: `starclaw version`, `starclaw --help`, `starclaw completion zsh`, `starclaw mcp --help`
- Extended CLI smoke on 2026-05-29 passed:
  - `starclaw daemon --help`
  - `starclaw daemon start` in foreground, verified with `starclaw daemon status`, stopped with `starclaw daemon stop`
  - `starclaw schedule --help`, `starclaw schedule list`
  - `starclaw schedule create --cron '0 0 1 1 *' --prompt 'release smoke schedule 2026-05-29'`, then `starclaw schedule remove <id>`
  - `starclaw mcp serve --help`
- Final verification for this pass should include:
  - `go test ./...`
  - `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon ./internal/tools ./internal/heartbeat`
  - `make build`
  - `make build-all`
  - CLI smoke: `starclaw version`, `starclaw --help`, `starclaw completion zsh`, `starclaw mcp --help`
- As of this pass, no release tag has been created.

## Pre-Release

- [ ] Update version in `cmd/root.go`
- [ ] Update CHANGELOG.md
- [x] Run full test suite: `go test ./...`
- [x] Check cross-platform builds: `make build-all`
- [ ] Update documentation if needed

## Release Process

1. **Tag the release**
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

2. **Build with GoReleaser**
   ```bash
   goreleaser release --clean
   ```

3. **Verify artifacts**
   - Check GitHub releases page
   - Download and test binaries
   - Verify checksums

4. **Update Homebrew tap**
   - GoReleaser should auto-update
   - Verify formula is correct

5. **Publish to npm** (if applicable)
   ```bash
   cd npm
   npm publish --access public
   ```

## Post-Release

- [ ] Announce on Twitter/X
- [ ] Update GitHub release notes
- [ ] Close milestone in GitHub
- [ ] Merge release branch to main

## Rollback Plan

If critical issues found:
1. Delete GitHub release
2. Delete git tag: `git push --delete origin v0.1.0`
3. Re-tag after fix: `git tag -a v0.1.1 -m "Release v0.1.1"`
