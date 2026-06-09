# Astria release acceptance gates design

## Boundary

This is a local release metadata gate. It describes what a production Astria
release must prove, but it does not access keychains, Apple services, notary
profiles, private keys, or release upload destinations.

## Manifest Shape

- `schema_version`
- `product`
- `local_validation_credential_free`
- `replacement`
- `distribution`
  - `developer_id_application`
  - `hardened_runtime`
  - `notarization`
  - `stapling`
- `artifacts`
  - `updater_metadata`
  - `compatibility_manifest`
  - `rollback_health_manifest`
  - `transaction_plan`
- `private_material`
  - must be absent or explicitly false booleans only

## Validation

The smoke should cover:

- Valid acceptance manifest.
- Missing signing/notarization/stapling readiness.
- Missing transaction/rollback references.
- Private credential references.
