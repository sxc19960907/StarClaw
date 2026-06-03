# Ignore generated local artifacts

## Goal

Keep the working tree focused by ignoring local generated artifacts that should not be committed.

## Requirements

- Ignore generated screenshot/log output under `output/`.
- Ignore the local untracked `obsidian-cli` agent skill directory.
- Do not delete or modify the existing local files in either ignored path.
- Keep the change limited to repository hygiene; no runtime code changes.

## Acceptance Criteria

- [x] `git status --short` no longer lists `output/`.
- [x] `git status --short` no longer lists `.agents/skills/obsidian-cli/`.
- [x] The only tracked change needed for the ignore behavior is `.gitignore`.

## Notes

- This is a lightweight task; PRD-only planning is sufficient.
