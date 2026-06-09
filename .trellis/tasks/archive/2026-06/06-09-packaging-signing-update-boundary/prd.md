# Packaging signing update boundary

## Goal

Define the standalone desktop app packaging, signing, notarization, update, and
release smoke-test boundary without introducing private credentials or cloud
sync by default.

## Requirements

- Document local development build steps for the desktop shell.
- Define release artifact shape for the standalone app and how it relates to
  existing CLI, npm, and archive artifacts.
- Define signing/notarization requirements and what can be validated in CI
  without private credentials.
- Define updater behavior and safety boundaries, including how the app and
  bundled daemon versions stay compatible.
- Preserve existing release checks for CLI/npm/binary artifacts.
- Do not commit signing credentials, notarization secrets, or update private
  keys.

## Acceptance Criteria

- [ ] Documentation covers local app build, release artifact shape,
      signing/notarization boundary, and updater/version compatibility.
- [ ] Smoke checks can validate unsigned/local builds.
- [ ] Existing `docs/RELEASE.md` and `docs/INSTALL.md` are updated if the app
      artifact becomes user-facing.
- [ ] Failure modes for version mismatch, unsigned dev builds, and unavailable
      update metadata are documented.
- [ ] No private credentials or remote telemetry defaults are added.

## Notes

Should run after the shell and launcher have a concrete artifact to package.
