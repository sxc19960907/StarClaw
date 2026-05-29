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
- [ ] Move and push `v0.2.1` tag to the fixed commit.
- [ ] Confirm GitHub Actions release workflow succeeds.
- [ ] Record release status.

## Validation

```bash
go test ./...
make build-all
git push origin main
git tag -a v0.2.1 -m "Release v0.2.1"
git push origin v0.2.1
gh run list --workflow release.yml --limit 3
```
