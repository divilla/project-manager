Implement `agent/changes/109-db-alters-and-views.md` with senior-level discipline.

The Change file is the implementation contract. The current branch documentation under `docs/` is the behavioral reference. Implement the smallest coherent code, test, seed, and database-file changes needed to satisfy that contract.

Before coding:
1. Read the full Change file.
2. Read every relevant doc under `docs/`, especially the docs listed in the Change file.
3. Compare the Change file and docs for conflicts.
4. Inspect the existing backend, frontend, CLI, database, seed, and test patterns before choosing an approach.
5. Inspect the current worktree and preserve unrelated local changes.

Stop conditions:
- If the Change file and docs conflict, stop before coding and report the exact file/section conflict.
- If the required external behavior, API contract, persistence contract, field naming, endpoint naming, history behavior, seed behavior, or verification expectation is unclear, stop and ask one specific clarifying question.
- If unrelated local changes block a safe implementation, stop and describe the conflict.
- If database behavior blocks implementation, report the blocker instead of mutating live/local database state outside approved verification commands.

Hard rules:
- Do not broaden scope beyond the Change file and docs.
- Do not refactor unrelated code.
- Do not revert or overwrite unrelated local changes.
- Follow existing project architecture, naming, transaction, DTO, API, frontend, CLI, and test patterns.
- Keep implementation scoped to the files required by this Change.
- Do not create foreign keys.
- Do not introduce broad locking, advisory locks, isolation escalation, or cross-path locking unless explicitly required by the Change.
- Do not mutate any live/local database manually. Only use documented disposable test-database verification commands when needed.
- Do not weaken, skip, delete, rebaseline, or bypass tests to make verification pass.

Verification:
- First run focused tests for touched behavior.
- Then run every required verification command from the Change file for each touched area.
- If a command fails, report:
    - exact command
    - failing test or error
    - whether local uncommitted changes may have influenced it
    - whether it appears to be a product bug, test bug, environment issue, or unclear contract dependency

Final report:
- Summarize the implemented behavior.
- List the main files changed.
- List verification commands run and their pass/fail status.
- Call out any unresolved follow-ups or risks. 
