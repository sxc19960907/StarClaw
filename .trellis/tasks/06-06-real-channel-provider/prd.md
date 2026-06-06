# Real Channel Provider

## Goal

Connect Astria's guarded Channel Inbox to one real external provider so tasks can arrive from outside the local Web UI while still requiring explicit user approval before execution.

## Requirements

- Use GitHub issue/comment webhooks as the first real provider because they can be tested locally without OAuth credentials and map cleanly to external task intake.
- Preserve the existing local webhook provider for testing.
- Deduplicate provider events by provider and external ID.
- Preserve sender, thread/link metadata, and delivery status.
- Show provider setup state clearly in Astria.
- Keep all execution guarded by Inbox approval.

## Acceptance Criteria

- [x] One real provider can ingest an inbound message into Inbox.
- [x] Provider setup failure is visible and actionable.
- [x] Duplicate events do not create duplicate pending tasks.
- [x] Approved provider items can become normal runs.
- [x] Rejected items remain auditable.
- [x] Backend tests and Web UI smoke/targeted tests cover the provider path.

## Non-Goals

- No hosted relay.
- No multi-provider framework beyond what the first provider needs.
- No unattended execution.

## Dependencies

- Depends on Channel Inbox MVP.
- GitHub webhook secret verification should be supported when a secret is configured, but local development can accept unsigned events when no secret is configured.
