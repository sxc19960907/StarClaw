# Design

## Boundary

This task defines adapter contracts, registry behavior, and diagnostics. It is not a transport implementation.

Package location: `internal/daemon`, because the existing local queue, inbox, route, and channel APIs already live there and Phase 7 is daemon-runtime scoped.

## Types

Add `channel_adapter.go`:

- `ChannelAdapter` interface:
  - `Metadata() ChannelAdapterMetadata`
  - `Install(ctx, ChannelInstallRequest) (ChannelInstall, error)`
  - `ListInstalls(ctx) ([]ChannelInstall, error)`
  - `DeleteInstall(ctx, id string) error`
- `ChannelAdapterMetadata`:
  - provider
  - display name
  - kind
  - configured
  - enabled
  - capabilities
  - privacy note
- `ChannelInstallRequest`:
  - provider
  - agent
  - display name
  - metadata map
- `ChannelInstall`:
  - id
  - provider
  - agent
  - display name
  - metadata map
  - created at

## Registry

`ChannelAdapterRegistry`:

- register adapter by provider.
- list metadata sorted by provider.
- get adapter by provider.

Register fake local adapters in `NewServer`:

- `feishu`
- `slack`
- `telegram`
- `webhook`

They should be disabled/configured=false except webhook/local fake where configured can be true. None may perform network I/O.

## API

Add:

- `GET /channel/adapters`

Response:

```json
{
  "adapters": [
    {
      "provider": "feishu",
      "display_name": "Feishu/Lark",
      "kind": "external",
      "configured": false,
      "enabled": false,
      "capabilities": ["install", "list", "delete"],
      "privacy_note": "Disabled local contract; no external transport is active."
    }
  ]
}
```

Management POST/DELETE endpoints are not required in this slice; fake lifecycle is proven through unit tests to avoid exposing misleading public API before credential policy exists.

## Safety

- Metadata must never include secrets.
- Fake install metadata is copied defensively.
- Registry/list output is read-only and content-free.
