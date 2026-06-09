# Astria Kocoro parity phase 19 design

## Architecture Boundary

Phase19 keeps updater work in validation and dry-run mode. It may verify signed
metadata and explain a replacement decision, but it must not replace the app or
bundled daemon.

## Native/Release Areas

- Signed updater dry-run: parse metadata, verify checksum/signature fields, and
  return an explicit no-replacement decision.
- Release compatibility manifest: describe app version, daemon version, and
  compatibility requirements in a credential-free artifact.
- OS crash artifacts: collect local crash evidence only when user-triggered and
  redacted.

## Compatibility

- Existing CLI self-update remains separate from Astria app update behavior.
- Existing release validation remains credential-free.
- Missing updater metadata remains unavailable-safe.

## Rollout

1. Signed updater no-replacement dry-run.
2. Release compatibility manifest.
3. Local OS crash artifact collection.
4. Final Kocoro gap review.
