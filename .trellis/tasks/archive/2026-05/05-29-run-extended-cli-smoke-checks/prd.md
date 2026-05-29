# Run extended CLI smoke checks

## Goal

Run a second release-readiness smoke pass over user-facing CLI workflows beyond basic help output.

## Requirements

- Validate daemon lifecycle commands against the built local binary.
- Validate schedule command help and lightweight schedule persistence behavior where practical.
- Validate MCP serve help remains available.
- Record results in `RELEASE_CHECKLIST.md`.
- Keep any fixes narrowly scoped to failures discovered during smoke testing.
- Do not tag or publish a release in this task.

## Acceptance Criteria

- [ ] Daemon start/status/stop smoke passes or any failure is fixed/documented.
- [ ] Schedule command smoke passes or any failure is fixed/documented.
- [ ] MCP serve help smoke passes.
- [ ] Release checklist includes this extended smoke pass.

## Notes

- This task builds on the 2026-05-29 release-hardening validation pass.
