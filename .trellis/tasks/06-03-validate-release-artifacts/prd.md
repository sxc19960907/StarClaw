# Validate release artifacts

## Goal

Verify the release artifact shape documented for users matches what GoReleaser actually produces.

## Requirements

- Run a local snapshot or dry-run release build when tooling is available.
- Verify archive names for macOS, Linux, and Windows match install docs.
- Verify release archives contain the `starclaw` or `starclaw.exe` executable.
- Verify Linux package formats are configured or produced as documented.
- If local GoReleaser execution is unavailable, document the blocker and add a script/check that can run where GoReleaser exists.

## Acceptance Criteria

- [x] Release artifact naming is checked against docs.
- [x] Archive executable contents are checked.
- [x] Linux package coverage is checked.
- [x] A reusable validation script or documented command exists.
- [x] Diff check passes.

## Validation Notes

- Added `scripts/validate_release_artifacts.sh`.
- The script checks documented macOS/Linux/Windows archive names, executable presence inside archives, and `.deb`/`.rpm`/`.apk` package artifacts.
- The script supports `--snapshot` for environments with GoReleaser installed.
- Local `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean --skip=publish` was blocked by a Go proxy network timeout, so the script was self-tested with synthetic release artifacts.
