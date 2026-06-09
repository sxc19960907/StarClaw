# Astria Kocoro parity phase 18 design

## Architecture Boundary

Phase18 keeps Astria as a local macOS shell around the daemon Web UI. Production
native reliability features should produce local, redacted, inspectable
artifacts and unavailable-safe status rather than remote upload or credentialed
release execution.

## Native Areas

- Crash summaries: discover recent local crash/fault artifacts where available,
  redact sensitive values, and export a support-friendly local summary.
- Notifications: surface notification permission/readiness and support a
  smoke-testable notification contract without surprise prompts.
- Release/updater boundary: document and validate checksum/signature metadata
  requirements while rejecting unsafe updater metadata and private credentials.

## Compatibility

- Existing CLI/browser launch remains unchanged.
- Existing Desktop RPC and daemon lifecycle remain unchanged.
- Unsigned local builds remain supported and must not require Apple credentials.

## Rollout

1. Local crash reporter summaries.
2. Notification readiness.
3. Signed updater/release boundary.
4. Final Kocoro gap review.
