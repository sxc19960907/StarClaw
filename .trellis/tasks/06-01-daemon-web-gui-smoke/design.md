# Design

## Approach

Create `scripts/smoke_webui.sh`, mirroring the style of `scripts/smoke_cli.sh`:

- build `./main.go` into a temporary binary
- create an isolated HOME with a minimal local config
- start `starclaw daemon start`
- wait for `/health`
- run browser checks via a Node script that uses Playwright
- stop the daemon through `/shutdown` and process cleanup

## Browser Smoke

Use a temporary Node script invoked through `npx --package playwright`.

The browser flow:

1. Open `http://127.0.0.1:7533/app/`.
2. Assert core UI elements: sidebar, chat composer, status strip, panels.
3. Navigate to Schedules, create a schedule, pause/enable it, delete it.
4. Subscribe to `/events` in the page.
5. Trigger an approval event by opening a fetch to `/events` and then posting a synthetic pending approval directly to the daemon broker is not possible through public API, so use the browser to verify the approval UI function with a controlled page-side `MessageEvent` only if public daemon API cannot create a pending approval without LLM.

## Approval Smoke Boundary

The real daemon only emits approval events from an active agent run. A full agent run would require a model client. For no-provider smoke, validate:

- `/events` connection is available.
- the browser can render approval cards and call `/approval` for a nonexistent request without crashing.

Backend unit tests already cover broker resolution for real pending approvals.

## Artifacts

Screenshots go under `output/playwright/daemon-webui-smoke.png`.

## Compatibility

The script is opt-in and not wired into default CI. It should fail fast with clear output if `npx` cannot run Playwright.
