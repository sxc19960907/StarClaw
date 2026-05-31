# Design

## Architecture

The first GUI is a daemon-hosted static web app embedded into the Go binary:

- `internal/daemon/webui.go` owns embedded assets and HTTP handlers.
- `internal/daemon/webui/` contains `index.html`, `styles.css`, and `app.js`.
- `internal/daemon/router.go` registers `/`, `/app`, `/app/`, and `/app/assets/*`.
- The browser calls existing daemon JSON endpoints with same-origin fetch.

## Visual Direction

The UI follows a Codex-like local workbench:

- left sidebar for workspace modules
- top status strip for daemon/version/activity
- central command/chat surface
- right inspector for agents, skills, sessions, and schedules
- compact typography, 8px-or-less radii, clear borders, subtle state colors
- no hero section, no decorative gradients or blobs

## API Usage

- Status: `GET /status`
- Chat: `POST /message`
- Agents: `GET /agents`, `GET /agents/{name}`
- Skills: `GET /skills`
- Sessions: `GET /sessions`, `GET /sessions/search?q=...`
- Schedules: `GET /schedules`, `POST /schedules`, `PATCH /schedules/{id}`, `DELETE /schedules/{id}`

## Compatibility

- Existing daemon API behavior remains unchanged.
- Static UI works without external network assets.
- GUI can show API errors when provider config is missing; it should still render and allow non-LLM API browsing.

## Testing

- Go route tests verify redirects, HTML, and asset serving.
- Full Go test/vet stays green.
- Browser rendering is checked with Playwright against a local static/daemon server.
