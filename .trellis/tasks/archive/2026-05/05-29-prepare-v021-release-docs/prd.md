# Prepare v0.2.1 release docs

## Goal

Finish the first three release follow-up tasks: version documentation, changelog/README calibration, release checklist closeout, and explicit self-update limitation handling.

## Requirements

- Prepare docs and package metadata for `v0.2.1` as the next patch release after `v0.2.0`.
- Keep Go runtime version injection behavior unchanged: release builds get the version through `ldflags`/GoReleaser and source builds remain `dev`.
- Update `CHANGELOG.md` so the hardening pass is represented as a concrete `v0.2.1` release section.
- Fix README inconsistencies around built-in tool count and self-update behavior.
- Make update command/user-facing copy explicit that automatic binary replacement is not implemented yet.
- Update `RELEASE_CHECKLIST.md` for the completed documentation/version work and the remaining tag/publish steps.

## Acceptance Criteria

- [ ] Version docs and npm package metadata target `v0.2.1`.
- [ ] README tool count matches the registered local tools plus the separately registered version/session/MCP tools described in the table.
- [ ] README and CLI update copy do not imply automatic install is complete.
- [ ] CHANGELOG has a dated `v0.2.1` section for the hardening pass.
- [ ] Release checklist reflects completed docs/version prep and remaining release actions.
- [ ] Tests/build sanity checks pass after the documentation/code copy changes.

## Notes

- Do not create a release tag or publish artifacts in this task.
