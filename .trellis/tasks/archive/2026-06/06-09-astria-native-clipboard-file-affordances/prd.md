# Astria native clipboard file affordances

## Goal

Add local, user-triggered clipboard and file affordances to Astria's native
macOS shell for support and diagnostics workflows.

## Requirements

- Add native actions for copying the current safe Astria route and copying a
  redacted diagnostics/support summary.
- Add a native action for revealing the diagnostics export directory.
- Reuse existing route safety and diagnostics redaction boundaries.
- Do not copy API keys, bearer tokens, Desktop RPC socket paths, pidfile paths,
  raw prompts, or raw Desktop RPC payloads.
- Keep existing Open/Export Diagnostics behavior.

## Acceptance Criteria

- [ ] Astria command/action model includes copy route, copy support summary, and
      reveal diagnostics folder.
- [ ] Copied route is same-origin and under `/app`, or falls back to `/app/`.
- [ ] Support summary is redacted and local-only.
- [ ] Smoke coverage validates command metadata, route safety, and redaction.

## Notes

- File picker/import workflows are out of scope for this first affordance
  slice.
