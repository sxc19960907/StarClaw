# Release Checklist

## Current Readiness Notes

- Target release: `v0.2.1`.
- Hardening pass is documented in `CHANGELOG.md` under `v0.2.1`.
- Go binary versions are injected by `ldflags`/GoReleaser; source builds keep `cmd.Version`/`main.Version` at `dev`.
- npm package metadata has been updated to `0.2.1`.
- Update checks are supported, but automatic binary replacement is not implemented yet and is documented as a follow-up.
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

- [x] Confirm Go version injection strategy in `cmd/root.go`/`main.go`
- [x] Update npm package version
- [x] Update CHANGELOG.md
- [x] Run full test suite: `go test ./...`
- [x] Check cross-platform builds: `make build-all`
- [x] Update documentation if needed

## Release Process

1. **Tag the release**
   ```bash
   git tag -a v0.2.1 -m "Release v0.2.1"
   git push origin v0.2.1
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
2. Delete git tag: `git push --delete origin v0.2.1`
3. Re-tag after fix: `git tag -a v0.2.2 -m "Release v0.2.2"`
