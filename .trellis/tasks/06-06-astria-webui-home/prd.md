# Rebrand Web UI to Astria with Celestial Home Launcher

## Summary

Rebrand the daemon Web UI surface from StarClaw to Astria and add a Kocoro-inspired home task launcher with Astria-specific celestial styling. This task is UI/brand layer only: do not rename the Go module, binary, npm package, release artifacts, config paths, or backend API contracts.

## Goals

- Make `starclaw app` open to a polished Astria home screen rather than directly to chat.
- Preserve existing chat, runs, agents, skills, schedules, diagnostics, config, permissions, and version workflows.
- Introduce a distinct Astria visual language: light macOS-style shell, celestial accents, orbit/status concepts, and a central mission/task composer.
- Improve discoverability of core capabilities through home shortcut chips and recommendation cards.

## Non-Goals

- No CLI command rename from `starclaw` to `astria`.
- No package/repo/module/release artifact rename.
- No backend data model changes unless needed to support existing UI data display.
- No new external frontend dependencies.

## Requirements

- Default panel is `Home`.
- Sidebar brand displays `Astria` with a celestial mark and the current shell keeps a native desktop-app feel.
- Sidebar navigation is reorganized into clear workflow/capability groups while preserving access to all existing panels.
- Home screen includes:
  - Astria welcome headline.
  - Current activity counts for pending approval, running, completed, and failed runs.
  - Large central task composer that submits through the existing chat flow.
  - Agent selector, new-session behavior, working-folder affordance, stop/send states.
  - Capability shortcut chips for schedules, browser, data, writing, research, multi-agent, desktop control, local files, and MCP.
  - Lightweight Astria capability cards for memory map, mission orbit, and MCP docking.
- Chat panel remains available for transcript-focused work.
- Existing E2E smoke scripts should still be able to submit a message and inspect output.
- Responsive layout must remain usable at desktop and narrow widths.

## Acceptance Criteria

- [x] Opening the Web UI shows `Astria` and the new home launcher as the default view.
- [x] Submitting a prompt from Home creates/streams a chat run using the existing `/chat/stream` flow.
- [x] Existing Chat panel still supports manual message submission, agent selection, session selection, stop, and approvals.
- [x] Existing Runs/Agents/Skills/Schedules/Settings panels remain reachable from sidebar or home cards.
- [x] Visual styling includes Astria-specific celestial accents without turning the UI into a dark sci-fi theme.
- [x] `go test ./...` passes.
- [x] Existing Web UI smoke coverage is run where feasible.
