# Migrate golangci-lint v2

## Goal

Upgrade CI linting to a Node 24-compatible `golangci/golangci-lint-action` release by migrating the project from `golangci-lint` v1.64.8 to v2.

## Requirements

- Upgrade `.github/workflows/ci.yml` from `golangci/golangci-lint-action@v6` to a v7+ release that supports Node 24.
- Upgrade the pinned `golangci-lint` binary to a compatible v2 release.
- Fix v2 lint findings in project code without weakening meaningful checks solely to pass CI.
- Keep code changes scoped to lint correctness unless a finding exposes a real bug.
- Preserve existing CI step ordering and smoke coverage.

## Acceptance Criteria

- [x] `golangci-lint` v2 passes in GitHub Actions.
- [x] `go test ./...` passes locally.
- [x] `go vet ./...` passes locally.
- [x] CI workflow uses a Node 24-compatible lint action.
- [x] GitHub Actions CI passes after push.
- [x] Node.js 20 action deprecation warning is gone from CI.

## Notes

- The previous runtime cleanup intentionally kept `golangci/golangci-lint-action@v6` because v7+ requires `golangci-lint` v2.
- Initial migration attempts showed existing lint debt under v2; this task owns resolving that debt.
- Local `golangci-lint` v2 execution was blocked by Go proxy timeouts in this environment; GitHub Actions run `26943210417` verified `golangci-lint v2.12.2` successfully.
- Final CI: `26943210417` passed on commit `e77c0453cf15638d42bb11db17e5c700dfbb71f2`.
