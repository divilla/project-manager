# Selectable Projects Table In `mch`

Types: feature|test

## Problem Statement

`mch` currently shows `ProjectsListState` as a placeholder selection flow instead of an actual project list. Users need `ProjectsListScreen` to reload project data, render it as a selectable table, move selection with the keyboard, and open `ProjectDetailsState` for the selected project.

## Primary Workflows

1. A user opens `/projects` from `MainState`.
2. `ProjectsListState` reloads projects from `POST /api/v1/project/list`.
3. `ProjectsListScreen` renders a selectable table with columns `#id`, `Name`, `Change Count`, `Created`, and `Modified`.
4. The user presses up/down arrows to move the selected row.
5. The user presses return/enter to open `ProjectDetailsState` with the selected project object.
6. While on `ProjectsListState`, the user presses `/` to open the `CommandMenu`.

## Acceptance Criteria

1. `ProjectsListState` requests fresh project data from `POST /api/v1/project/list` each time `/projects` opens.
2. Each project row renders `id`, `name`, `change_count`, `created`, and `modified`.
3. Column labels render exactly as `#id`, `Name`, `Change Count`, `Created`, and `Modified`.
4. The `id` column displays IDs prefixed with `#`.
5. `created` and `modified` display in `YYYY-MM-DD HH:mm` format.
6. `created` and `modified` are converted to the current local timezone before display.
7. When projects exist, exactly one row is selected by default.
8. Pressing down moves selection to the next row when one exists.
9. Pressing up moves selection to the previous row when one exists.
10. Pressing up on the first row keeps the first row selected.
11. Pressing down on the last row keeps the last row selected.
12. Pressing return/enter on a selected row transitions to `ProjectDetailsState`.
13. `ProjectDetailsState` receives the selected project object with `id`, `name`, `change_count`, `created`, and `modified`.
14. Pressing return/enter on a project does not update the current project context.
15. Pressing `/` on `ProjectsListState` opens the `CommandMenu`.
16. The `CommandMenu` preserves the underlying `ProjectsListScreen - Title: Projects List` screen title.
17. Existing project commands `/new-project`, `/help`, `/find`, and `/return` remain available.
18. Tests cover project reload, table rendering, timestamp formatting, row selection, arrow navigation, enter transition, current project non-mutation, and `CommandMenu` opening.

## Edge Cases

1. Backend project list loading fails.
2. Backend returns no projects.
3. User presses enter when no project is selectable.
4. Backend returns invalid project IDs or malformed timestamps.
5. Terminal width is too narrow for all columns.

## Non-Goals

1. No project create, update, delete, or persistence changes.
2. No frontend SPA changes.
3. No backend API contract changes unless existing project list data is insufficient.
4. No pagination, sorting, filtering, or search changes.
5. Selecting a project from the table does not set or persist current project context.

## Dependencies And Risks

1. Depends on `POST /api/v1/project/list` returning the required project fields.
2. Requires `mch` to store selected project data beyond the current placeholder `api.Option`.
3. Depends on the existing slash-triggered menu behavior, named here as `CommandMenu`.
4. Live backend type reference verification was unavailable in this session; `feature|test` should be checked against `POST /api/v1/change/reference` before saving.

## Open Questions

None.
