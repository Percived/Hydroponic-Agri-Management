# AGENTS.md

## Session Startup Context Rule

Before doing full repository scans, read these files first:
1. `docs/PROJECT_STATUS.md`
2. `docs/HANDOFF.md`

If the user request can be answered using these files, avoid full-codebase traversal.
Only do deep scanning when:
- the user explicitly asks for code-level verification,
- the context docs are stale or missing required details,
- or implementation work requires touching specific modules.

## Context Hygiene

- Prefer incremental reads over full reads.
- Prefer targeted file opens based on module index in `docs/PROJECT_STATUS.md`.
- After making meaningful code changes, update `docs/HANDOFF.md` (required) and `docs/PROJECT_STATUS.md` (if scope/status changed).
 