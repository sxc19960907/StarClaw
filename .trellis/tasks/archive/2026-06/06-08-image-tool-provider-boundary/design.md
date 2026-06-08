# Image tool provider boundary design

## Reference

Kocoro exposes image generation/editing through:

- `/Users/timmy/PycharmProjects/Kocoro/internal/images/client.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/generate_image.go`
- `/Users/timmy/PycharmProjects/Kocoro/internal/tools/edit_image.go`

StarClaw will add the same capability shape but keep it behind explicit provider wiring. The default local registry remains local-first and must not expose paid/cloud image tools.

## Package boundary

New package: `internal/images`

Responsibilities:

- Define `GenerateRequest`, `EditRequest`, `Image`, and `GenerateResponse`.
- Define sentinel errors for auth, bad request, request too large, endpoint missing, upstream timeout, content rejected, server config, transient, invalid edit URL, and source too large.
- Provide `Client.Generate(ctx, req)` and `Client.Edit(ctx, req)`.
- POST JSON to:
  - `/api/v1/images/generations`
  - `/api/v1/images/edits`
- Retry only errors wrapping `ErrTransient`.
- Preserve context cancellation/deadline errors.
- Avoid logging or exposing API keys.

## Tool boundary

New tools:

- `generate_image`
- `edit_image`

Tool registration:

- `RegisterImageTools(reg, client)` in `internal/tools/register.go`.
- No-op if registry or client is nil.
- `RegisterLocalTools()` does not call it.

Tool contracts:

- Both tools require approval.
- Both implement `agent.SafeChecker` and always return false.
- Both return tool errors as `agent.ToolResult`, not Go errors.
- Both format successful provider responses as compact URL lines plus metadata.
- Tool descriptions explicitly warn that provider outputs may be permanent public URLs.

Validation:

- Prompt trim + rune count max 32000.
- Size enum: `1024x1024`, `1024x1536`, `1536x1024`, `auto`.
- Quality enum: `auto`, `low`, `medium`, `high`.
- Background enum: `transparent`, `opaque`, `auto`.
- `n`: 0 allowed as provider default; otherwise 1..10.
- `edit_image.image_urls`: 1..4 entries.
- `edit_image` requires every source URL to start with a configurable provider CDN prefix. Initial StarClaw default: `https://static.kocoro.ai/`, matching the Kocoro/Shannon provider contract, but only used once a provider client is explicitly wired.

## Local-first behavior

This task does not:

- Add image provider fields to default config.
- Read Shannon/Kocoro credentials.
- Register provider tools by default.
- Upload local files to public CDN.
- Enable sync or telemetry.

Future callers can explicitly build `images.NewClient(endpoint, apiKey, httpClient)` and call `tools.RegisterImageTools`.

## Rollback

Rollback is deleting `internal/images`, deleting the image tool files/tests, and removing `RegisterImageTools`. Because no default config or persisted data is changed, rollback has no migration step.
