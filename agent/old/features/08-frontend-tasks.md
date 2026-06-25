# Feature 08: Frontend Tasks

## Current State

The frontend currently provides a Kanban-style task view. Tasks are displayed as cards grouped by their workflow phase, giving users a board-oriented view of project work.

The Tasks page does not show a page title or descriptive header copy above the board. The top page action area is reserved for task search controls.

The top search form contains:

- blue New Task button with an add-task icon
- task name text input
- task type dropdown
- task phase dropdown
- blue Search button with a search icon
- red Clear button with a clear icon

The app top toolbar includes a dark mode switch. It uses Quasar dark mode so toggling it applies Quasar's `body--dark` class to the document body.

New Task behavior:

- New Task is always enabled when a project is selected.
- New Task opens the dedicated task creation page at `/tasks/create/0`.
- The task creation page owns its own task name, description, task type, task phase, difficulty, priority, and parent context.
- The project is not shown as a form field. Task creation always uses the globally selected current project.
- The creation payload must match `POST /api/v1/task/create`: `project_id`, `name`, optional `description`, `task_phase`, `task_type`, `difficulty`, `priority`, and `parent_id`.
- Creating a task must not read from, overwrite, or clear the search form fields.
- The create action remains clickable and uses field validation to explain missing required input.
- After successful creation, the app opens the created task detail route.

Task creation route:

- The Tasks feature has a task creation route at `/tasks/create/:parentId`.
- Root task creation takes `project_id` from `currentProjectId` when opened at `/tasks/create/0`.
- Child task creation takes `project_id` from the loaded parent task. This keeps child creation valid when a task detail route was opened directly from a pasted/bookmarked URL and the global project selector initially points at another project.
- When opened with `parentId` set to `0`, the page creates a root task.
- When opened with a positive `parentId`, the page loads that parent task, includes `parent_id` in the create payload, and displays the parent task in the first line of the page.
- The task detail page's Add Child Task action opens `/tasks/create/<current task id>`.
- Parent context must come from the backend task detail response, not only from the project task cache.
- Required field validation is currently limited to fields required by the backend contract. The Create action remains clickable so validation messages can be shown on submit.

Project switching guardrail:

- If the user changes the current project from any task route, the app redirects to `/loading`, refreshes project-scoped data, and then returns to `/tasks`.
- This applies to `/tasks/:id`, `/tasks/create/:parentId`, and `/tasks/edit/:taskId`.
- The project selector is disabled while the switch is in progress.
- Nested task routes must not keep stale task or parent context across a project change.
- If the user opens a task route directly and the loaded task belongs to a different project than the global selection, the task detail page must ask before switching project context.
- The mismatch prompt must name both projects and offer:
  - `Switch`, which programmatically selects the task's project and preserves the requested route after the `/loading` refresh flow
  - `Stay`, which keeps the current project and leaves the mismatched task route by going back, or by replacing with `/tasks` when browser history cannot go back
- Programmatic route-driven project switches must use explicit route intent state, not a loose global boolean, so the normal selector-change flow still collapses nested task routes to `/tasks`.

Task editing route:

- The Tasks feature has a task editing route at `/tasks/edit/:taskId`.
- The edit page uses the same form layout as task creation, but it omits difficulty and priority.
- The edit page loads the task from the backend, pre-fills task name, description, task type, and task phase, and saves through the existing task update endpoints.
- The edit page description field is a large markdown editing textarea with a 600px minimum height.
- Task detail rows use a single minimal-width action column after the `Version` column.
- The current task row and child task rows show a white three-dots menu icon in that action column.
- The task action menu contains `Edit` and `Delete` items.
- `Edit` opens `/tasks/edit/<task id>`.
- `Delete` is enabled only for leaf tasks. If a task has children, the delete menu item must always be disabled.
- Deleting from the task detail page opens a persistent Quasar confirmation dialog. Clicking the modal surrounding area must not close it.
- Task detail deletion uses `frontend/src/shared/ui/DeleteConfirmationDialog.vue`.
- The task detail delete confirmation dialog title is `Are you sure?`.
- The task detail delete confirmation follows the app-wide confirmation button rule: `Cancel` is flat with no explicit color, and `OK` is not flat.
- Because task deletion is destructive, `OK` uses `color="negative"`.

Task board delete behavior:

- Task cards keep their delete icon action.
- Clicking a task card delete action opens a persistent Quasar confirmation dialog before calling the delete endpoint.
- Task board deletion uses `frontend/src/shared/ui/DeleteConfirmationDialog.vue`.
- The task board delete confirmation dialog title is `Are you sure?`.
- The task board delete confirmation follows the app-wide confirmation button rule: `Cancel` is flat with no explicit color, and `OK` is not flat.
- Because task deletion is destructive, `OK` uses `color="negative"`.

Search behavior:

- Search refreshes the selected project's task list and applies the current search parameters.
- Empty search fields mean no filtering for that field.
- Name search matches task names case-insensitively.
- Type and phase search match the selected database-provided slug exactly.
- The board remains grouped by workflow phase after filtering.

Clear behavior:

- Clear resets task name, task type, and task phase search parameters.
- Clear refreshes the selected project's task list after resetting the parameters.

## Task Detail Route

The Tasks feature now has a nested task detail route at `/tasks/:id`.

The task detail page uses a table-oriented hierarchy view:

- Opening `/tasks/:id` directly must render the task returned by the backend task detail endpoint, even if the currently selected project cache points at another project.
- Rows above the current task show the parent chain.
- The parent chain is ordered from the root task with no parent down to the opened task.
- The opened task row is visually emphasized.
- Rows below the opened task show child tasks of the opened task.

For example, opening `/tasks/17` should show the ancestor path first, then task `17` as the current task, then task `17`'s child tasks below it.

The same table should also show requirements for the opened task. Requirements use the same table grid, but their columns map differently from task rows.

Add a requirement header row before requirement rows:

```html
<tr class="text-weight-bold">
  <td class="text-right">nr</td>
  <td class="text-center">&nbsp;</td>
  <td class="text-left" colspan="4">Requirement</td>
  <td class="text-center">Complete</td>
  <td class="text-center">Modified</td>
  <td class="text-center">Version</td>
  <td class="task-actions-cell"></td>
</tr>
```

Requirement row mapping:

- `nr`: requirement ID
- second column: intentionally blank spacer cell using `&nbsp;`
- `Requirement`: requirement definition, spanning the next four table columns with `colspan="4"`
- `Complete`: requirement done state displayed as a clickable checkbox wired to the backend
- `Modified`: requirement modified timestamp
- `Version`: requirement version
- final action column: a three-dots menu containing `Edit` and `Delete`
- The blank second column is required to keep requirement rows aligned with the task rows above. Do not remove it or collapse the requirement row to fewer cells.

Requirement editing:

- Requirement rows do not have a dedicated pencil icon column.
- The requirement three-dots action menu has an `Edit` item.
- The requirement `Edit` item opens a separate `Edit Requirement` popup dialog.
- The edit requirement definition field is a textarea with at least three visible lines.
- Saving the dialog calls the backend requirement update endpoint.
- The requirement list and visible version column are updated from the backend response.

Requirement deletion:

- The requirement three-dots action menu has a `Delete` item.
- Requirement deletion uses the same page-level `frontend/src/shared/ui/DeleteConfirmationDialog.vue` instance as task deletion.
- The requirement delete confirmation dialog title is `Are you sure?`.
- The requirement delete confirmation follows the app-wide confirmation button rule: `Cancel` is flat with no explicit color, and `OK` is not flat.
- Because requirement deletion is destructive, `OK` uses `color="negative"`.

Requirement creation:

- The Add Requirement popup title is `Add Requirement`.
- The add requirement definition field is a textarea with at least three visible lines.
- The submit button label is `Add`, not `Create`.

Task type display:

- Task type is displayed as a Quasar button in task rows.
- `epic` and `group` use purple.
- `issue` uses red.
- Other task types use teal.

Below the hierarchy/requirements table, the task detail page also shows a second task detail table for markdown descriptions:

- the opened task description is rendered from `currentTask.description_html`
- parent task descriptions are rendered from the backend batch endpoint `POST /api/v1/task/rendered-descriptions`
- the hierarchy table lists parents from root parent down toward the opened task
- the description table lists parent descriptions in the opposite direction, from the closest parent back toward the root parent
- each parent description row contains only the markdown description cell
- each markdown description is rendered inside an element with `class="apply-markdown"`
- description cells must wrap long text and links instead of stretching the table horizontally

Markdown rendering:

- Store and edit task descriptions as raw markdown.
- Prefer rendering markdown on the backend instead of parsing markdown in the Vue page.
- Backend markdown rendering is abstracted behind `backend/pkg/markdown`.
- The service layer receives an injected parser and sanitizer, then renders task descriptions as sanitized HTML.
- Current backend stack: `goldmark` with GFM extensions for markdown rendering, then `bluemonday` for HTML sanitization.
- Add a rendered/sanitized field named `description_html` to task API responses while keeping `description` as the raw markdown source.
- The batch rendered-description endpoint accepts `{"ids":[1,2,3]}` and returns `{ "descriptions": [{ "id": 1, "description_html": "..." }] }`.
- Use the batch endpoint when a page only needs rendered descriptions for known task IDs, such as ancestor descriptions on `/tasks/:id`, instead of expanding the task list payload with rendered HTML for every task.
- Frontend should render the sanitized HTML inside the existing description table cell using a scoped markdown container class.
- Use GitHub markdown CSS for headings, lists, links, blockquotes, tables, and code blocks.
- The GitHub markdown stylesheet must be scoped so it only applies inside elements with `class="apply-markdown"`.
- Dark markdown styling should use Quasar's `body.body--dark` class and still target only `.apply-markdown` descendants.
- Long URLs and code must not stretch the page; text should wrap, while code blocks may scroll horizontally inside their own block.

The final implementation should make the row groups understandable without relying only on color. Prefer subtle section labels or spacing for:

- ancestor path
- current task
- child tasks

If rows become clickable, do not wrap `<tr>` elements in `<a>` tags. Use row click handlers, router links inside cells, or Quasar table row APIs so the rendered table remains valid HTML.

## Development Workflow

After every app-affecting task, leave both local development servers running and report their local URLs:

- Backend: run `go run ./cmd/server` from `backend/`; expected API URL is `http://localhost:8080`.
- Frontend: run `pnpm dev` from `frontend/`; expected app URL is `http://localhost:8000`.
