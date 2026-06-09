# Astria sandbox updater rehearsal fixture

## Goal

Add a sandbox-only updater rehearsal smoke that stages a disposable candidate
app fixture, rehearses replacement inside a temporary install directory, rolls
back to the previous fixture, and proves no real Astria paths are touched.

## Requirements

- Create all rehearsal inputs under `mktemp`.
- Rehearse only against fake `Astria.app` fixture directories.
- Record touched paths and require every touched path to stay under the sandbox.
- Rehearse candidate staging, sandbox replacement, and rollback.
- Do not touch `/Applications`, real Astria app bundles, real Application
  Support, network, or Apple credentials.

## Acceptance Criteria

- [ ] Valid sandbox rehearsal reports staged, replaced, and rolled back states.
- [ ] Guard rejects any touched path outside the sandbox.
- [ ] Release validation includes a dedicated sandbox updater rehearsal smoke.
- [ ] `--npm-only --astria-local` runs the sandbox rehearsal.
