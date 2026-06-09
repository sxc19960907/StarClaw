# Astria local crash reporter summaries

## Goal

Add local crash summary and export affordances to Astria so support workflows can
inspect recent local failure information without automatic upload.

## Requirements

- Generate a local crash summary/report from known local crash/failure inputs.
- Reuse diagnostics redaction for API keys, bearer tokens, Desktop RPC socket
  paths, pidfile paths, raw prompts, and private local paths.
- Keep crash summary export local-only and user-triggered.
- Do not add remote crash ingestion, telemetry, cloud auth, or automatic upload.

## Acceptance Criteria

- [ ] Astria has a local crash summary model/export boundary.
- [ ] Crash summary output is redacted and local-only.
- [ ] Smoke coverage validates summary content and redaction.

## Notes

- Full OS crash reporter integration can expand later; this slice establishes
  the local summary/export contract first.
