# Validate daemon API smoke workflow

## Goal

Validate daemon HTTP API smoke workflow for status, agents, skills, sessions, schedules, and stable JSON contracts.

## Requirements

- Add a deterministic daemon API smoke test from an external client perspective using `httptest`.
- Cover the API surfaces needed by a future GUI shell:
  - `/health` and `/status`
  - `/agents` and `/agents/{name}`
  - `/skills`
  - `/sessions` and `/sessions/search`
  - `/schedules` create/list/get/update/delete
  - representative error responses for bad or missing resources
- Implement missing read-only skills listing route if absent.
- Keep all tests offline and isolated from the developer's real StarClaw home.
- Preserve existing daemon endpoint behavior unless a route is missing or its response shape is unstable.
- Leave unrelated untracked workspace files untouched.

## Acceptance Criteria

- [x] `GET /skills` returns stable JSON with a `skills` array, including name, description, and source for valid skills.
- [x] A daemon smoke test exercises health/status, agents, skills, sessions/search, and schedule CRUD in one isolated server.
- [x] Smoke assertions verify response status codes and key JSON field names for GUI/client compatibility.
- [x] Missing agent/schedule/session-like requests return appropriate non-2xx JSON errors.
- [x] `go test ./internal/daemon` passes.
- [x] `go test ./...` and `go vet ./...` pass.

## Notes

- Evidence from code inspection: daemon currently registers health, message, schedule, agent, config, instructions, session, and permission routes, but no `/skills` route despite `ServerDeps.SkillsDir` existing and README advertising daemon skills capability.
