# Design

## Scope

This is a release process/documentation hardening task. It should not publish to npm or require npm credentials.

## Changes

- Add `docs/RELEASE.md` as the release operator checklist.
- Extend `scripts/validate_release_artifacts.sh` to run `npm pack --dry-run --json` and verify expected npm package files.
- Add release workflow steps before GoReleaser to run local smoke commands that do not require secrets.

## Compatibility

Existing release workflow remains tag-triggered. npm publish remains manual until credentials and policy are decided.
