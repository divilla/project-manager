# Feature 02: Projects & Tasks

## 1. Purpose
Allow the developer to create project workspaces and manage the task board that represents active product work. This is the main operational surface for the prototype.

## 2. Prototype Scope
- Create, list, view, update, and delete projects.
- Create, list, view, update, phase-change, and delete tasks.
- Display tasks grouped by `task_phase`.
- Show each task's calculated completeness percentage.
- Open a task detail view that includes its description, phase, and requirements.
- Preserve task history before any task update or delete.

## 3. Workflow Phases and Types
The prototype uses the existing database-provided task phases and task types.

- Load phases from the existing `task_phase` table.
- Load types from the existing `task_type` table.
- Do not hardcode a replacement list of valid options.
- Do not insert, update, or delete phase/type reference rows.
- Preserve the existing database identifiers, labels, and ordering.

## 4. Core User Flows

### Create a Project
1. User opens Projects.
2. User clicks create project.
3. User enters name and optional description.
4. System creates the project and shows it in the project list.

### Create a Task
1. User selects a project.
2. User creates a task with title and optional description.
3. System places the task in a valid default phase/type according to the existing database rules.
4. User can open the task and add requirements.

### Move a Task Between Phases
1. User opens the task or uses a board control.
2. User selects a new phase.
3. System updates the task to reference the selected existing `task_phase` option.
4. Dashboard aggregates update on the next refresh or response payload.

## 5. API Notes
Expected conceptual endpoints:

- `POST /api/project/list`
- `POST /api/project/get`
- `POST /api/project/create`
- `POST /api/project/update`
- `POST /api/project/delete`
- `POST /api/task/list`
- `POST /api/task/get`
- `POST /api/task/create`
- `POST /api/task/update`
- `POST /api/task/phase`
- `POST /api/task/delete`

Mutating operations should stay POST-based for prototype consistency.

## 6. Data Notes
Project records store:

- `id`
- `name`
- `description`
- `status`
- timestamps

Task records store:

- `id`
- `project_id`
- `title`
- `description`
- phase reference using the existing database field/relationship
- type reference using the existing database field/relationship
- `completeness`
- created and modified timestamps

Task history records:

- are written to `task_history`
- capture the current task row before an update or delete
- use `deleted = false` for updates and phase changes
- use `deleted = true` for deletes
- are written in the same transaction as the active-row change

## 7. Acceptance Criteria
- Projects can be created, listed, edited, and deleted.
- Tasks can be created, listed by project, edited, moved between phases, and deleted.
- A task defaults according to the existing database rules and starts with 0% completeness unless existing schema behavior says otherwise.
- Task cards show title, phase, and completeness.
- Deleting a project removes its tasks and requirements.
- `task_phase` and `task_type` values are loaded from the database and never modified by the application.
- Updating, phase-changing, or deleting a task writes its previous current version to `task_history` first.
- Deleting a task archives affected child requirements to `requirement_history` with `deleted = true` before removal.
- Deleting a project archives affected tasks and requirements to their history tables with `deleted = true` before removal.
- Task modified timestamps change when active task rows are updated.
