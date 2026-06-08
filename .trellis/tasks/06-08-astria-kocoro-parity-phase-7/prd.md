# Astria Kocoro parity phase 7: channel and cloud delivery parity

## Goal

Continue Kocoro parity after Phase6 by closing channel/cloud delivery gaps: cloudflow dispatch contracts, route indexing, connection state, system events, delivery injection lifecycle, and external channel adapter boundaries.

## Constraints

- Keep implementation local-first unless external cloud/channel transport is explicitly approved.
- Do not add real Feishu/Slack/Telegram credentials, cloud uploaders, or off-machine transport in this parent by default.
- Preserve StarClaw naming for code/packages/docs; Astria remains product/UI-facing.
- Use `/Users/timmy/PycharmProjects/Kocoro` commit `74cdb3c` as local comparison evidence until refreshed.

## Child Order

1. `cloudflow-dispatch-contract`
2. `channel-route-index-connection-state`
3. `system-event-store-suggestions`
4. `delivery-inject-lifecycle-depth`
5. `external-channel-adapter-boundaries`

## Acceptance Criteria

- Each child is independently planned, implemented, tested, committed, and archived.
- Parent remains the source of Phase7 scope and ordering.
- Kocoro parity research is updated as children complete.

