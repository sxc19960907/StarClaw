# Astria local OS crash artifact collection

## Goal

Add an explicit user-triggered boundary for collecting local OS crash artifacts
into redacted Astria support exports.

## Requirements

- Collect only local crash artifacts selected or explicitly approved by the
  user.
- Redact secrets, raw prompts, Desktop RPC payloads, socket/pidfile paths, and
  private local paths.
- Do not upload crash artifacts automatically.

## Acceptance Criteria

- [ ] Local OS crash artifact collection boundary is documented and smoke-tested.
- [ ] Redaction covers crash content and private paths.
