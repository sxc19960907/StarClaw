# Implementation Plan

## Checklist

- [x] Read relevant backend quality/error-handling specs before editing.
- [x] Inspect CLI command behavior and existing test helpers.
- [x] Build the binary locally.
- [x] Run manual smoke commands with isolated environment and record failures.
- [x] Add a repeatable smoke validation script if missing.
- [x] Fix narrow blocking issues found by smoke validation.
- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Run the new smoke validation script.
- [ ] Push changes and verify GitHub Actions CI.

## Validation Commands

```bash
go build ./...
go test ./...
go vet ./...
```

If added:

```bash
./scripts/smoke_cli.sh
```

## Risk Points

- CLI setup may enter an interactive prompt if config is missing. Smoke checks must avoid hanging.
- Config discovery may read the real user home unless `HOME` is isolated.
- Desktop tools and real browser commands are not stable in CI or headless environments.
