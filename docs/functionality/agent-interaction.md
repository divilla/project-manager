# Agent Interaction

## Purpose
Agents help refine planning, maintain documentation, implement scoped changes, and run verification. They operate against the Change file as the contract.

## Commands
Supported workflow prompts:

- `new change <change-name-or-path>`
- `commit change`
- `implement change`

## Planning Behavior
During planning, the agent:

- creates or checks out the matching branch
- commits rough user edits
- rewrites the Change file into the standard structure
- updates or links relevant docs
- commits the agent checkpoint

## Implementation Behavior
During implementation, the agent:

- reads the current Change file
- reads referenced docs
- verifies readiness
- changes only files needed for the Change
- records follow-ups instead of silently expanding scope
- runs verification when feasible
- runs `make lint` and fixes all findings after code changes in `backend` or `make-a-change`
- commits with the implementation message

## Autonomy
The agent may edit code, docs, and tests within the active Change. It should stop when a product decision is missing, when docs conflict with requested behavior, or when unrelated worktree changes make the workflow unsafe.

## Database Safety
Agents must treat the repository-root `db` folder as read-only unless the user explicitly requests a specific database-file change. This applies to every file and subfolder under `db`.

Agents must not run PostgreSQL commands that alter database structure, including create, alter, drop, truncate, grant, revoke, migration, or restore operations, unless the user explicitly requests that exact structural change.

When a database function, procedure, schema object, seed file, or backup appears incorrect or blocks implementation, the agent must report the blocker and adapt only application or test code that is within scope.

## Text Quality
Generated Change and documentation text should be grammar-checked, readable, and concise.
