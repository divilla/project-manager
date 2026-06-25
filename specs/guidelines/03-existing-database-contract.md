# Spec 03: Existing Database Contract

## 1. Authoritative Rule
The project already has a PostgreSQL database. The existing database is the source of truth and must be used exactly as it is.

Agents and implementation work must not:

- create migrations
- create, rename, or drop tables
- create, rename, or drop columns
- change constraints, indexes, triggers, or reference data
- seed or modify lookup/reference tables
- invent replacement table names when an existing table already provides the data

All backend code must inspect and use the current database contract as-is.

## 2. Reference Data
The `task_phase` and `task_type` tables are already filled with all possible options.

These tables are fixed reference data for the prototype:

- Do not insert new rows into `task_phase`.
- Do not update or delete rows from `task_phase`.
- Do not insert new rows into `task_type`.
- Do not update or delete rows from `task_type`.
- Do not hardcode a separate application-only list of valid phases or task types.

The UI, API validation, dashboard grouping, and AI planning workflows must read valid phase and type options from the existing database.

## 3. Application Data Usage
The application may create and update normal product data through supported API workflows, such as projects, tasks, and task requirements, but only using the existing tables and columns.

The current task and requirement records live in the existing `task` and `requirement` tables. These tables include created and modified timestamp fields. Application writes must preserve the intended timestamp semantics:

- create operations set the created and modified timestamps according to the existing database convention
- update operations refresh the modified timestamp on the current row
- delete operations must archive the current row before removing or otherwise deleting the active version

Before implementing a repository, endpoint, or AI-generated task write, the implementation must confirm:

- the exact existing table name
- the exact existing column names
- the correct primary key and foreign key fields
- whether phase/type fields reference `task_phase` and `task_type` by id, code, or another existing key
- any required fields, defaults, or constraints already present in the database

If the desired feature does not map cleanly to the existing schema, do not change the database. Update the application design or ask for a human database change decision.

## 4. History Tables
The database includes `task_history` and `requirement_history` tables. These tables preserve the previous current version of task and requirement records before destructive or mutating changes.

History rules:

- Before any user-initiated task update, copy the current `task` row into `task_history` with `deleted = false`.
- Before any AI-initiated task update, copy the current `task` row into `task_history` with `deleted = false`.
- Before any user-initiated task delete, copy the current `task` row into `task_history` with `deleted = true`.
- Before any AI-initiated task delete, copy the current `task` row into `task_history` with `deleted = true`.
- Before any user-initiated requirement update, copy the current `requirement` row into `requirement_history` with `deleted = false`.
- Before any AI-initiated requirement update, copy the current `requirement` row into `requirement_history` with `deleted = false`.
- Before any user-initiated requirement delete, copy the current `requirement` row into `requirement_history` with `deleted = true`.
- Before any AI-initiated requirement delete, copy the current `requirement` row into `requirement_history` with `deleted = true`.
- If deleting a project or task deletes child tasks or requirements, archive every affected current `task` and `requirement` row first with `deleted = true`.

The history insert and the active-row update/delete must happen in the same database transaction. If the history insert fails, the active row must not be changed.

History records should represent the row state before the change, not the new state after the change.

## 5. Completeness Calculations
Completeness is still derived from task requirements, but calculation must use the existing schema.

The conceptual calculation remains:

```text
completeness = completed_requirements / total_requirements * 100
```

Implementation details must adapt to the actual existing tables and columns. If the database already stores cached completeness values, use the existing field. If it does not, calculate completeness in queries or service logic without adding schema.

## 6. Task Hierarchy Helpers
The database provides `public.fn_task_descendants(_task_id bigint, _descendants bigint[])` for retrieving the complete descendant task ID set below a task.

Call it with an empty bigint accumulator:

```sql
select public.fn_task_descendants(6, ARRAY[]::bigint[]);
```

The returned array contains descendants only and excludes the root task ID. Use this helper before re-parenting a task to reject a requested parent that is already one of the task's descendants.

## 7. Phase and Type Handling
Task phases and task types are not free-form strings in product logic. They are existing database-managed options.

Application behavior:

- Load phase options from `task_phase`.
- Load type options from `task_type`.
- Store task phase/type using the existing relationship or column format.
- Display labels from the reference tables.
- Preserve the database ordering if an order/sort column exists.

AI behavior:

- Planning prompts may include the current phase/type options fetched from the database.
- AI output must be validated against those fetched options.
- Invalid AI phase/type values must be rejected or mapped by the user before saving.

## 8. Documentation Note
Any diagrams or examples in the documentation are conceptual only. They describe product behavior, not a proposed database schema. The actual database always wins.
