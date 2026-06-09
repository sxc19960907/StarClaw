# Astria sandbox updater rehearsal fixture design

## Boundary

This is a shell-script validation helper inside
`scripts/validate_release_artifacts.sh`. It creates fake bundle directories and
uses normal filesystem operations only under a temporary sandbox.

## Rehearsal Steps

1. Create `current/Astria.app` with version marker `1.0.0`.
2. Create `candidate/Astria.app` with version marker `1.1.0`.
3. Copy candidate into `staging/Astria.app`.
4. Move current install fixture aside as rollback source.
5. Copy staged candidate into `install/Astria.app`.
6. Verify install marker is candidate version.
7. Roll back from backup to `install/Astria.app`.
8. Verify install marker is previous version.
9. Assert every touched path has the sandbox prefix.

## Output

The smoke may emit only a success line; internal JSON output is optional for
this first child as long as guard failures are explicit.
