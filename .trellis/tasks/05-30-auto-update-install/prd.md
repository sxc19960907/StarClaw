# Implement Automatic Update Installation

## Goal

Make `starclaw update` install a newer GitHub Release binary for the current platform instead of only reporting that automatic installation is not implemented.

## User Value

Users can upgrade an installed StarClaw binary with one command, while `starclaw update --check` remains a safe read-only way to inspect available releases.

## Confirmed Facts

- `cmd/root.go` already exposes `starclaw update` and `starclaw update --check`.
- `internal/update/selfupdate.go` already checks GitHub releases and can download an asset.
- `DoUpdate` currently finds a matching asset and then returns `automatic update installation is not implemented yet`.
- GoReleaser publishes platform archives named like `starclaw_Darwin_arm64.tar.gz`, `starclaw_Linux_x86_64.tar.gz`, and `starclaw_Windows_x86_64.zip`.
- GoReleaser also uploads `checksums.txt` with release asset checksums.
- `v0.2.1` release assets exist for darwin/linux/windows on amd64 and arm64.

## Requirements

- `starclaw update --check` must keep its current check-only behavior and must not download or replace files.
- `starclaw update` must:
  - reject non-semver development builds as it does today;
  - detect when the current version is already latest;
  - choose the asset that exactly matches the current `GOOS/GOARCH`;
  - download the release archive to a temporary working directory;
  - download and verify `checksums.txt` before installing;
  - extract the `starclaw` binary from `.tar.gz` archives and `starclaw.exe` from `.zip` archives;
  - replace the currently running executable as safely as the platform allows;
  - preserve executable permissions on Unix platforms;
  - leave the old binary in place if download, checksum, extraction, or replacement fails.
- Error messages must be actionable and include the missing platform or failed phase.
- Unit tests must cover asset matching, checksum verification, archive extraction, and replacement failure behavior.

## Out of Scope

- Creating or publishing a Homebrew tap.
- npm package update behavior.
- Background auto-install on startup through `update.auto_install`.
- Code signing, notarization, or OS package manager integration.
- Delta updates or partial binary patches.

## Acceptance Criteria

- [ ] `starclaw update` can install from a mocked GitHub release server in tests.
- [ ] `starclaw update --check` remains check-only.
- [ ] Checksum mismatch prevents installation and keeps the existing binary unchanged.
- [ ] Missing platform asset returns a clear error.
- [ ] Archive extraction supports `.tar.gz` and `.zip`.
- [ ] Unix replacement path sets executable mode.
- [ ] Existing update tests are updated so the previous "not implemented" behavior is gone.
- [ ] `go test ./internal/update ./cmd` passes.
- [ ] `go test ./...` passes.
