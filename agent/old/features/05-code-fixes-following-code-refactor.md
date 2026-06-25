# Feature 05: Code Fixes Following Database Refactor

## 1. Purpose

Update the backend, frontend contracts, and application tests to use the already-refactored PostgreSQL database correctly.

The current live database and `db/init.sql` are a synchronized, authoritative baseline supplied for this feature. The application must adapt to that contract.

Database evolution and migration delivery remain owned separately by the database owner. Migrations are intentionally deferred and are not part of this feature.

## 2. Scope Boundary

Allowed implementation areas:

- Go DTOs, repositories, services, and HTTP APIs under `backend/`
- Vue/TypeScript API contracts and consumers under `frontend/`
- Existing and new Go API tests
- Documentation needed to describe the authoritative database contract

Forbidden implementation areas:

- `db/seed.sql`
- `db/seed-demo.sql`
- every file under `db/migrations/`
- independent application-driven creation or modification of a live database object

The `db/init.sql` state supplied with this feature describes the database contract; it is not a migration mechanism. Application work must not introduce additional database changes. If an application change needs a different database contract, report the dependency to the database owner rather than changing SQL or live objects.

## 3. Authoritative Database Contract

The database owner is responsible for database objects and for keeping the live database synchronized with `db/init.sql`. Application implementation must not create, alter, replace, rename, or drop database objects.

The following existing objects must be used exactly as currently defined:

- `public.vw_task`
- `public.fn_task_descendants`
- `public.sp_task_to_history`
- `public.sp_task_phase_recalculate`
- `public.sp_requirement_to_history`
- `public.sp_task_requirement_recalculate`

The application may rely on the current signatures and behavior of these objects. If an object appears incorrect or prevents an application change or test from passing, report the blocker to the database owner. Do not independently fix, replace, recreate, or work around it by mutating the database.

In the authoritative contract, `public.sp_task_phase_recalculate(_parent_id bigint)` accepts the immediate affected parent ID, not the mutated task ID. A null parent is valid for a root task and produces no parent-phase update. Callers must capture the relevant parent ID before operations that can remove or re-parent a task.

Use `public.fn_task_descendants(_task_id bigint, _descendants bigint[])` when task mutation logic needs the complete descendant set for a task. Call it with an empty bigint accumulator, for example `select public.fn_task_descendants(6, ARRAY[]::bigint[])`. The returned array contains descendant task IDs only; it does not include the root `_task_id`.

## 4. Data Type Audit

Before changing behavior, examine every DTO, repository scan destination, request type, response type, helper, and frontend interface affected by the database refactor. Apply these mappings consistently throughout the codebase:

| PostgreSQL type | Go type |
| --- | --- |
| `smallint` | `int16` |
| `bigint` | `int` |
| `integer` | `int` |
| `boolean` | `bool` |
| `text` | `string` |
| `timestamp with time zone` | `time.Time` |
| nullable `bigint` | `*int` or an equivalent nullable scan type converted to `*int` |

Specific fields requiring review include:

- `project.id`
- `task.id`, `task.parent_id`, and `task.project_id`
- `task.version`, `task.difficulty`, `task.priority`, `task.done_req`, and `task.total_req`
- `requirement.id` and `requirement.task_id`
- `requirement.version`
- history-table identifiers and versions
- `task_phase.priority` and `task_type.priority`
- `vw_task.completed`, which is a PostgreSQL `smallint` and therefore uses Go `int16`
- all request DTOs that carry an identifier

Remove remaining UUID-specific types, parsing, casts, fixtures, and string-ID assumptions. JSON request and response contracts must use numeric identifiers consistently with their Go DTOs.

Do not change the database to accommodate a mismatched Go type.

## 5. Task Read Contract

Task completion is already calculated by the database view.

- Select tasks from `public.vw_task`, not directly from `public.task`, whenever a task response requires completion.
- Read the pre-calculated `completed` column from `vw_task` into the task DTO.
- Expose the field as `completed` in JSON and frontend types.
- Do not calculate completion in Go or TypeScript.
- Do not read or write a `completed` column on `public.task`.
- Keep `done_req` and `total_req` mapped using their PostgreSQL `smallint` type where those counters are part of the task response.
- After a task mutation, re-select the task from `vw_task` inside the transaction when the response requires the current `completed` value.

## 6. Versioning and History

Both `task.version` and `requirement.version` are PostgreSQL `smallint` values and must use Go `int16`.

Rules:

- Omit `version` from inserts and use the database default.
- Expose `version` in task and requirement responses as informative state for the frontend.
- Do not accept `version` in create, update, phase-change, parent-change, or delete requests.
- Do not use `version` in `WHERE` clauses or for optimistic-concurrency conflicts.
- Increment `task.version` only when an update actually changes at least one task history-bearing field: `task_type`, `name`, `description`, or `parent_id`.
- Increment `requirement.version` only when an update actually changes the history-bearing field `definition`.
- Call the matching history procedure before the update that increments the version, so the procedure archives the current row and current version.
- Updates that change only non-history fields must not increment `version` and must not call a history procedure.
- History capture, the mutation, required recalculation procedures, and response reads must remain in one transaction.

Conceptual task history-bearing update shape:

```sql
update public.task
set
    task_type = :task_type,
    name = :name,
    description = :description,
    parent_id = :parent_id,
    version = version + 1
where id = :id;
```

Conceptual requirement history-bearing update shape:

```sql
update public.requirement
set
    definition = :definition,
    version = version + 1
where id = :id;
```

Conceptual non-history update shape:

```sql
update public.requirement
set done = :done
where id = :id;
```

The application must never assign a caller-provided version or increment a version for a non-history change.

## 7. Task Mutation Rules

Every task insert, update, and delete must run in an explicit transaction.

### Insert Task

1. Begin a transaction.
2. Insert the task without explicitly writing `version`.
3. Call `public.sp_task_phase_recalculate(parent_id)` after the insert, using the new task's immediate parent ID.
4. Re-select the task from `public.vw_task` if a task response is returned.
5. Commit the transaction.

Task history is not written for a newly inserted row.

### Update Task Details

`POST /api/v1/task/update` changes only `task_type`, `name`, and `description`.

Before updating, determine whether any included history-bearing field will actually change.

History-bearing fields are:

- `task_type`
- `name`
- `description`
- `parent_id`

Required order:

1. Begin a transaction and read the current task.
2. Determine whether at least one history-bearing field actually changes value.
3. If a history-bearing field changes, call `public.sp_task_to_history(task_id, false)` before the update.
4. Apply the business-field update and `version = version + 1` in the same statement only when a history-bearing field changes. Otherwise, apply the non-history update without changing `version`.
5. Call `public.sp_task_phase_recalculate(parent_id)` after the update, even when the update changes only a field not stored in task history. If `parent_id` changed, recalculate each distinct old and new parent.
6. Re-select the task from `public.vw_task`.
7. Commit the transaction.

Do not call `sp_task_to_history` when none of these fields actually changes. Call `sp_task_phase_recalculate` after every successful task update.

### Update Task Difficulty

`POST /api/v1/task/update-difficulty` changes only `difficulty`. It does not call `sp_task_to_history` or increment `version`. It calls `sp_task_phase_recalculate(parent_id)` and returns the refreshed task.

### Update Task Priority

`POST /api/v1/task/update-priority` changes only `priority`. It does not call `sp_task_to_history` or increment `version`. It calls `sp_task_phase_recalculate(parent_id)` and returns the refreshed task.

### Update Task Parent

`POST /api/v1/task/update-parent` is the only endpoint that changes `task.parent_id`.

The request contains:

- `id`: the task ID
- `parent_id`: the new parent ID, or `null` to make the task root-level

The request does not contain `version`. The `parent_id` field uses `*int` without `omitempty`.

Required order:

1. Begin a transaction and read the current task.
2. Validate a non-null parent and determine whether `parent_id` actually changes.
3. If the parent changes, call `public.sp_task_to_history(task_id, false)`.
4. Update `parent_id` and increment `version` with `version = version + 1` in the same statement.
5. Recalculate each distinct old and new parent with `public.sp_task_phase_recalculate(parent_id)`.
6. Re-select the task from `public.vw_task`.
7. Commit the transaction.

If the requested parent equals the current parent, do not write history or increment `version`.

When the requested parent is non-null, reject it if it appears in `public.fn_task_descendants(id, ARRAY[]::bigint[])`; a task cannot be re-parented under one of its own descendants.

### Update Task Phase

`POST /api/v1/task/update-phase` changes only `task_phase`. It does not call `sp_task_to_history` or increment `version`. It calls `sp_task_phase_recalculate(parent_id)` and returns the refreshed task.

### Delete Task

Required order for every deleted task row, including rows deleted through a task-tree or project operation:

1. Begin or join the surrounding delete transaction.
2. Read the task and capture its current state.
3. Call `public.sp_task_to_history(task_id, true)` before deleting the row.
4. Delete the task.
5. Call `public.sp_task_phase_recalculate(parent_id)` after the delete using the parent ID captured before deletion.
6. Commit only after history, delete, and recalculation all succeed.

The application must not duplicate or replace the database-owned stored procedure definitions.

## 8. Requirement Mutation Rules

Every requirement insert, update, and delete must run in an explicit transaction. Capture the relevant `task_id` before any operation that may remove or change it.

### Insert Requirement

1. Begin a transaction.
2. Insert the requirement without explicitly writing `version`.
3. Call `public.sp_task_requirement_recalculate(task_id)` after the insert.
4. Read the recalculated task from `public.vw_task` and read the current requirement list if required by the API response.
5. Commit the transaction.

Requirement history is not written for a newly inserted row.

### Update Requirement Definition

`POST /api/v1/requirement/update` changes only `definition`.

The history-bearing field is:

- `definition`

Required order:

1. Begin a transaction and read the current requirement.
2. Determine whether `definition` actually changes value.
3. If `definition` changes, call `public.sp_requirement_to_history(requirement_id, false)` before the update.
4. Apply the definition update and `version = version + 1` in the same statement.
5. Re-read the task from `public.vw_task` and the current requirement list.
6. Commit the transaction.

If `definition` is unchanged, do not write history or increment `version`.

### Update Requirement Done

`POST /api/v1/requirement/update-done` changes only `done`. It does not call `sp_requirement_to_history` or increment `version`. When `done` actually changes, call `public.sp_task_requirement_recalculate(task_id)` and return the recalculated task and current requirements.

### Update Requirement Task

`POST /api/v1/requirement/update-task` changes only `task_id`. It does not call `sp_requirement_to_history` or increment `version`. When `task_id` actually changes, recalculate both the previous and new task IDs, then return the new task and its current requirements.

### Delete Requirement

1. Begin a transaction and read the requirement.
2. Capture `task_id`.
3. Call `public.sp_requirement_to_history(requirement_id, true)` before deleting the row.
4. Delete the requirement by ID.
5. Call `public.sp_task_requirement_recalculate(task_id)` after the delete using the captured task ID.
6. Re-read the recalculated task from `public.vw_task` and the current requirement list.
7. Commit the transaction.

## 9. DTO, API, and Frontend Work

- Apply the data-type mappings from Section 4 to every affected DTO and scan target.
- Keep `version` in task and requirement response DTOs as `int16`, but remove it from request DTOs.
- Ensure task DTOs expose `completed` selected from `vw_task`.
- Remove obsolete task completion calculations and obsolete fields that are no longer returned by the database contract.
- Keep request and response field names consistent across Go DTOs, API tests, TypeScript interfaces, and Vue callers.
- Do not send versions from frontend mutation requests.
- Replace frontend state with successful mutation responses so displayed version information remains current.
- Remove `parent_id` from the standard task update request and expose parent changes only through `POST /api/v1/task/update-parent`.
- Use `POST /api/v1/task/update-difficulty`, `update-priority`, and `update-phase` for their respective task fields.
- Split requirement mutations across `POST /api/v1/requirement/update`, `update-done`, and `update-task`.

## 10. API Test Order

Testing must be completed in this order:

1. Update all current API tests for the refactored identifiers, `smallint` versions and counters, `vw_task.completed`, request fields, and response fields.
2. Run the complete existing API-test suite and make it pass before adding new coverage.
3. Add focused API tests for every rule below.
4. Run the complete API-test suite again.

Required focused API-test coverage:

- PostgreSQL `smallint` fields scan into and round-trip through Go `int16` fields.
- PostgreSQL `bigint` identifiers scan into and round-trip through Go `int` fields.
- Task responses read `completed` from `vw_task`.
- Task insertion calls `sp_task_phase_recalculate` after the insert and commits the resulting phase aggregation.
- A task update that changes `task_type`, `name`, `description`, or `parent_id` archives the previous row before updating.
- A task update that changes no history-bearing field does not create task history.
- A task update increments `version` only when at least one history-bearing field actually changes.
- Task phase, difficulty, or priority-only updates leave `version` unchanged and do not create task history.
- `POST /api/v1/task/update-parent` changes or clears `parent_id`, archives the old task, and increments `version` only when the parent actually changes.
- Task deletion archives the row before deletion and calls `sp_task_phase_recalculate` after deletion.
- Requirement insertion calls `sp_task_requirement_recalculate` after the insert.
- A requirement definition change archives the previous row before updating.
- A `done`-only requirement update does not create requirement history.
- A requirement definition change increments `version`; a `done`-only update leaves `version` unchanged.
- Requirement deletion archives the row before deletion and calls `sp_task_requirement_recalculate` afterward with the captured `task_id`.
- Requirement insert, update, and delete responses contain the task state read from `vw_task` after recalculation.
- A forced failure proves each history/mutation/recalculation sequence rolls back as one transaction.
- Task-tree and project delete paths apply the same per-row history and recalculation rules.

Tests must use the authoritative database contract. Test setup must not create, alter, replace, or drop database objects and must not modify SQL files.

## 11. Out of Scope

- Implementing or deploying database migrations.
- Application-owned modification of `vw_task` or an existing procedure/function.
- Reference-data or demo-data changes.
- Editing Feature 04 or earlier feature documents during implementation.
- Product behavior unrelated to compatibility with the refactored database.

## 12. Acceptance Criteria

- All affected PostgreSQL `smallint` values use Go `int16`, and all affected PostgreSQL `bigint` values use Go `int` consistently throughout DTOs and code paths.
- Task completion is selected as `completed` from `public.vw_task`; it is not calculated in application code.
- Task history is written before task deletion and before updates that change `task_type`, `name`, `description`, or `parent_id`.
- Task updates increment `version` with `version = version + 1` only when `task_type`, `name`, `description`, or `parent_id` actually changes.
- Every task insert, update, and delete calls `sp_task_phase_recalculate` after the mutation in the same transaction.
- Requirement history is written before requirement deletion and before updates that change `definition`.
- Requirement updates increment `version` with `version = version + 1` only when `definition` actually changes.
- Requirement mutations call `sp_task_requirement_recalculate` in the same transaction whenever `done`, `task_id`, insertion, or deletion changes task aggregates.
- Existing API tests are fixed and passing before new API tests are added.
- Focused API tests cover every data-type, view, history, versioning, procedure-ordering, and transaction rule in this feature.
- Application implementation introduces no database change beyond the authoritative baseline supplied by the database owner.
