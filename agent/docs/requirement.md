## Requirement

- Defines the target.
- Says what must become true.
- States a condition, capability, constraint, or quality the system must satisfy. 
- Can be functional or non-functional. 
- Can be high-level or detailed.
- Decompose until each piece is clear, implementable, and verifiable.


Example requirement:

> Users can reset their password by email.

Decomposed requirements:

- User can request a reset link.
- Reset link expires after 30 minutes.
- Reset link can be used only once.
- User can set a new password.
- Old sessions are revoked after reset.

> An AI coding agent tends to perform better when the task is expressed as hierarchical requirements plus references to relevant specifications, especially for non-trivial work.

## Definition of Done 

- Defines whether the work is shippable.
- Says what conditions must be met before we consider the work finished.
- Includes quality, process, and delivery criteria.
- Shared quality bar applied to work items.

Each of those can have acceptance criteria. But Definition of Done is usually broader:

- Requirement implemented.
- Tests pass.
- Code reviewed.
- Security concerns handled.
- UI copy reviewed.
- Logs/metrics added if needed.
- Documentation updated if relevant.
- Deployed or releasable.

In agile terms:

- Acceptance criteria are specific to a story/requirement.
- Definition of Done is a 



## Task

Goal:
- What outcome we want.

Requirements:
- Functional requirement 1
  - Acceptance criteria
  - Edge cases
- Functional requirement 2
  - Acceptance criteria

Constraints:
- Architecture constraints
- Existing APIs/specs to follow
- Things not to change

References:
- Design/spec document
- API contract
- Existing module/file examples
- Tests that should pass

Why this helps:

- Hierarchy gives scope control. The AI can see the big picture and the smaller obligations.
- Specifications reduce ambiguity. They say what “correct” means.
- References anchor the work in reality. They prevent the AI from inventing APIs, styles, or behavior.
- Acceptance criteria make verification easier. The AI can translate them into tests or manual checks.

But there is a tradeoff: overly long or contradictory requirements can hurt. The best form is not “maximum detail”; it is structured, relevant detail.

A useful hierarchy for AI work is:

Objective
Context
Requirements
Acceptance criteria
Non-goals
Relevant files/specs
Verification commands

Example:

Objective:
Add task deletion confirmation to the task board.

Requirements:
1. Clicking delete must not immediately call the delete API.
2. A persistent confirmation dialog must open.
3. Cancel closes the dialog without deleting.
4. OK calls the existing delete endpoint.
5. After success, reload the current project tasks.

Constraints:
- Reuse DeleteConfirmationDialog.vue.
- Do not change backend APIs.
- Keep existing board layout.

Verification:
- Add/update component tests.
- Run pnpm --dir frontend test.



## Pull Request

### Goal
What user/system outcome this PR delivers.

### Scope
What is included.

### Requirements
- Requirement 1
  - Acceptance: ...
- Requirement 2
  - Acceptance: ...

### Non-Goals
What this PR intentionally does not change.

### Specifications
- Relevant API contracts
- Relevant architecture decisions
- Any module boundaries touched

### Verification
- `make test`
- `make api-test`
- `pnpm --dir frontend test`
- Manual checks, if any

### Risks / Review Focus
- Areas where reviewer attention is useful

Example:

### Goal
Add dedicated frontend task workflows for task search, detail, creation, editing, deletion confirmation, requirements editing, and markdown-rendered task descriptions.

### Requirements Implemented
- Task board search supports name, type, and phase filters.
- Task detail route `/tasks/:id` renders task hierarchy, children, requirements, and markdown descriptions.
- Task create route `/tasks/create/:parentId` supports root and child task creation.
- Task edit route `/tasks/edit/:taskId` supports name, type, phase, and markdown description editing.
- Task and requirement deletion use persistent confirmation dialogs.
- Backend renders sanitized markdown into `description_html`.

### Non-Goals
- No schema changes.
- No changes to task type or task phase reference data.
- No backend auth changes.

### Design Notes
- Backend markdown rendering is isolated behind `backend/pkg/markdown`.
- `description` remains raw markdown; `description_html` is rendered/sanitized HTML.
- Project-scoped task data is cached in `taskCache.store`.

### Verification
- `make test`
- `make api-test`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`

### Review Focus
- Project-switch behavior on nested task routes.
- Create/edit route behavior when opened directly.
- Markdown sanitization and rendering boundaries.

The important practice is: don’t make the PR description just a change log. Make it traceable:

> Goal → requirements → implementation boundaries → verification.

For AI-assisted PRs especially, I’d add one more section:

## Specification References
- `agent/features/08-frontend-tasks.md`
- `agent/specs/...`

That lets reviewers compare the implementation against the intended behavior instead of reverse-engineering intent from the diff.
