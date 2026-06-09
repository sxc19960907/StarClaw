# Astria native diagnostics export and crash reports

## Goal

Add local-only diagnostics export and crash/failure report boundaries for the
macOS Astria shell.

## Requirements

- Export local diagnostics without off-machine telemetry.
- Include daemon health/status, Desktop RPC session state, app version, and
  recent failure summaries where available.
- Redact API keys, provider headers, raw Desktop RPC payloads, user prompts,
  socket paths, and pidfile paths.
- Preserve existing diagnostics URL behavior.

## Acceptance Criteria

- [ ] Astria has a local diagnostics export boundary.
- [ ] Crash/failure report content is redacted and locally generated.
- [ ] Smoke/tests cover redaction and export metadata.

## Notes

- Full crash reporter upload is out of scope.
