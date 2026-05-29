# Implementation

## Steps

- [ ] Run final `go test ./...`.
- [ ] Run final `make build-all`.
- [ ] Push `main` to `origin`.
- [ ] Create annotated `v0.2.1` tag.
- [ ] Push `v0.2.1` tag to `origin`.
- [ ] Check GitHub Actions release workflow.
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
