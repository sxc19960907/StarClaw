# Implementation Plan

## Checklist

1. Add missing `Health`, `Status API`, and `Config` lines to `app --check`.
2. Confirm `/version` and `/diagnostics` contain matching runtime fields; add missing fields only.
3. Adjust GUI Version page rows only if needed after API review.
4. Extend CLI/unit tests for expanded readiness output.
5. Extend app launch and release install smoke assertions.
6. Update README launch section with `doctor` as the status/support command.
7. Run validation commands.

## Validation Commands

- `scripts/smoke_app_launch.sh`
- `scripts/smoke_release_install.sh`
- `scripts/smoke_webui_core.sh`
- `go test ./...`
- `go vet ./...`
- `git diff --check`

## Risk Points

- Smoke scripts compare exact spacing on CLI labels.
- Runtime URLs are fixed to the daemon port today; keep formatting consistent with existing constants.
- Do not expose API keys or provider secrets in readiness output.
