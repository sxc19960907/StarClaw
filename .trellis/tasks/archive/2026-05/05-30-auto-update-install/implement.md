# Implementation

## Steps

- [x] Load project specs and current update tests.
- [x] Add exact platform asset naming and checksum asset lookup.
- [x] Add checksum download/parsing/verification helpers.
- [x] Add archive extraction helpers for `.tar.gz` and `.zip`.
- [x] Add executable replacement helper with backup/restore behavior.
- [x] Wire `DoUpdate` to install the selected asset.
- [x] Update CLI text to remove the "not implemented" message.
- [x] Update README/config docs where `auto_install` is described as future-only.
- [x] Replace the old not-implemented test with install success/failure tests.
- [x] Run focused update and command tests.
- [x] Run full test suite.

## Results

- `starclaw update` now downloads the matching release archive, verifies `checksums.txt`, extracts the binary, and replaces the current executable.
- `starclaw update --check` remains read-only.
- Release lookup now points to `sxc19960907/StarClaw`, matching the published `v0.2.1` GitHub Release.
- Tests cover exact asset naming, checksum mismatch, tar.gz and zip extraction, replacement rollback, and mocked install success.

## Validation Commands

```bash
go test ./internal/update ./cmd
go test ./...
```

## Risk Points

- Replacing a running executable is platform-sensitive, especially on Windows.
- Incorrect asset matching can install the wrong architecture.
- Checksum parsing must match GoReleaser checksum file format.
- Tests should avoid modifying the real running test binary.
