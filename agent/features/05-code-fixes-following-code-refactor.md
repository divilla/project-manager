# Feature 05: Code Fixes Following Database Refactor

## 1. Purpose

Update the application code to use the already-refactored PostgreSQL contract while treating the database schema, data scripts, and all database objects as fixed external dependencies.

This feature owns the backend, frontend, and test changes that were explicitly excluded from Feature 04.

## 2. Scope Boundary

Allowed implementation areas:

- Go DTOs, repositories, services, and HTTP APIs under `backend/`
- Vue/TypeScript API contracts and consumers under `frontend/`
- Go unit and API tests needed to verify application behavior

Forbidden implementation areas:

- `db/init.sql`
- `db/seed.sql`
- `db/seed-demo.sql`
- every file under `db/migrations/`
- any direct creation or modification of a live database object

No SQL file may be created, edited, renamed, or deleted by this feature.

## 3. Database Object Protection

The application must adapt to the database as defined; the database must not be adapted to the application.

Do not create, alter, replace, rename, or drop any table, column, constraint, index, sequence, view, function, procedure, trigger, type, extension, role, or permission.

Existing procedures may be invoked exactly as defined, including:

- `public.sp_task_to_history`
- `public.sp_requirement_to_history`
- `public.sp_task_requirement_recalculate`

If an existing object appears incorrect or prevents the application change, report the blocker and request a separate explicitly scoped database feature. Do not fix, replace, or work around the object by mutating the database.

## 4. Refactored Application Contract

### Identifiers and Removed Fields

- Adapt repositories to the current identity-based `bigint` IDs.
- Preserve string IDs at the JSON boundary if required for frontend compatibility, but validate that incoming IDs are positive integers before querying PostgreSQL.
- Remove database reads and writes for columns that no longer exist, including `task.complete` and `task.depth`.

### Task Completeness

- Read `task.done_req` and `task.total_req` from the database.
- Return both counters in task responses.
- Derive the existing API `complete` percentage without persisting a separate percentage:

```text
complete = 0                              when total_req = 0
complete = done_req / total_req * 100     otherwise
```

- After a requirement create, update, toggle, or delete, call the existing `public.sp_task_requirement_recalculate(task_id)` in the same transaction.
- Return the recalculated task and current requirement list in requirement mutation responses.

### Versions and Conflicts

- Return `version` for every task and requirement.
- Require the current version for task update, phase change, and delete requests.
- Require the current version for requirement update, toggle, and delete requests.
- Guard mutations with both `id` and expected `version`.
- Increment the active row version atomically on successful updates.
- Treat a stale version as a conflict and return HTTP 409.
- Treat a missing or negative version as invalid input and return HTTP 400.

### History and Transactions

- Use the existing history procedures rather than recreating their SQL behavior when they apply to a single-row mutation.
- Archive the current row before an update or delete.
- Keep history capture, guarded mutation, requirement recalculation, and response reads in one transaction.
- For project or task-tree deletion, archive all affected tasks and requirements according to the existing history-table contract before deleting active rows.
- Roll back the entire operation on a stale version or any database failure.

## 5. Backend Work

- Update task and requirement DTOs for `version`, `done_req`, and `total_req`.
- Remove the obsolete `depth` DTO field and obsolete stored-completeness assumptions.
- Update task and requirement scans and queries for the current column order and types.
- Replace UUID-specific handling and casts with `bigint`-compatible handling.
- Add optimistic concurrency checks to task and requirement mutations.
- Map stale mutations to HTTP 409 responses.
- Validate numeric identifiers before sending queries to PostgreSQL.
- Keep cascade archive/delete operations transactional and lock affected active rows to prevent inconsistent history during concurrent mutations.
- Do not add an automatic migration runner or schema bootstrap behavior.

## 6. Frontend Work

- Add `version`, `done_req`, and `total_req` to frontend task types.
- Add `version` to frontend requirement types.
- Remove the obsolete `depth` field.
- Send the latest task version with update, phase-change, and delete requests.
- Send the latest requirement version with update, toggle, and delete requests.
- Replace local task and requirement state with mutation responses so subsequent actions use the new version.
- Surface backend conflict messages through the existing error UI.

## 7. Tests

Update or add application tests that verify:

- bigint identifiers are accepted and malformed identifiers return HTTP 400.
- task responses derive `complete` from `done_req` and `total_req`.
- requirement changes propagate counters through multiple task ancestors by invoking the existing procedure.
- successful task and requirement updates increment their versions.
- stale task and requirement mutations return HTTP 409 and do not change active or history data.
- missing versions return HTTP 400.
- task-tree and project deletion archive affected rows before removal.
- frontend types and calls include the latest versions.

Tests must use an existing test database contract. They must not create, alter, replace, or drop database objects and must not modify SQL files as test setup.

## 8. Out of Scope

- Any database migration or schema correction.
- Any modification to an existing procedure or function.
- Reference-data or demo-data changes.
- Editing Feature 04 or earlier feature documents during implementation.
- New product behavior unrelated to compatibility with the refactored database.

## 9. Acceptance Criteria

- All application compatibility changes are contained in Feature 05 rather than Feature 04.
- No SQL file or live database object is changed.
- Backend queries use current bigint IDs, counters, versions, and history shapes without referencing removed columns.
- Task and requirement writes are version-guarded and stale writes return HTTP 409.
- Requirement mutations call the existing recalculation procedure transactionally.
- Frontend mutations send current versions and retain returned versions.
- Unit and API tests cover derived completeness, ancestor propagation, history, and stale-write conflicts.
- Existing backend and frontend checks pass against the already-refactored database contract.
