# Image tool provider boundary

## Goal

Close the Phase 9 image generation/editing tool gap against local Kocoro commit `74cdb3c` by adding StarClaw-owned provider-gated image tool contracts and a typed image provider client. The capability must remain disabled by default unless a caller explicitly wires a provider client.

## Requirements

- Use `/Users/timmy/PycharmProjects/Kocoro/internal/images/client.go`, `internal/tools/generate_image.go`, and `internal/tools/edit_image.go` as the parity reference.
- Add an `internal/images` package with typed request/response structs, sentinel errors, response classification, and bounded retries for provider-backed image generation/editing endpoints.
- Add `generate_image` and `edit_image` tools that:
  - Validate prompt length by rune count.
  - Validate size, quality, background, `n`, and edit source URL limits before calling the provider.
  - Require approval and never auto-approve safe args.
  - Classify provider errors into StarClaw tool error categories.
  - Warn in tool descriptions and results that provider outputs may be permanent public URLs.
- Add explicit image tool registration that requires a non-nil provider client.
- Do not register provider-backed image tools from `RegisterLocalTools`.
- Do not enable Shannon/Kocoro credentials, cloud sync, public CDN upload, or off-machine telemetry by default.
- Keep StarClaw naming in code and user-facing text.
- Preserve existing local `imaging` and `publish_to_web` behavior.

## Acceptance Criteria

- [ ] `RegisterLocalTools()` does not expose `generate_image` or `edit_image`.
- [ ] `RegisterImageTools(reg, client)` registers `generate_image` and `edit_image` only when both registry and provider client are non-nil.
- [ ] `internal/images.Client` posts JSON to generation/edit endpoints and sends an API key header only when configured.
- [ ] Image client retries transient provider/network failures and does not retry permanent validation, auth, endpoint, timeout, moderation, or source-size failures.
- [ ] Image tools validate arguments client-side and do not call the provider on invalid input.
- [ ] `generate_image` and `edit_image` require approval and implement `SafeChecker` with `false`.
- [ ] `edit_image` restricts source URLs to the explicitly configured provider CDN prefix.
- [ ] Tests cover client success/error/retry behavior, tool validation, error mapping, approval/safe-checker behavior, and registration boundaries.
- [ ] `go test ./internal/images ./internal/tools` passes.

## Notes

- Parent task: `.trellis/tasks/06-08-astria-kocoro-parity-phase-9`
- Reference research: `.trellis/research/kocoro-native-tool-parity-phase9-plan.md`
