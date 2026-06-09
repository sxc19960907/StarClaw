# Astria updater rollback health gates design

## Boundary

This task adds validation artifacts only. It does not execute rollback,
replacement, launch, daemon mutation, or Desktop RPC calls.

## Manifest Shape

The manifest is credential-free JSON:

- `schema_version`
- `product`
- `local_only`
- `replacement`
- `rollback`
  - `gate_id`
  - `required`
  - `source`
  - `restore_target`
  - `daemon_compatibility_guard`
  - `manual_approval_required`
- `post_update_health`
  - `gate_id`
  - `required`
  - `checks`

## Required Checks

- `app_launch`
- `daemon_health`
- `desktop_rpc_capabilities`
- `web_ui_readiness`

## Validation

The smoke should cover valid, incomplete rollback, and incomplete health-check
manifests without requiring Apple credentials.
