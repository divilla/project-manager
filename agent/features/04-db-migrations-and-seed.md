# Feature 04: Database Seed Scripts

## 1. Purpose

Provide repeatable SQL scripts for required reference data and realistic local demo data using the database exactly as it already exists.

This feature is limited to seed-data scripts. It does not authorize application-code changes or changes to database object definitions.

## 2. Allowed Deliverables

Implementation of this feature may create or update only these SQL files:

- `db/seed.sql`
- `db/seed-demo.sql`

No non-SQL implementation file may be changed by this feature. Any backend, frontend, DTO, API, repository, service, configuration, or test change belongs to Feature 05.

`db/init.sql` is the authoritative description of the database that has already been applied. It may be read to understand tables, columns, defaults, and existing procedures, but it must not be modified.

`db/migrations/` is the reserved location for future explicitly requested migrations. This feature does not request a schema migration, so the directory and its contents must remain unchanged.

## 3. Database Object Protection

Existing database objects are strictly read-only unless a future feature explicitly names the object and requests a specific change.

This feature must not create, alter, replace, rename, or drop any:

- database or schema
- table or column
- constraint or foreign key
- index or sequence
- view or materialized view
- function or procedure
- trigger, type, extension, role, or permission

This prohibition includes preventative fixes, cleanup, performance improvements, inferred corrections, and changes to `public.sp_task_requirement_recalculate`, `public.sp_task_to_history`, or `public.sp_requirement_to_history`.

Seed scripts may call an existing procedure exactly as currently defined. They must never recreate, replace, or refactor it. If an existing database object does not support the requested seed operation, stop and report the blocker; do not change the object under this feature.

## 4. `db/seed.sql`

`db/seed.sql` seeds only the required `task_phase` and `task_type` reference rows.

Requirements:

- Use only `INSERT`/data-conflict handling needed to make the script repeatable.
- Running the script more than once must not create duplicate rows.
- Do not delete reference values or invent additional phases or types.
- Do not create projects, tasks, requirements, or history rows.
- Do not contain DDL or database-object definitions.

Required `task_phase` rows:

| slug | priority |
| --- | ---: |
| `backlog` | 0 |
| `progress` | 1 |
| `review` | 2 |
| `staging` | 3 |
| `production` | 4 |
| `repair` | 5 |

Required `task_type` rows:

| slug | priority |
| --- | ---: |
| `epic` | 0 |
| `feature` | 0 |
| `group` | 0 |
| `issue` | 0 |
| `spike` | 0 |
| `task` | 0 |
| `upgrade` | 0 |

These values come from the existing database and must be preserved exactly.

## 5. `db/seed-demo.sql`

`db/seed-demo.sql` provides destructive local demo-data reset and seeding. It is not a production migration.

### Reset Rules

- Truncate application and history data from `requirement_history`, `task_history`, `requirement`, `task`, and `project`.
- Preserve all rows in `task_phase` and `task_type`.
- Do not alter any table, identity definition, sequence definition, constraint, or other database object.
- Do not insert history rows for newly seeded records.
- Run the reset and inserts transactionally so a failed seed does not leave a partial demo dataset.

### Projects

Create exactly three projects:

- `demo1`
- `demo2`
- `demo3`

`demo1` is the populated showcase project. `demo2` and `demo3` remain empty to exercise empty-project states.

### Tasks

Create exactly 60 tasks for `demo1`.

- The count includes parent and leaf tasks.
- Model a coherent software project with realistic names and descriptions.
- Include a meaningful multi-level hierarchy.
- Use only existing seeded phase and type slugs.
- Include varied phases, types, priorities, difficulties, and completion states.
- Every `parent_id` must reference another task in `demo1`.
- Cycles and cross-project parent relationships are invalid.
- Do not create tasks for `demo2` or `demo3`.

### Requirements

- Add between 1 and 9 requirements, inclusive, to every leaf task in `demo1`.
- Do not add direct requirements to non-leaf tasks.
- Requirement definitions must be concrete, binary, verifiable, and relevant to their task.
- Include completed and incomplete requirements so the dataset contains 0%, partial, and 100% completeness examples.

After inserting requirements, call the existing `public.sp_task_requirement_recalculate` procedure as needed to populate task counters. Do not manually maintain cached counters and do not define or modify the procedure.

## 6. Verification

Verification for this feature is limited to executing and querying the SQL scripts. It must not require application-code changes.

Verify that:

- `db/seed.sql` can run twice without duplicate reference rows.
- `db/seed-demo.sql` can run twice and produces the same required record counts.
- Exactly three projects exist after demo seeding.
- `demo1` contains exactly 60 tasks.
- `demo2` and `demo3` contain no tasks.
- Every leaf task has between 1 and 9 requirements.
- Non-leaf tasks have no direct requirements.
- Seeded task counters match the active requirements and descendant aggregates.
- Task and requirement history tables are empty immediately after seeding.
- `task_phase` and `task_type` retain the required values.

## 7. Acceptance Criteria

- Feature implementation changes only `db/seed.sql` and `db/seed-demo.sql`.
- No application, configuration, test, or non-SQL file is changed.
- `db/init.sql` and `db/migrations/` are unchanged.
- No database object is created, altered, replaced, renamed, or dropped.
- No existing procedure or function definition is changed.
- Reference seeding is repeatable and contains only the specified values.
- Demo seeding is repeatable and produces the specified projects, tasks, hierarchy, requirements, completion coverage, and empty history tables.
