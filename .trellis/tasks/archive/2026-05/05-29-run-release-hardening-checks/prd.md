# Run release hardening checks

## Goal

Run the first release-hardening verification loop and fix any blocking issues found by the release checklist smoke tests.

## Requirements

- Execute the release checklist's core validation commands:
  - `go test ./...`
  - `go test -race ./internal/client ./internal/agent ./internal/context ./internal/daemon ./internal/tools ./internal/heartbeat`
  - `make build`
  - `make build-all`
  - CLI smoke checks for version/help/completion/MCP help.
- Keep fixes narrowly scoped to failures discovered during validation.
- Do not tag or publish a release in this task.
- Leave unrelated untracked files untouched.

## Acceptance Criteria

- [ ] Full test suite passes.
- [ ] Targeted race test suite passes.
- [ ] Local and cross-platform builds pass.
- [ ] CLI smoke checks pass.
- [ ] Any release-blocking failures discovered in this loop are either fixed or documented with a concrete next task.

## Notes

- Source checklist: `RELEASE_CHECKLIST.md`.
