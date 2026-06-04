# Update GitHub Actions runtime

## Goal

Remove the GitHub Actions Node.js 20 deprecation warning by upgrading workflow actions to current major versions.

## Requirements

- Update `actions/checkout` from `v4` to `v6.0.3` in CI and release workflows.
- Update `actions/setup-go` from `v5` to `v6.4.0` in CI and release workflows.
- Keep `golangci/golangci-lint-action` on `v6` for this release because v7+ does not support the existing `golangci-lint` v1 pin.
- Preserve the pinned `golangci-lint` binary at `v1.64.8`; v2 enables stricter checks that require a separate lint cleanup task.
- Preserve existing workflow behavior and step ordering.
- Push the change and confirm CI passes.

## Acceptance Criteria

- [x] CI workflow uses updated checkout/setup-go action versions while preserving compatible lint behavior.
- [x] Release workflow uses the updated action versions.
- [x] GitHub Actions CI passes after push.
- [x] Node.js 20 deprecation warning is no longer present for updated checkout/setup-go actions; the remaining lint action compatibility constraint is recorded below.

## Notes

- Latest versions checked through GitHub API on 2026-06-04:
  - `actions/checkout`: `v6.0.3`
  - `actions/setup-go`: `v6.4.0`
  - `golangci/golangci-lint-action`: `v9.2.1`
- CI attempts showed `golangci/golangci-lint-action@v9.2.1` requires golangci-lint v2; v2 currently reports existing lint debt. Keep the v6 action for this release and handle v2 migration separately.
- Final CI: `26933356798` passed on commit `4f0b10c849e4d31ab420053890ad5c47a1e5d0f4`.
- Follow-up: migrate to `golangci/golangci-lint-action` v7+ only after resolving the current golangci-lint v2 findings.
