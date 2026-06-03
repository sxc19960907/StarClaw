# Fix manual GUI experience issues

## Goal

Fix the concrete GUI experience issues found during manual browser testing.

## Requirements

- Chat empty-state should not claim StarClaw is ready when diagnostics needs setup.
- Version page should not present update checks as an active primary action when update checks are unavailable.
- Permissions pending preview should be easier to scan and avoid cramped label/value wrapping.
- Agent editor should give the editor more room and make Test Runner easier to reach.
- Toasts should disappear when navigating to another panel so stale action feedback does not linger.
- Sidebar navigation should group high-frequency workspace actions separately from build/system panels so the left rail is easier to scan.
- Sidebar navigation should stay focused on chat/session work, with lower-frequency tools moved into Manage and Settings hub panels.

## Acceptance Criteria

- [x] Chat empty-state reflects diagnostics needs-setup status.
- [x] Update check button is disabled for unsupported development builds.
- [x] Permissions pending preview uses compact, readable category rows.
- [x] Agent editor layout gives the detail pane more width on desktop.
- [x] Test Runner appears before lower-frequency command/import sections.
- [x] Toast clears on panel navigation.
- [x] Sidebar navigation is grouped into Workspace, Build, and System sections.
- [x] Sidebar navigation is reduced to Chat, Runs, sessions, Manage, and Settings.
- [x] Manage hub provides Agents, Skills, and Schedules entry points.
- [x] Settings hub provides Diagnostics, Config, Permissions, and Version entry points.
- [x] Targeted Web UI checks pass.
