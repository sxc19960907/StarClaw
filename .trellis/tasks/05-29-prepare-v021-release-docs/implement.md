# Implementation

## Steps

- [x] Set package metadata release target to `v0.2.1`.
- [x] Convert `CHANGELOG.md` Unreleased hardening notes into a `v0.2.1` section.
- [x] Fix README built-in tool count and update behavior text.
- [x] Adjust update command/user-facing copy to say installation is not implemented yet.
- [x] Update release checklist remaining/completed items.
- [x] Run focused update/cmd tests.
- [x] Run full test suite and build sanity checks.
- [x] Run whitespace check.

## Validation

```bash
go test ./cmd ./internal/update
go test ./...
make build
make build-all
git diff --check
```
