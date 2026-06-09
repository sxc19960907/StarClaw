# Astria staged updater transaction plan design

## Boundary

The planner is a release-validation helper, not a runtime updater executor. It
may read JSON metadata and produce a deterministic decision object, but it must
not touch installed app files or bundled daemon files.

## Inputs

- Astria updater metadata fields already validated in Phase19:
  checksum, signature algorithm, public key id, app version, daemon version,
  replacement disabled state, and compatibility fields.
- New transaction safety declarations:
  rollback gate id, post-update health gate id, staging mode, and replacement
  mode.

## Output

A local JSON-compatible decision containing:

- `planReady`
- `replacementEnabled`
- `stagingMode`
- `rollbackGate`
- `postUpdateHealthGate`
- `requiredChecks`
- `blockingReasons`

## Validation

The smoke should cover:

- Valid no-replacement metadata with safety gates.
- Metadata that enables replacement.
- Metadata missing rollback or health gates.
