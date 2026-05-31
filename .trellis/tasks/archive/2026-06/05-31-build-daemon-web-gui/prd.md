# Build daemon web GUI

## Goal

Design and implement a Codex-inspired daemon-hosted Web GUI for StarClaw status, chat, agents, skills, sessions, and schedules.

## Requirements

- Add a daemon-hosted Web GUI available at `/app`, with `/` redirecting there.
- Use a Codex-inspired local app style: restrained neutral palette, dense workbench layout, left navigation, compact controls, clear status surfaces, and no marketing/landing page.
- The first screen must be a usable StarClaw console, not a splash page.
- Provide GUI workflows for:
  - daemon status refresh
  - chat/message execution through `/message`
  - agent list and detail viewing
  - skill list viewing
  - session list and search
  - schedule list/create/enable-disable/delete
- Keep implementation dependency-free: embedded static HTML/CSS/JS served by Go daemon.
- GUI must handle loading, empty, and error states without crashing.
- Keep API calls same-origin so it works directly from `starclaw daemon start`.
- Add backend tests for UI routes.
- Validate rendering in a real browser screenshot.
- Leave unrelated untracked workspace files untouched.

## Acceptance Criteria

- [x] `GET /` redirects to `/app`.
- [x] `GET /app` redirects to `/app/`.
- [x] `GET /app/` serves the GUI HTML.
- [x] Static assets under `/app/assets/` are served from embedded files.
- [x] The GUI includes functional controls for status, chat, agents, skills, sessions, and schedules.
- [x] `go test ./internal/daemon ./cmd` passes.
- [x] `go test ./...` and `go vet ./...` pass.
- [x] Browser screenshot confirms the app renders as a nonblank desktop-style workbench.

## Notes

- No separate Node/Vite project for this first GUI pass; the repo currently has no frontend build system and daemon APIs are already available.
