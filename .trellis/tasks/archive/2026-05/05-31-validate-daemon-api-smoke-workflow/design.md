# Design

## Scope

This task validates the daemon as an HTTP API surface for future GUI and integration clients. It adds a read-only skills listing endpoint because the existing daemon dependency set already includes `SkillsDir`, while no route exposes it.

## API Contract

- `GET /skills`
  - Response: `200 OK`
  - Body: `{"skills":[...]}` where each entry is `skills.SkillMeta`.
  - Empty or missing skills directory returns an empty array, not `null`.
  - Invalid skill metadata returns `500` with the existing daemon error envelope.

Existing endpoints keep their current response shapes:

- `GET /health`: `status`, `version`
- `GET /status`: `uptime`, `version`, `active_agents`
- `GET /agents`: `agents`
- `GET /agents/{name}`: loaded agent payload
- `GET /sessions`: `sessions`
- `GET /sessions/search?q=...`: `results`
- Schedule CRUD: concrete `schedule.Schedule` objects and `{"schedules":[...]}`

## Implementation Boundaries

- Add skills route registration in `internal/daemon/router.go`.
- Add a small handler in `internal/daemon/server.go`.
- Use `internal/skills.ListSkills` directly with `deps.SkillsDir`.
- Add focused tests in `internal/daemon/server_test.go`, following existing `httptest` style.

## Compatibility

- No authentication or write-capable skills management is introduced.
- No real daemon process is started; tests use `httptest.NewServer`.
- No network, provider credential, or user home dependency is required.
