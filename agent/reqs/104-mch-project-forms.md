# Persist Project Create And Edit In mch

Types: feature|test|docs

## Problem Statement

`mch` has project create, update, and detail states, but project save actions are currently navigation placeholders. Users need these screens to persist project data through the backend API and display project metadata consistently.

## Primary Workflows

1. From `ProjectsListState`, `/new-project` opens `ProjectCreateState`.
2. The user enters a project name and saves.
3. `mch` sends `POST /api/v1/project/create` with `{ "name": "<trimmed name>" }`.
4. On success, `mch` fetches the created project with `POST /api/v1/project/get` and opens `ProjectDetailsState`.
5. From `ProjectDetailsState`, `/edit` opens `ProjectUpdateState` with the current name prefilled.
6. Saving update sends `POST /api/v1/project/update` with `{ "id": <project id>, "name": "<trimmed name>" }`.
7. On success, `mch` fetches the updated project with `POST /api/v1/project/get` and refreshes `ProjectDetailsState`.
8. `ProjectDetailsState` displays `id`, `name`, `change_count`, `created`, and `modified`.

## Acceptance Criteria

1. Create save sends `POST /api/v1/project/create` only after the name is non-empty after trimming.
2. Create save sends only the required `name` field.
3. Successful create fetches the project through `POST /api/v1/project/get` using the created project id before rendering details.
4. Update save sends `POST /api/v1/project/update`, not `POST /api/v1/project/create`.
5. Update save sends project `id` as a JSON number and `name` as the trimmed string.
6. Successful update fetches the project through `POST /api/v1/project/get` using the updated project id before refreshing details.
7. Update is blocked when the selected project has no valid positive numeric id.
8. Failed create, update, or get keeps the user in a recoverable state and displays the backend error.
9. `/cancel` and `Esc` from create return to `ProjectsListState` without an API call.
10. `/cancel` and `Esc` from update return to `ProjectDetailsState` without an API call.
11. Details display labels and values for `ID`, `Name`, `Changes`, `Created`, and `Modified`.
12. Created and modified render in local timezone as `YYYY-MM-DD HH:mm`, for example `2026-06-29 13:04`.
13. Timestamp rendering truncates seconds and sub-second precision.
14. Invalid or missing timestamps render exactly `not a date`.
15. Create/update does not change `.config/config.yaml` or current project selection.
16. Tests cover successful create, successful update, get-after-create, get-after-update, validation failure, backend failure, cancel behavior, timestamp formatting, and detail rendering.

## Edge Cases

1. Empty or whitespace-only project name.
2. Backend returns `400 invalid project payload`.
3. Backend returns `404 project not found` during update or get.
4. Backend is unavailable.
5. User saves the same name unchanged.
6. Project timestamps are missing or malformed.
7. Narrow terminal width must not cause overlapping detail text.

## Non-Goals

1. Do not add new backend project fields.
2. Do not change project delete behavior.
3. Do not select the created project as current project context.
4. Do not mutate `.config/config.yaml`.
5. Do not write directly to the database.
6. Do not implement frontend SPA project screens.

## Dependencies And Risks

1. Depends on `POST /api/v1/project/create`, `POST /api/v1/project/update`, `POST /api/v1/project/get`, and `POST /api/v1/project/list`.
2. Depends on backend project fields `id`, `name`, `created`, `modified`, and `change_count`.
3. The initial idea named `ProjectUplateScreen`; this requirement treats that as `ProjectUpdateScreen`.
4. The initial idea said update should POST to create; repository contracts show update must use `/api/v1/project/update`.
5. Live reference data was unavailable, so `Types:` should be verified against `POST /api/v1/change/reference` before saving.

## Open Questions

None.
