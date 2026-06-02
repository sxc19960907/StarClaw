# Add GUI global permissions editor

## Goal

Make the Web UI permissions panel editable so users can update global tool policy from the daemon GUI instead of editing `config.yaml` by hand.

## Requirements

- Backend:
  - Extend daemon config PATCH support to accept a `permissions` object.
  - Persist permissions to YAML under the existing `permissions:` config key.
  - Refresh `deps.Config` after save so daemon runs use the edited policy immediately.
  - Keep `GET /permissions` returning the current effective permission policy.
  - Preserve existing provider config patch behavior and blank-secret preservation.
- GUI:
  - Replace the read-only permissions overview with editable controls for:
    - allowed directories
    - allowed commands
    - denied commands
    - network allowlist
    - sensitive patterns
  - Show loaded/configured state and save feedback.
  - Allow clearing all permission lists back to an unconfigured/default state.
  - Keep the existing policy preview readable while editing.
- Compatibility:
  - No frontend build step.
  - Existing diagnostics, config, chat, and smoke behavior must continue to work.

## Acceptance Criteria

- [ ] `PATCH /config` can update `permissions` and write valid YAML.
- [ ] `GET /permissions` returns updated permission lists after patch.
- [ ] GUI permissions panel loads current values into editable controls.
- [ ] GUI can save edited permission rules and refresh the overview.
- [ ] GUI can clear all permission rules.
- [ ] Go tests and Web UI smoke cover the edit/save workflow.

## Notes

- This task intentionally reuses the existing `/config` write path instead of adding a second persistence mechanism.
