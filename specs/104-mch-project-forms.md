# Persist Project Create And Edit In `mch`

## Goal

Make the `mch` project create, update, list, and details screens persist project names through the backend API, reload project data on arrival, and render project metadata and multiline prompt input consistently.

## Scope

- Implement project create saving from `ProjectCreateState` through `POST /api/v1/project/create`.
- Implement project update saving from `ProjectUpdateState` through `POST /api/v1/project/update`.
- Reload created and updated project records through `POST /api/v1/project/get` before rendering `ProjectDetailsState`.
- Reload project list and detail data every time the user arrives at project list or detail screens.
- Render `ProjectDetailsState` with right-aligned labels, project `#ID`, multiline name, change count, created timestamp, and modified timestamp.
- Render project table names with whole-word trimming and a width derived from the longest rendered name.
- Add global prompt behavior for Enter save, Shift+Enter newline, Ctrl+E editor handoff, Ctrl+C clear-then-cancel, multiline rendering, and visible cursor movement.
- Preserve create/update cancel behavior and keep project form persistence separate from current project context selection.
- Add focused tests for create, update, get-after-save, reload-on-arrival, validation, backend failure, cancel behavior, prompt controls, timestamp formatting, detail rendering, and config non-mutation.
- Update project API and `mch` architecture documentation for project form persistence and prompt behavior.

## Requirements

- `ProjectCreateState` must send `POST /api/v1/project/create` only after the submitted name is non-empty after trimming.
- Project create requests must send only the required JSON `name` field. The name must be sent exactly as entered, including newline characters, after trim-based validation passes.
- Successful project create must fetch the created project through `POST /api/v1/project/get` using the created project ID before opening `ProjectDetailsState`.
- `/edit` from `ProjectDetailsState` must open `ProjectUpdateState` with the selected project name prefilled.
- `ProjectUpdateState` must send `POST /api/v1/project/update`, not `POST /api/v1/project/create`.
- Project update requests must send project `id` as a JSON number and `name` exactly as entered, including newline characters, after trim-based validation passes.
- Project update must be blocked with a recoverable error when the selected project has no valid positive numeric ID.
- Successful project update must fetch the updated project through `POST /api/v1/project/get` using the updated project ID before refreshing `ProjectDetailsState`.
- Project list and project detail screens must call the backend every time the user arrives; stale cached data must not be reused as the source of truth.
- Failed create, update, or get requests must keep the user in a recoverable state and display the backend error.
- `/cancel` and `Esc` from project create must return to `ProjectsListState` without making a project create API call.
- `/cancel` and `Esc` from project update must return to `ProjectDetailsState` without making a project update API call.
- The prompt must use Enter as `/save` on screens where `/save` exists.
- The prompt must use Shift+Enter for a newline and grow vertically while preserving entered text.
- Ctrl+E must open `$EDITOR`, falling back to `nano`, with a temporary `.md` file. On editor exit, the content must be submitted immediately through the same flow as Enter and the terminal must be cleared.
- Ctrl+C must clear non-empty prompt text first. A second Ctrl+C on an empty prompt must run `/cancel`, `/return`, or `/quit`, depending on the active screen.
- Project create and update command menus must expose `/editor` as the first option, followed by `/save` and `/cancel`.
- `ProjectDetailsState` must display labels and values for `#ID`, `Name`, `Changes`, `Created`, and `Modified`.
- `ProjectDetailsState` labels must be shifted four spaces to the right, labels must align to the right, values must align to the left, `#ID` must render instead of `ID`, the ID value must be light pink, the name value must be bright cyan, and created/modified values must be grey between label grey and white.
- Project detail names must wrap at 80 characters without breaking words and must preserve explicit newline characters.
- Project table names longer than 80 characters must trim whole words from the right until the rendered name plus `...` is shorter than 78 characters, then append `...`.
- The project table Name column must be as wide as the maximum rendered project name.
- Created and modified timestamps must render in the current local timezone as `YYYY-MM-DD HH:mm`, for example `2026-06-29 13:04`.
- Timestamp rendering must truncate seconds and sub-second precision.
- Invalid or missing project timestamps must render exactly as `not a date`.
- Project create and update must not change `.config/config.yaml` or the current project selection.
- Narrow terminal widths must not cause overlapping project detail text.
- Tests must cover successful create, successful update, get-after-create, get-after-update, reload-on-arrival, validation failure, backend failure, cancel behavior, prompt controls, timestamp formatting, detail rendering, and current project config non-mutation.

## Acceptance Criteria

- Saving a project create form with `New Project` sends `POST /api/v1/project/create` with `{"name":"New Project"}` and no extra fields.
- Saving a project create form with `Line 1\nLine 2` preserves the newline in the `name` payload.
- Saving a project create form with an empty or whitespace-only name does not call the backend and shows a recoverable validation error.
- After a successful create response, `mch` calls `POST /api/v1/project/get` for the created project ID and renders the fetched project in `ProjectDetailsState`.
- Saving a project update form sends `POST /api/v1/project/update` with numeric `id` and the entered `name`.
- Updating a project with a missing, zero, negative, or non-numeric ID does not call the update endpoint and shows a recoverable error.
- After a successful update response, `mch` calls `POST /api/v1/project/get` for the updated project ID and refreshes `ProjectDetailsState` with the fetched project.
- Arriving at `ProjectsListState` calls `POST /api/v1/project/list`, and arriving at `ProjectDetailsState` calls `POST /api/v1/project/get`.
- Backend failures from create, update, or get remain visible to the user without losing the form state needed to retry or cancel.
- `/cancel` and `Esc` from create return to `ProjectsListState` without a create request, and `/cancel` and `Esc` from update return to `ProjectDetailsState` without an update request.
- Enter submits save-capable forms, Shift+Enter inserts and renders a newline, Ctrl+E opens the external editor, Ctrl+C clears prompt text before canceling, and up/down move the prompt cursor across multiline prompt rows.
- Project create and update command menus render `/editor`, `/save`, and `/cancel` in that order.
- `ProjectDetailsState` renders `#ID`, `Name`, `Changes`, `Created`, and `Modified` labels with values from the selected or fetched project using the documented alignment and colors.
- Project detail names wrap at 80 characters without breaking words and preserve explicit newline characters.
- Long project table names are trimmed by words with `...`, and the Name column width is derived from the longest rendered name.
- Created and modified timestamps render in local time to minute precision, and invalid or missing timestamps render as `not a date`.
- Create and update flows leave `.config/config.yaml` and the saved current project ID unchanged.
- `cd cli && make lint`, `cd cli && go test ./...`, and `cd cli && go build -o /tmp/mch ./cmd/mch` pass.

## Non-Goals

- No backend project schema changes or new project fields.
- No project delete behavior changes.
- No project list table, sorting, filtering, or pagination changes beyond interactions needed to reach create and detail screens.
- No current project selection or `.config/config.yaml` persistence changes.
- No direct database writes from `mch`.
- No frontend SPA project screen changes.

## Design Notes

- `docs/architecture/mch.md` defines the authoritative `mch` project state navigation, create/update form behavior, detail rendering, timestamp formatting, and current project context boundary.
- `docs/architecture/backend-api.md` defines the project endpoints and project create, update, and get payload expectations.
- `docs/functionality/current-project-context.md` defines current project selection behavior; this Change intentionally does not select a created or edited project as the current project.
- The initial rough idea named `ProjectUplateScreen` and suggested update should post to create; this Change treats the state as `ProjectUpdateState` and uses `/api/v1/project/update` to match repository API contracts.
- The prompt uses Bubble Tea key messages; current terminal behavior emits Shift+Enter as `Esc O M`, so the implementation treats that leaked sequence as newline and prevents `OM` from being inserted.
- Tests must continue to use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`.

## Relevant Specs

- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/current-project-context.md`

## Verification

- From the repository root: `cd cli && make lint`
- From the repository root: `cd cli && go test ./...`
- From the repository root: `cd cli && go build -o /tmp/mch ./cmd/mch`

## Review Focus

- Verify create and update send the documented endpoint-specific payloads and do not reuse the wrong endpoint.
- Check that get-after-create and get-after-update use the returned project ID and render fetched data rather than stale form data.
- Confirm project list/detail arrival paths reload from the backend instead of relying on cached data.
- Inspect validation, backend error, cancel, and retry paths for recoverable state behavior.
- Verify prompt keyboard behavior on project create/update screens, including Shift+Enter, Ctrl+E, Ctrl+C, Enter-as-save, and multiline cursor movement.
- Confirm timestamp formatting is shared or consistent with project list behavior and handles malformed values deterministically.
- Confirm project form saves do not mutate current project context or write `.config/config.yaml`.

## Follow-Ups

- Fixed PR comment `IC_kwDOTA2Xls8AAAABIDWidw`: `/save` now preserves raw project prompt values after trim-based validation, and the tracked default `cli/.config/config.yaml` contract is restored.
