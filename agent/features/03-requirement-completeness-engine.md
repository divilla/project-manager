# Feature 03: Requirement Completeness Engine

## 1. Purpose
Implement the core product philosophy: completeness is measured from concrete, binary requirements instead of time estimates, story points, or manually entered progress percentages.

Each Definition of Done item is represented as a `requirement`.

## 2. Prototype Scope
- Add, edit, order, complete, uncomplete, and delete requirements for a task.
- Recalculate task completeness whenever requirements change.
- Return updated completeness values to the frontend immediately after changes.
- Support dashboard aggregation by phase and project.
- Preserve requirement history before any requirement update, toggle, or delete.

## 3. Requirement Rules
A good requirement is:

- Binary: complete or incomplete.
- Verifiable: the developer can prove whether it is done.
- Concrete: it names the artifact, behavior, test, or decision required.
- Small enough to be checked independently.

Examples:

- Good: "Add repository query using the existing project table and columns."
- Good: "Add API test for creating a task with the database-provided default phase."
- Weak: "Finish backend."
- Weak: "Make UI better."

## 4. Completeness Calculation
For a task with one or more requirements:

```text
completeness = completed_requirements / total_requirements * 100
```

Rules:

- A task with no requirements is 0% complete unless the existing database phase semantics explicitly mark it complete.
- If a database-provided completion phase exists, a requirement-less task in that phase may be treated as 100% complete.
- Requirement changes should recalculate the parent task inside the same backend operation.
- Completeness should be stored as a cached task field for fast board and dashboard reads.

## 5. Core User Flows

### Toggle Requirement
1. User checks or unchecks a requirement.
2. Frontend sends the updated completion state.
3. Backend starts a transaction and copies the current requirement row to `requirement_history` with `deleted = false`.
4. Backend updates the requirement and modified timestamp.
5. Backend copies the current parent task row to `task_history` with `deleted = false` if recalculated completeness will change the task row.
6. Backend recalculates parent task completeness and updates the task modified timestamp.
7. Frontend updates the task progress display.

### Add Requirement
1. User adds requirement text to a task.
2. Backend stores it with the next `order_index`.
3. Parent task completeness is recalculated.
4. Requirement appears in the task detail view.

## 6. API Notes
Expected conceptual endpoints:

- `POST /api/requirement/list`
- `POST /api/requirement/create`
- `POST /api/requirement/update`
- `POST /api/requirement/delete`

Every create, update, or delete operation should return enough data for the frontend to refresh the task completeness without performing unnecessary extra requests.

Every update or delete of an existing requirement must write the current row to `requirement_history` before changing the active row. Deletes must write `deleted = true`.

## 7. Acceptance Criteria
- Requirements can be created, edited, completed, uncompleted, and deleted.
- Task completeness updates after each requirement mutation.
- Task completeness never depends on free-form user-entered percentages.
- Requirement text is required and cannot be blank.
- Requirement completion changes are reflected in the task board and dashboard data.
- Implementation uses existing database tables/columns only and introduces no migrations.
- Requirement update/toggle/delete operations preserve the previous current row in `requirement_history`.
- Requirement deletes write history with `deleted = true`.
