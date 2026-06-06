# Real Channel Provider Design

## Provider Choice

Use GitHub webhooks as the first real provider. This avoids OAuth setup for MVP, works with local tunnel tools, and maps issue/comment events into reviewable Inbox items.

## API Surface

- `POST /inbox/github` ingests GitHub webhook payloads.
- Supported events:
  - `issues`
  - `issue_comment`
- The handler reads standard GitHub headers:
  - `X-GitHub-Event`
  - `X-GitHub-Delivery`
  - `X-Hub-Signature-256`

## Deduplication

The Inbox store already deduplicates by `provider + external_id`.

External IDs:
- issues: `issue:<repository.full_name>:<issue.id>:<action>`
- issue comments: `issue_comment:<repository.full_name>:<comment.id>:<action>`

## Metadata

Preserve:
- repository full name
- action
- issue number
- issue/comment HTML URL
- sender login
- delivery ID

## Signature Verification

For this MVP, signature verification is optional and activated only when a local environment variable is set:

- `STARCLAW_GITHUB_WEBHOOK_SECRET`

If configured, unsigned or invalid requests return `401`. If not configured, unsigned requests are accepted so local smoke/tests remain credential-free.

## UI

Expose GitHub provider setup state in the Inbox side panel:
- endpoint path
- secret state from `/inbox/providers`
- supported events
- guarded execution note

## Compatibility

Existing `/inbox/webhook` remains unchanged.
No hosted relay or background polling is introduced.
