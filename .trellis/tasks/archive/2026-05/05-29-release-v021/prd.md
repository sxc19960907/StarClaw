# Release v0.2.1

## Goal

Publish the prepared `v0.2.1` release tag after final local verification.

## Requirements

- Confirm the working tree has no release-related uncommitted changes.
- Run final local release checks before tagging.
- Push `main` so the release commit exists on the remote.
- Create and push annotated tag `v0.2.1`.
- Check the GitHub release workflow status after the tag push.
- Do not publish npm manually unless the repository release process explicitly requires it and local package contents are verified.

## Acceptance Criteria

- [ ] Final tests/builds pass locally.
- [ ] `main` is pushed to `origin`.
- [ ] Annotated tag `v0.2.1` exists locally and on `origin`.
- [ ] GitHub release workflow is triggered or its status is reported.

## Notes

- Source release checklist: `RELEASE_CHECKLIST.md`.
