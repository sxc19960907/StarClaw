# Design

## Permissions Page

- Add a `permissions-pending-preview` container above the loaded policy list.
- Build preview from `buildPermissionsPayload()` so it reflects exactly what will be submitted.
- Risk hints are client-side only and informational.

## Agent Preview

- Extend `renderAgentPermissionPreview()` to add warning rows:
  - auto approve enabled;
  - intersection between allow and deny lists.

## Smoke

- Extend permissions smoke to assert pending preview updates before saving and after clearing.
- Extend agents smoke to assert auto-approve warning and allow/deny conflict warning.
