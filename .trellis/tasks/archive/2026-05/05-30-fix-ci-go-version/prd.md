# Fix CI Go Version Mismatch

## Goal

Restore CI by ensuring golangci-lint is built with a Go toolchain compatible with the module target.

## Confirmed Facts

- Latest CI run `26666739420` failed in the `Lint` step.
- The failure was: `the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.0)`.
- `.github/workflows/ci.yml` uses `go-version-file: go.mod`.
- The repository currently declares `go 1.25.0`.
- Running local `go mod tidy` with a newer Go toolchain keeps the module target at `go 1.25.0`; downgrading `go.mod` is not a stable fix.
- `golangci-lint-action` supports `install-mode: goinstall`, which builds golangci-lint using the configured Go toolchain.

## Requirements

- Keep the module Go target unchanged.
- Configure CI lint installation so golangci-lint is built with the Go version from `go.mod`.
- Pin the golangci-lint version instead of using floating `latest`.
- Update development documentation if it describes the wrong Go target.
- Validate with local module, vet, lint-equivalent, build, and test checks where available.

## Acceptance Criteria

- [ ] `go.mod` remains at Go 1.25.
- [ ] CI lint uses `install-mode: goinstall` with a pinned golangci-lint version.
- [ ] Documentation states the project currently targets Go 1.25.
- [ ] `go mod tidy` produces no unexpected dependency churn.
- [ ] `go vet ./...` passes.
- [ ] `go build ./...` passes.
- [ ] `go test ./...` passes.
