# Astria native permission helper flows

## Goal

Define local permission helper flows for future native desktop tools that may
need macOS privacy/TCC permissions.

## Requirements

- Surface permission status/guidance locally without requiring signed
  entitlements in development builds.
- Keep guidance actionable for Calendar, Contacts, Reminders, file access, and
  notifications where applicable.
- Do not request broad permissions silently.
- Do not add cloud auth or remote telemetry.

## Acceptance Criteria

- [ ] Astria has a local permission helper status/guidance boundary.
- [ ] Smoke/tests cover guidance text and unavailable-safe behavior.

## Notes

- Actual signed entitlement provisioning is out of scope.
