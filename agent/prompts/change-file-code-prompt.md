Implement Change spec `specs/<change-slug>.md` with senior-level discipline.

The Change spec defines the desired future state and implementation scope. Current code defines
current behavior and technical contracts. Implement the smallest coherent code, test,
configuration, seed, and explicitly authorized database-file changes needed to satisfy the Spec.

Documentation is outside the default Change Flow. Do not inspect, create, update, or reconcile
documentation unless the user explicitly requests documentation work or the active Spec explicitly
includes it.

Before coding:
1. Read the full Change spec.
2. Inspect the initial code and tests that define current behavior.
3. Read a retained ADR or operational runbook only when the Spec explicitly cites it as a required
   decision or operational constraint.
4. Inspect the existing backend, frontend, CLI, configuration, database, seed, and test patterns before choosing an approach.
5. Inspect the current worktree and preserve unrelated local changes.

Stop conditions:
- If the required external behavior, API contract, persistence contract, field naming, endpoint naming, history behavior, seed behavior, or verification expectation is unclear, stop and ask one specific clarifying question.
- If unrelated local changes block a safe implementation, stop and describe the conflict.
- If database behavior blocks implementation, report the blocker instead of mutating live/local database state outside approved verification commands.

Hard rules:
- Do not broaden scope beyond the Change spec.
- Do not refactor unrelated code.
- Do not revert or overwrite unrelated local changes.
- Follow existing project architecture, naming, transaction, DTO, API, frontend, CLI, and test patterns.
- Keep implementation scoped to the files required by this Change.
- Do not create foreign keys.
- Do not introduce broad locking, advisory locks, isolation escalation, or cross-path locking unless explicitly required by the Change.
- Do not mutate any live/local database manually. Only use repository Make targets that operate on
  disposable test databases when needed.
- Do not weaken, skip, delete, rebaseline, or bypass tests to make verification pass.

Verification:
- First run focused tests for touched behavior.
- Then run every required verification command from the Change spec for each touched area.
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
