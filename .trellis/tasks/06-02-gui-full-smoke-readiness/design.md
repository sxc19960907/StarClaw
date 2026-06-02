# Design

This is a validation/fix task. The implementation path depends on full smoke results.

Expected fix categories:

- overly broad Playwright locators after recent UI additions;
- stale UI expectations from recent flow changes;
- small GUI state regressions surfaced by the full suite.

No backend API or product scope changes are planned unless full smoke exposes a direct blocker.
