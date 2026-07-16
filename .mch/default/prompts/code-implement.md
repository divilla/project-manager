Extract `<change-slug>` from the current Git branch using this regex: `^change/([0-9]+-[a-z0-9-_]+)$`

If the current branch does not match, stop and output one concise error explaining that the branch must be named `change/<change-slug>`.

Stop without editing if `specs/<change-slug>.md` does not exist.

Implement Change spec `specs/<change-slug>.md` with senior-level discipline.

The Change spec is the implementation contract. The current branch documentation under `docs/` is the behavioral reference. Implement the smallest coherent code, test, seed, and database-file changes needed to satisfy that contract.

Before coding:
1. Read the full Change spec.
2. Read every relevant doc under `docs/`, especially the docs listed in the Change spec.
3. Compare the Change spec and docs for conflicts.
4. Inspect the existing backend, frontend, CLI, database, seed, and test patterns before choosing an approach.
5. Inspect the current worktree and preserve unrelated local changes.
6. Verify the Change spec is implementation-ready: it must follow the standard Change structure, define a clear Goal, Scope, Requirements, Acceptance Criteria, Non-Goals, Relevant Specs, Verification, and QA Test Cases, and contain enough detail to implement and verify the behavior without relying on chat history. 

Stop conditions:
- If the Change spec and docs conflict, stop before coding and report the exact file/section conflict.
- If any Change spec required section is missing, ambiguous, untestable, or conflicts with the linked docs, stop before coding and report the exact readiness gap.
- If the required external behavior, API contract, persistence contract, field naming, endpoint naming, history behavior, seed behavior, or verification expectation is unclear, stop and ask one specific clarifying question.
- If unrelated local changes block a safe implementation, stop and describe the conflict.
- If database behavior blocks implementation, report the blocker instead of mutating live/local database state outside approved verification commands.

Hard rules:
- Do not broaden scope beyond the Change spec and docs.
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
- Run every verification command required by the Change spec and by the repository rules for each touched area; if the Change spec omits a required repository verification command, run the repository-required command anyway and report the omission.
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
