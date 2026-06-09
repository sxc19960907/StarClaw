# Astria native notification readiness

## Goal

Add notification readiness and permission guidance for Astria without surprise
permission prompts.

## Requirements

- Surface notification authorization/readiness locally.
- Keep permission requests explicit and user-triggered.
- Add smoke coverage for status/guidance text and unavailable-safe behavior.
- Do not add remote notification services or cloud auth.

## Acceptance Criteria

- [ ] Astria has a local notification readiness boundary.
- [ ] Passive readiness checks do not request notification permission.
- [ ] Smoke/tests cover guidance and unavailable-safe behavior.
