# Run GUI End-to-End User Regression

## Problem

The GUI has broad smoke coverage, but after the recent agent editor, run detail, launch status, and npm installer work, we need a real user-flow regression pass to catch interaction or layout issues that scripted assertions may miss.

## User Value

Before calling the project close to releasable, the main GUI path should work as a user would experience it: launch, inspect readiness, configure provider, create/test an agent, run chat, inspect history/detail, and verify persistence/navigation.

## Scope

- Use the existing local deterministic smoke setup when possible.
- Open the GUI in a real browser and inspect screenshots/DOM state.
- Exercise a first-run style path across Diagnostics, Config, Agents, Chat, Sessions, Runs, Permissions, and Version.
- Fix discovered regressions that are clearly in scope.

## Out of Scope

- Testing a real paid LLM provider.
- Publishing artifacts.
- Large redesign work unless a blocking usability issue appears.
- Waiting on GitHub push/CI if network is unavailable.

## Acceptance Criteria

- [x] GUI launches locally and loads without console/page fatal errors.
- [x] Diagnostics and Version pages show launch/runtime context.
- [x] Provider config can be edited and saved.
- [x] Agent can be created/edited, command editor works, import/export remains usable.
- [x] Agent Test can run and show result/run/session actions.
- [x] Chat can run a prompt, show streaming/tool events where applicable, and persist a session.
- [x] Runs panel opens detail, grouped tool event, copy actions, re-run, and session navigation.
- [x] Permissions page loads global editor state.
- [x] Screenshots are captured for review under `output/playwright/`.
- [x] Any discovered blocking GUI issue is fixed or explicitly recorded as follow-up.
