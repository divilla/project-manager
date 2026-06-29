# State-Based Navigation For `mch`

## Goal

Build the first complete state-based navigation shell for `mch` so users can move through Main, Changes, Requirements, Epics, Projects, selectors, confirmations, help, find, and quit behavior with deterministic dummy screen rendering and focused transition tests.

## Scope

- Add the local `make-a-change/.config/config.yaml` file with the default backend URL and current project ID used by `mch`.
- Replace the Hello World app state with a navigation model rooted at `MainState`.
- Add screen states and dummy screen titles for Main, Changes, Requirements, Epics, Projects, selectors, confirmations, help, and find input.
- Add shared dropdown behavior for slash commands, list selection, backend selectors, and confirmations.
- Load project, epic, phase, and type selector options from the existing backend API client boundary.
- Add model and rendering tests for every state transition and dummy screen title included in this Change.
- Use `github.com/ridgelines/go-config` to load local YAML configuration.
- Clarify the `mch` architecture documentation for navigation-shell state, command, selector, escape, and persistence boundaries.

## Requirements

- `mch` starts in `MainState` and renders `MainScreen - Title: Main`.
- `MainState` accepts `/new-change`, `/changes`, `/epics`, `/projects`, `/select-project`, `/help`, and `/quit`, with `/new-change` first in the command list.
- `/changes` opens `ChangesListState`, renders `ChangesListScreen - Title: Changes List`, and exposes `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order; those filter commands stay inside `ChangesListState` and are not separate states or screens.
- Phase, epic, and type filter option lists render normal options with a leading `-`, such as `-done`, and append `/clear` as the final item to remove only that field's filter.
- Change selection opens `ChangeDetailsState`, renders `ChangeDetailsScreen - Title: Change Details`, and exposes exactly `/new-requirement`, `/phase`, `/epic`, `/types`, `/edit`, `/delete`, and `/return` in that order.
- Requirement selection opens `RequirementDetailsState`, renders `RequirementDetailsScreen - Title: Requirement Details`, and supports `/new-requirement`, `/edit`, delete, save, cancel, and return behavior.
- `ChangeDetailsState /phase` opens `SelectPhaseDropDown`, loads phases from `POST /api/v1/change/reference`, and returns to `ChangeDetailsState` after selection or cancel.
- `ChangeDetailsState /epic` opens `SelectEpicDropDown`, loads epics from `POST /api/v1/epic/list` with the current `project_id`, and returns to `ChangeDetailsState` after selection or cancel.
- `ChangeDetailsState /types` opens `SelectTypesDropDown`, loads type slugs from the `types` group returned by `POST /api/v1/change/reference`, and returns to `ChangeDetailsState` after selection or cancel.
- `/epics` opens `EpicsListState`, renders `EpicsListScreen - Title: Epics List`, and supports list, detail, `/new-epic`, `/edit`, delete, help, find, and return transitions.
- `/projects` opens `ProjectsListState`, renders `ProjectsListScreen - Title: Projects List`, and supports list, detail, `/new-project`, `/edit`, delete, help, find, and return transitions.
- Startup loads `backend_url` and `project_id` from `make-a-change/.config/config.yaml`; when `project_id` is greater than zero, `mch` uses it as the current project context for project-scoped selectors.
- When startup has no positive `project_id`, `mch` triggers the same project selector behavior as `/select-project` from `MainState`; if the backend returns no projects, `mch` stays on `MainState` and shows `No projects to select from. Please create new project and select it on Main Screen.`
- `/select-project` opens `SelectProjectDropDown`, loads projects from `POST /api/v1/project/list`, stores the selected current project in TUI state, saves its numeric `project_id` to `make-a-change/.config/config.yaml`, and returns to `MainState`.
- `/help` from `MainState` opens `MainHelpState`; `/find` highlights matching help text; `/return` returns to `MainState`.
- Slash command dropdowns and list selection dropdowns allow filtering, up/down highlighted selection, and confirmation of the highlighted option.
- Slash command dropdowns opened with `/` are overlays and must preserve the active state and rendered screen title while showing the command list below the page content.
- Confirmation dropdowns accept `/yes` or `/cancel`; any other command leaves the confirmation active and shows a recoverable error.
- `Esc` behaves like `/quit` from `MainState`, `/return` from returnable states, and `/cancel` from create, update, dropdown, confirmation, loading selector, and input states.
- `/quit` outside `MainState` leaves the current state unchanged and shows a recoverable error.
- Unknown commands leave the current state unchanged and show a recoverable error.
- Internal state names keep CRUD-style `CreateState` suffixes, but all user-facing creation commands are context-specific `/new-change`, `/new-requirement`, `/new-epic`, or `/new-project`, and all create-state screen titles use `New ...` wording.
- Internal state names keep CRUD-style `UpdateState` suffixes, but all user-facing update commands are `/edit` and all update-state screen titles use `Edit ...` wording.
- Save, delete, filter, selector, and selection transitions are navigation-only in this Change and must not write directly to the database; `/select-project` may write only local CLI config.

## Acceptance Criteria

- `make-a-change/.config/config.yaml` exists and contains `backend_url: http://localhost:8080` and `project_id: 0`.
- Starting `mch` initializes `MainState` and renders `MainScreen - Title: Main`.
- Main, Changes, Requirements, Epics, Projects, selector, confirmation, help, find, cancel, return, and quit transitions follow the navigation-shell rules in `docs/architecture/mch.md`.
- Backend selectors request options through the documented API endpoints and expose recoverable loading failures without losing the previous state.
- Selecting a current project stores it in TUI state, saves its numeric project ID to local config, and returns to `MainState`.
- A fresh checkout must keep `make-a-change/.config/config.yaml` at `project_id: 0`; non-zero project IDs are user-local state written only after project selection.
- Epic selectors and `/epic-filter` send numeric `project_id` values to `POST /api/v1/epic/list`; they must not send JSON strings for project IDs.
- `/quit` exits only from `MainState`; outside `MainState` it reports a recoverable error without changing state.
- Every state transition has a focused model test.
- Rendering tests verify every dummy screen title exactly.
- Tests prove no save, delete, filter, selector, or selection transition writes directly to the database; tests also cover local config project ID load/save.

## Non-Goals

- Do not implement real Change, Requirement, Epic, or Project create/update/delete persistence.
- Do not change backend API contracts.
- Do not add database migrations or seed changes.
- Do not implement full forms beyond the controls needed for navigation states.
- Do not implement Codex-assisted planning flows.
- Do not change the frontend SPA.
- Do not add production packaging or installer behavior.

## Design Notes

- `docs/architecture/mch.md` defines the authoritative navigation shell, package boundaries, model/update/view separation, style tokens, and test strategy.
- `docs/architecture/backend-api.md` defines the `POST /api/v1/change/reference`, `POST /api/v1/epic/list`, and `POST /api/v1/project/list` endpoints used by selector dropdowns.
- `docs/functionality/current-project-context.md` defines current project selection and repair behavior that `mch` should mirror when selector data is available.
- `docs/functionality/change-lifecycle.md` keeps backend APIs authoritative for Change lifecycle persistence; this Change deliberately keeps save and delete navigation-only.
- Treat the rough `ChangesDetailsState` spelling as `ChangeDetailsState` in implementation, tests, and UI copy.
- `mch` tests must use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` for assertions; avoid hand-written `t.Fatal` assertion blocks.
- Local `mch` config is read through `github.com/ridgelines/go-config`; writes are limited to `.config/config.yaml` keys owned by `mch`, currently `backend_url` and `project_id`.

## Relevant Specs

- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/current-project-context.md`
- `docs/functionality/change-lifecycle.md`
- `docs/operations/verification.md`

## Verification

- From the repository root: `cd make-a-change && go test ./...`
- From the repository root: `cd make-a-change && go build -o ./bin/mch ./cmd/mch`
- From the repository root: `cd make-a-change && ./bin/mch --version`
- From the repository root: `find docs -type f -name '*.md' -not -path 'docs/research/*' -exec wc -l {} +`

## Review Focus

- Confirm the navigation shell covers every state and transition without slipping persistence into navigation-only actions.
- Confirm selector loading uses backend API boundaries and preserves previous state on failure or cancel.
- Confirm `Esc`, `/quit`, unknown command, and confirmation-command behavior are tested for both allowed and disallowed states.
- Confirm dummy screen title tests are exact and deterministic.

## Follow-Ups

- None.
