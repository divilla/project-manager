# Selectable Projects Table In `mch`

## Goal

Replace the placeholder `ProjectsListState` flow in `mch` with a freshly loaded, keyboard-selectable projects table that opens `ProjectDetailsState` for the highlighted project without mutating the current project context.

## Scope

- Load project data through the existing `POST /api/v1/project/list` backend endpoint whenever `/projects` opens from `MainState`.
- Render `ProjectsListScreen - Title: Projects List` as a selectable table with project ID, name, changes, created timestamp, and modified timestamp columns.
- Support up/down row navigation, bounded selection behavior, enter-to-details transition, and slash command menu opening while the projects table is active.
- Preserve existing project commands and navigation shell behavior around `/new-project`, `/help`, `/find`, `/return`, and command overlays.
- Refactor the `mch` code layout so DTOs live in `internal/dto`, HTTP backend access lives in `pkg/client`, section behavior lives under `internal/projects`, `internal/changes`, `internal/epics`, `internal/requirements`, `internal/planning`, and `internal/help`, and shared routing/UI helpers live under `internal/navigation` and `internal/ui`.
- Prevent package tests or nested execution from creating `.config` folders under `cli/internal`.
- Add tests for project reload, table rendering, timestamp formatting, selection, details transition, current project non-mutation, and command menu behavior.

## Requirements

- `ProjectsListState` must request fresh project data from `POST /api/v1/project/list` each time `/projects` opens.
- Each project row must render `id`, `name`, `change_count`, `created`, and `modified`.
- The table headers must render exactly as `id`, `Name`, `Changes`, `Created`, and `Modified`.
- Project IDs must display as bare numeric IDs without a leading `#`.
- The `id` and `Changes` columns must be right-aligned.
- `created` and `modified` timestamps must be converted to the current local timezone and displayed as `YYYY-MM-DD HH:mm`.
- When projects exist, exactly one row must be selected by default.
- Down arrow must move selection to the next row when a next row exists.
- Up arrow must move selection to the previous row when a previous row exists.
- Up arrow on the first row must keep the first row selected.
- Down arrow on the last row must keep the last row selected.
- Enter on a selected project row must transition to `ProjectDetailsState`.
- `ProjectDetailsState` must receive the selected project object with `id`, `name`, `change_count`, `created`, and `modified`.
- Entering a project details view from the table must not update, save, or otherwise mutate the current project context.
- Pressing `/` from `ProjectsListState` must open the command menu while preserving the underlying `ProjectsListScreen - Title: Projects List` title.
- Existing project commands `/new-project`, `/help`, `/find`, and `/return` must remain available from the project list flow.
- Loading failures, empty project lists, malformed timestamps, invalid project IDs, no-selectable-row enter presses, and narrow terminal widths must produce deterministic, non-panicking UI behavior.
- Shared DTOs must live under `internal/dto`; the reusable HTTP backend client must live under `pkg/client`; feature packages must expose section-local `api.go` interfaces.
- `internal/app/model.go` must remain focused on root model state and construction, with update, dropdown, selector, command, and view responsibilities split into separate files.
- No `.config` directory may be created or tracked under `cli/internal`.

## Acceptance Criteria

- Opening `/projects` from `MainState` calls the project list endpoint and renders `ProjectsListScreen - Title: Projects List`.
- The rendered table includes the exact headers `id`, `Name`, `Changes`, `Created`, and `Modified`.
- Rendered rows show bare project IDs, names, change counts, and local-time `created` and `modified` values formatted as `YYYY-MM-DD HH:mm`, with ID and Changes values right-aligned.
- A non-empty project list starts with one selected row, supports bounded up/down movement, and keeps selection stable at the first and last rows.
- Enter on a selected row transitions to `ProjectDetailsState` with the complete selected project data.
- Enter on a selected row leaves the current project context unchanged and does not write `.config/config.yaml`.
- Pressing `/` opens the command menu and keeps the project list screen title visible under the overlay.
- Existing project commands remain in the project list command set and continue to transition according to the navigation shell.
- The codebase contains the documented package layout from `docs/architecture/mch.md`, including `pkg/client`, `internal/dto`, feature packages, `internal/navigation`, and `internal/ui`.
- Running tests from nested packages resolves local config to the `cli` module root rather than creating nested internal `.config` folders.
- Tests cover reload, rendering, timestamp formatting, row selection, arrow navigation, enter transition, current project non-mutation, and command menu opening.

## Non-Goals

- No project create, update, delete, or persistence changes.
- No frontend SPA changes.
- No backend API contract changes unless existing project list data is insufficient.
- No pagination, sorting, filtering, or search behavior changes.
- Selecting a project from the table does not set or persist current project context.
- No changes to startup current-project selector behavior.
- No product behavior changes beyond project list rendering/selection and architecture-preserving code movement.

## Design Notes

- `docs/architecture/mch.md` defines the authoritative `mch` navigation shell, shared command overlay behavior, and table-selection expectations for `ProjectsListState`.
- `docs/architecture/backend-api.md` defines `POST /api/v1/project/list` and documents that project list responses include `change_count`.
- `docs/functionality/current-project-context.md` defines current project context behavior; this Change intentionally opens project details without changing that context.
- The `mch` API layer may need to keep a richer project representation than the existing selector-only `api.Option` shape so details navigation can carry `id`, `name`, `change_count`, `created`, and `modified`.
- Documentation in `docs/architecture/mch.md` defines the required package boundaries for future agents.
- Tests must continue to use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`.

## Relevant Specs

- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/current-project-context.md`

## Verification

- From the repository root: `cd cli && make lint`
- From the repository root: `cd cli && go test ./...`
- From the repository root: `cd cli && go build -o ./bin/mch ./cmd/mch`

## Review Focus

- Verify the project-list API parser preserves full project fields and still supports existing selector behavior.
- Check that table selection and command menu state do not conflict with the existing shared dropdown/navigation model.
- Confirm details navigation does not mutate local current project state or write `.config/config.yaml`.
- Inspect empty, error, malformed timestamp, and narrow-width rendering paths.
- Confirm the refactor preserves existing navigation behavior while moving code to the documented package boundaries.

## Follow-Ups

- None.
