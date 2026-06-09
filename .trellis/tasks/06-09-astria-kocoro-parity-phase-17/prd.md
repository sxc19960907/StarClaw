# Astria Kocoro parity phase 17: native OS tool depth

## Goal

Close the next Kocoro parity gap after Phase16 by adding deeper local macOS
native affordances around the Astria shell: clipboard/file actions,
permission-helper flows, and richer multi-window state restoration.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at
`74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Confirmed Facts

- Phase16 completed native app commands, local diagnostics export, and
  credential-free distribution boundary validation.
- Phase16 final gap review estimates Astria is roughly 90-93% aligned for
  local-first desktop lifecycle and baseline native shell behavior.
- Remaining gaps are deeper OS affordances and native lifecycle polish.

## Child Plan

1. `astria-native-clipboard-file-affordances`: add native clipboard and file
   reveal/copy affordances for local diagnostics and support workflows.
2. `astria-native-permission-helper-flows`: add local permission status/help
   boundaries for future TCC-backed desktop tools.
3. `astria-multi-window-state-restoration`: improve multi-window route/state
   restoration and lifecycle polish.

## Requirements

- Keep the daemon-served Web UI as the primary experience.
- Keep native affordances local-only and user-triggered.
- Do not add off-machine telemetry, cloud auth, or automatic upload.
- Do not expose secrets, raw prompts, Desktop RPC payloads, socket paths, or
  pidfile paths.
- Add smoke/test coverage for native affordance contracts.

## Acceptance Criteria

- [ ] Each child task has independent planning artifacts and testable
      acceptance criteria before implementation.
- [ ] Astria exposes native clipboard/file affordances for local support
      workflows.
- [ ] Astria defines permission helper flow boundaries for native desktop
      tools.
- [ ] Astria improves multi-window state restoration without breaking route
      safety.
- [ ] Final gap review updates Kocoro parity and remaining native OS gaps.

## Out of Scope

- Remote diagnostics upload.
- Full native rewrite of the Web UI.
- Actual Apple TCC entitlement provisioning or signed public release.

## Notes

Parent task only. Start child tasks for implementation.
