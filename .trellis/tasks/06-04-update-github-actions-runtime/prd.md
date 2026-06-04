# Update GitHub Actions runtime

## Goal

Remove the GitHub Actions Node.js 20 deprecation warning by upgrading workflow actions to current major versions.

## Requirements

- Update `actions/checkout` from `v4` to `v6.0.3` in CI and release workflows.
- Update `actions/setup-go` from `v5` to `v6.4.0` in CI and release workflows.
- Update `golangci/golangci-lint-action` from `v6` to `v9.2.1` in CI.
- Preserve existing workflow behavior and step ordering.
- Push the change and confirm CI passes.

## Acceptance Criteria

- [x] CI workflow uses the updated action versions.
- [x] Release workflow uses the updated action versions.
- [ ] GitHub Actions CI passes after push.
- [ ] Node.js 20 deprecation warning is no longer present, or any remaining warning is recorded.

## Notes

- Latest versions checked through GitHub API on 2026-06-04:
  - `actions/checkout`: `v6.0.3`
  - `actions/setup-go`: `v6.4.0`
  - `golangci/golangci-lint-action`: `v9.2.1`
