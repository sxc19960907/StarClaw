# Improve GUI permissions preview

## Goal

Make global and agent-level permission editing safer by showing what will be saved and surfacing obvious risky configurations before the user saves.

## Requirements

- The Permissions page should show a live pending preview based on the current form values, not only the last loaded config.
- The preview should summarize counts for allowed dirs, allowed commands, denied commands, network allowlist, and sensitive patterns.
- The preview should show risk hints for broad local access, missing deny rules, missing sensitive patterns, and broad network allowlist entries.
- Clearing rules should refresh the pending preview immediately.
- Agent permission preview should flag auto-approve and conflicting allow/deny entries.
- Browser permissions/agents smoke should cover the new preview text.

## Acceptance Criteria

- [x] Permissions form edits update a pending preview before save.
- [x] Preview shows broad access and empty-deny warnings.
- [x] Clear rules updates the preview before/after save.
- [x] Agent permission preview warns when auto-approve is enabled.
- [x] Agent permission preview warns when the same tool is in allow and deny.
- [x] Targeted JS and Web UI smoke checks pass.
