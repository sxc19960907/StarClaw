# Design

Phase7 extends the Phase6 local runtime into Kocoro-style delivery behavior:

- `cloudflow` boundary for slash command parsing/display/dispatch.
- Route index and connection-state cache for future IM/channel replies.
- Durable system events and suggestion events for diagnostics/UI.
- Deeper delivery injection lifecycle for busy/orphan/re-enqueue cases.
- External adapter interfaces only after local contracts are stable.

The parent task owns the roadmap. Children own implementation.

