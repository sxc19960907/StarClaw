# Harden path and approval security

## Goal

Harden high-priority tool security boundaries identified in `BUG_REVIEW.md`, focusing on path containment, symlink escape prevention, and approval requirements for externally visible actions.

## Requirements

- Fix safe path containment checks so sibling-prefix paths and symlink escapes cannot bypass project boundaries.
- Ensure file-oriented tools in this scope use the established path expansion and safety checks before filesystem access.
- Require approval for `publish_to_web` so the agent cannot silently expose local files.
- Add command argument separators where missing for shell-backed tools in this scope.
- Add focused regression tests for each fixed issue.
- Keep changes narrowly scoped to the high-priority security findings.

## Acceptance Criteria

- [ ] Path checks reject sibling-prefix paths such as `<cwd>-other/...`.
- [ ] Path checks reject symlinks that resolve outside the allowed root.
- [ ] Scoped tools validate paths before reading or writing.
- [ ] `publish_to_web` requires approval.
- [ ] Focused package tests pass.
- [ ] Full `go test ./...` is attempted and results are reported.

## Notes

- Source list: `BUG_REVIEW.md` high-priority security findings, especially `tools/safe_path.go`, `tools/imaging.go`, `tools/publish_to_web.go`, `tools/screenshot.go`, and `tools/grep.go`.
