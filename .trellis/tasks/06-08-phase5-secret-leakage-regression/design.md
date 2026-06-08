# Phase 5 secret leakage regression design

## Scope

Add regression coverage for surfaces that are supposed to be aggregate-safe or redacted after Phase 4 platform work: metrics, structured events, trace read/export, run summaries, recovery metadata, support/diagnostic output, and Web UI trace/recovery rendering helpers.

## Fixture

Use one reusable forbidden-value fixture with:

- Prompt/user text.
- Assistant text.
- Tool args with raw request/response-like bodies.
- Secret keys and values: API keys, bearer tokens, passwords, provider request/response fields, and secret-looking scalar values.
- Workflow step metadata and replay/control metadata where applicable.

## Test Targets

1. Backend daemon run-store/HTTP tests verify:
   - `/metrics` remains aggregate-only.
   - `/runs/{id}/trace` and `/traces/export` contain valid redacted JSON but no forbidden values or raw risky keys.
   - `/runs` summary metadata remains safe for control/step/recovery summaries.
   - replay-control response stays redacted.
2. Support/diagnostic tests verify provider/config output surfaces do not reveal configured secrets.
3. Web UI asset tests verify static trace/recovery rendering code does not intentionally render raw prompt/tool/provider payload fields for trace and replay/recovery panels.

## Safety

This task adds tests first. Production changes are limited to fixing leaks if a test proves one. Intentional detailed run views remain out of scope unless they are part of metrics, trace, summary, support, or handoff-style surfaces.

## Rollback

Remove the new regression tests and any leak fixes. No external services or persistent state are introduced.
