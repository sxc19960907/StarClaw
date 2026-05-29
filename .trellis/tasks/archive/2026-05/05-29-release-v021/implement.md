# Implementation

## Steps

- [x] Run final `go test ./...`.
- [x] Run final `make build-all`.
- [x] Push `main` to `origin`.
- [x] Create annotated `v0.2.1` tag.
- [x] Push `v0.2.1` tag to `origin`.
- [x] Check GitHub Actions release workflow.
- [x] Fix Windows release build failure.
- [x] Re-run `go test ./...`.
- [x] Re-run `make build-all` with Windows targets.
- [x] Move and push `v0.2.1` tag to the fixed commit.
- [x] Confirm GitHub release assets were published.
- [x] Disable Homebrew publishing until a real tap repository exists.
- [x] Update installation docs so Homebrew is not advertised as available.
- [x] Re-run `go test ./...`.
- [x] Re-run `make build-all`.
- [x] Record release status.

## Release Status

- `v0.2.1` is tagged and published at `https://github.com/sxc19960907/StarClaw/releases/tag/v0.2.1`.
- GitHub Release assets were uploaded successfully for macOS, Linux, Windows, Linux packages, and checksums.
- The GitHub Actions release run failed after asset upload while attempting to publish a Homebrew formula to the missing `starclaw/homebrew-tap` repository.
- `.goreleaser.yaml` no longer publishes Homebrew formulas until a real tap repository exists.
- `README.md`, `docs/INSTALL.md`, and `RELEASE_CHECKLIST.md` now describe Homebrew as unavailable.

## Validation

```bash
go test ./...
make build-all
git push origin main
git tag -a v0.2.1 -m "Release v0.2.1"
git push origin v0.2.1
gh run list --workflow release.yml --limit 3
```
