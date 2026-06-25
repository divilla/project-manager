# Feature 07: Project Management and Current Project Selection

## 1. Purpose

Make projects a first-class application concept instead of hiding selection inside task-board behavior.

Users should be able to create, rename, inspect, select, and safely retire projects. A project is the context for dashboards, planning, task creation, and requirement management, so the top menu needs a durable current-project selector on its right side.

The critical product rule is safety: projects that contain tasks must not be deleted. Deleting a project must never cascade-delete tasks or requirements.

## 2. Scope

Allowed implementation areas:

- project backend service, repository, DTOs, and API tests
- project API contract documentation
- `frontend/src/features/projects/`
- shared project-selection state in Pinia
- top-menu right-side current-project selector
- Projects route for project CRUD only
- Tasks route for task-board workflows
- Home, Planning, and task-board flows that need the current project ID
- frontend tests for project CRUD and current-project selection

Out of scope:

- task CRUD redesign beyond moving the existing task board to `/tasks`
- dashboard metrics beyond consuming the current project ID
- planning copilot implementation beyond receiving the current project ID
- bulk project archival, export, or import
- database schema changes unless required to enforce safe deletion

## 3. Backend Contract

Project deletion must be guarded by the backend, not only by frontend UI state.

Current behavior deletes project requirements and tasks before deleting the project. Replace that behavior with a safety check:

- `POST /api/v1/project/delete` must return a client error if any task exists for the project.
- The project row must remain unchanged when tasks exist.
- No task or requirement history procedures should run as part of a blocked project delete.
- Deleting a project with no tasks may delete only the project row.
- Deleting a missing project should continue to return not found.

Recommended response for a project with tasks:

```json
{
  "message": "project has tasks and cannot be deleted"
}
```

Use an explicit domain error such as `ErrProjectHasTasks` and map it to HTTP `409 Conflict`. If the existing API error style strongly prefers `400 Bad Request`, document the choice and keep the message specific enough for the frontend to render.

Backend tests must cover:

- deleting an empty project succeeds
- deleting a project with tasks fails
- failed deletion leaves the project, tasks, and requirements intact
- deleting a missing project still returns not found

## 4. Projects Page

The Projects page is the CRUD surface for project records. It must not contain the task board.

Required behavior:

- list all projects
- create a project by name
- rename an existing project
- delete only projects with no tasks
- confirm deletion for deletable projects before calling the delete endpoint
- show a clear blocked-deletion message when tasks exist
- refresh the project list after create, rename, and delete
- keep current-project selection valid after changes

Task-board behavior belongs on `/tasks`.

The page should make the safe-deletion rule visible. A project with tasks can show a disabled delete action or a tooltip that explains the task dependency. The backend remains authoritative even if the frontend cannot determine task usage locally.

Project deletion confirmation:

- Deletable projects open a persistent Quasar confirmation dialog before deletion.
- Project deletion uses `frontend/src/shared/ui/DeleteConfirmationDialog.vue`.
- Clicking the modal surrounding area must not close the dialog.
- The dialog title is `Are you sure?`.
- The dialog follows the app-wide confirmation button rule: `Cancel` is flat with no explicit color, and `OK` is not flat.
- Because project deletion is destructive, `OK` uses `color="negative"`.

Project list/get/create/update responses use the canonical `public.vw_project` shape:

```json
{
  "id": 1,
  "name": "Example project",
  "created": "2026-06-23T00:00:00Z",
  "modified": "2026-06-23T00:00:00Z",
  "task_count": 3
}
```

The view owns project ordering and includes `task_count`, so the frontend should not apply its own name-based reordering after create or rename. The count lets the UI communicate the safe-deletion rule before the user clicks delete.

## 5. Current Project Selector

Add a current-project selector to the right side of the top menu.

Behavior:

- load projects once the application shell starts
- display project names in a compact dropdown/select
- persist the selected project ID in local storage through Pinia
- expose the current project ID to pages and feature workflows as `currentProjectId`
- expose the derived current project as `currentProject`
- update the current project when the user selects a different project
- reload the current project if the persisted ID no longer exists
- never replace the selector with an ambiguous create-project button; if no projects exist, show the selector disabled and keep project creation on the Projects page

Selection rules:

1. If a persisted project ID exists and still belongs to a listed project, select it.
2. If no valid persisted project ID exists and projects are available, select the project with the lowest ID.
3. If no projects exist, clear the current project and route the user to the Projects page.
4. If the current project is renamed, keep it selected.
5. If the current project is deleted, select the remaining project with the lowest ID.
6. If the deleted project was the last project, clear selection and keep the user on the Projects page.

Use ID ordering for default selection because it is stable and independent of display sorting.

Project switching flow:

1. When the user changes the project selector, set `isSwitchingProject`.
2. Disable the selector while `isSwitchingProject` is true.
3. Redirect to `/loading`.
4. Refresh shared project-scoped data from the backend. Today this means projects and all tasks for `currentProjectId`; future project-scoped caches should be refreshed through the same flow.
5. Redirect to the current topic index. Examples: `/tasks/:id`, `/tasks/create/:parentId`, and `/tasks/edit/:taskId` return to `/tasks`.
6. Clear `isSwitchingProject` after the destination route is reached.

## 6. State Model

Use Pinia for durable project selection state.

Suggested store responsibilities:

- project list loading
- current project ID persistence
- current project derived value
- selection validation
- project switching state
- create, rename, and delete actions if they need to update shared state

Local form state belongs in components or composables. The store should not become a broad task-board store.

Suggested shape:

```text
frontend/src/features/projects/
  api/projectApi.ts
  model/project.types.ts
  model/projectSelection.store.ts
  composables/useProjectsCrud.ts
  components/ProjectSelector.vue
  components/ProjectList.vue
  components/ProjectCreateForm.vue
  components/ProjectRenameDialog.vue
```

The exact file names can follow existing conventions, but the current-project state should be feature-owned and reusable from the app shell, Home, Planning, and Projects workflows.

## 7. Integration Rules

Pages and workflows that operate inside a project must consume the current project ID instead of each owning a separate project selector.

Home:

- use current project ID for dashboard loading
- show an empty state when no project exists

Planning:

- use current project ID when committing generated tasks
- block commit actions until a project exists

Projects:

- provide project CRUD only
- keep the top-menu current project selector valid after project CRUD changes

Tasks:

- use the current project ID to choose which task board context is shown
- include the existing task create, board, detail, and requirement workflows

Task creation:

- send the current project ID as `project_id`
- block task creation when no current project exists

## 8. User Experience

The project selector should be operational UI, not a marketing element.

Expected UI behavior:

- compact select in the right side of the top menu
- clear distinction between "no projects exist" and "project list failed to load"
- clear empty state that sends users to create their first project
- loading and error states that do not block unrelated navigation
- no silent fallback to the wrong project after a persisted ID becomes invalid

Project deletion should be deliberate:

- use a persistent confirmation dialog for deletable projects
- explain blocked deletion when tasks exist
- do not offer destructive cascade deletion in this feature

## 9. Migration Steps

1. Change backend project deletion so projects with tasks cannot be deleted.
2. Add backend tests for blocked project deletion and intact dependent data.
3. Extend project list responses with task counts if that is chosen for the UI.
4. Add or update project API client types.
5. Add a Pinia project-selection store persisted to local storage.
6. Add the top-menu right-side project selector.
7. Rewrite the Projects route as the project CRUD surface.
8. Move the existing task-board surface to `/tasks` and add a Tasks navigation item.
9. Update Home, Planning, and task-board flows to consume the current project ID.
10. Add frontend tests for default selection, persisted selection, invalid persisted IDs, and blocked deletion messaging.

Each step should keep the existing app buildable and should not remove the current ability to create and manage tasks inside a selected project.

## 10. Acceptance Criteria

- Projects can be listed, created, renamed, and deleted from the Projects page.
- The task board is available from `/tasks`, not embedded in `/projects`.
- Projects with one or more tasks cannot be deleted.
- Backend tests prove blocked deletion preserves project, task, and requirement data.
- Empty projects can be deleted.
- The top menu includes a right-side current-project selector.
- Current project selection persists across reloads.
- Invalid persisted project IDs are repaired deterministically.
- When no projects exist, the user is directed to the Projects page.
- Task creation and other project-scoped workflows use the current project ID.
- Frontend tests cover the current-project selection rules.
