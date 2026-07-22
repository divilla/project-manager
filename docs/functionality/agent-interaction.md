# Agent Interaction

## Purpose
Agents help refine planning, maintain documentation, implement scoped changes, and run
verification. Before implementation, they use the Change spec as the requested scope. During and
after implementation, they treat the complete branch diff or published PR code as accepted
behavior and reconcile the Spec and documentation to it.

## Commands
Supported workflow prompts:

- `make-change new <change-slug-or-path>`
- `make-change commit`
- `make-change implement`
- `make-change pr`

Repository Change workflow branches use `change/<change-slug>`. The matching idea and Change spec live at `agent/ideas/<change-slug>.md` and `specs/<change-slug>.md`.

Workflow prompts that operate on an existing Change must fail fast with a clear error when the current branch is not `change/<change-slug>`.

## Change Artifacts
A Change is the full delivery flow. Its artifacts include the branch, idea, spec, docs, code, PR body, PR, review, and follow-up fixes.

The Change, Idea, and Spec share one title. When the idea title changes through an agent-assisted idea workflow, the Change title is updated with it. When the spec title changes through a supported spec edit workflow, the Idea and Change titles are updated with it.

The canonical Change spec structure template is `.mch/default/prompts/spec-file-structure.md`. Active workflow prompts and agent instructions must use that path, not legacy prompt locations under `agent/prompts`.

`mch` refreshes `.mch/default/prompts/change-types.md` from the startup-loaded Change type catalog. Writing prompts such as `spec-write.md` may read that file to select allowed slugs. The canonical structure template keeps `Types:` optional and validates only metadata placement and formatting, so review, implementation, documentation, and reconciliation prompts do not need the catalog. Prompt Markdown must not call backend APIs.

## Planning Behavior
During planning, the agent:

- creates or checks out the matching branch
- commits rough user edits
- rewrites the Change spec into the standard structure
- updates or links relevant docs
- commits the agent checkpoint

## Implementation Behavior
During implementation, the agent:

- reads the current Change spec
- reads referenced docs
- verifies readiness
- changes only files needed for the Change
- records follow-ups instead of silently expanding scope
- runs verification when feasible
- runs `make lint` and fixes all findings after code changes in `backend` or `cli`
- commits with the implementation message

## Autonomy
The agent may edit code, docs, and tests within the active Change. It should stop when a product decision is missing, when docs conflict with requested behavior, or when unrelated worktree changes make the workflow unsafe.

## Database Safety
Agents must treat the repository-root `db` folder as read-only unless the user explicitly requests a specific database-file change. This applies to every file and subfolder under `db`.

Agents must not run PostgreSQL commands that alter database structure, including create, alter, drop, truncate, grant, revoke, migration, or restore operations, unless the user explicitly requests that exact structural change.

When a database function, procedure, schema object, seed file, or backup appears incorrect or blocks implementation, the agent must report the blocker and adapt only application or test code that is within scope.

## Text Quality
Generated Change and documentation text should be grammar-checked, readable, and concise.
