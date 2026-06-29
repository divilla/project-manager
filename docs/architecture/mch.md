# `mch` TUI Architecture

## Purpose

`mch` is the Go terminal UI for planning Changes. The formal app name is `Make a Change`, but product documentation, UI labels, command examples, requirements, tests, and executable references use `mch` unless an approved about or version view explicitly needs the formal name.

The first executable version is `0.1`. The executable name is `mch`.

## Libraries

`mch` uses:

- Bubble Tea for the application loop, model updates, messages, and commands
- Bubbles for reusable terminal controls such as text input, viewport, spinner, and list behavior
- Lip Gloss for rendering styles and layout tokens
- `github.com/ridgelines/go-config` for loading local YAML configuration

## Package Boundaries

Recommended layout:

```text
make-a-change/
  cmd/mch/
  pkg/client/
  internal/dto/
  internal/app/
  internal/projects/
  internal/changes/
  internal/epics/
  internal/requirements/
  internal/planning/
  internal/help/
  internal/navigation/
  internal/ui/
  internal/styles/
```

Responsibilities:

- `cmd/mch`: parse process arguments only far enough to call the app runner and set exit status.
- `pkg/client`: own reusable HTTP transport, endpoint methods, JSON request/response helpers, and backend error normalization.
- `internal/dto`: own shared request, response, and view DTOs such as selector options and project rows.
- `internal/app`: own startup wiring, config, version output, the root Bubble Tea model, and dispatch between feature packages. It must not own feature-specific table rendering or section navigation rules.
- `internal/projects`: own project list/detail/create/update/delete state, project commands, project navigation, project rendering, and a section-local `api.go` interface used by the app.
- `internal/changes`: own change list/detail/create/update/delete state, filters, change commands, change navigation, change rendering, and a section-local `api.go` interface.
- `internal/epics`: own epic list/detail/create/update/delete state, epic commands, epic navigation, epic rendering, and a section-local `api.go` interface.
- `internal/requirements`: own requirement detail/create/update/delete state, requirement commands, requirement navigation, requirement rendering, and a section-local `api.go` interface.
- `internal/planning`: own future AI-assisted planning flow state, commands, navigation, rendering, and a section-local `api.go` interface.
- `internal/help`: own help screen state, commands, navigation, and rendering.
- `internal/navigation`: own shared state names, screen titles, return/cancel/delete route targets, and cross-section command assembly.
- `internal/ui`: own reusable terminal UI primitives such as dropdowns, input bands, layout helpers, and generic table helpers.
- `internal/styles`: define Lip Gloss style tokens and shared components.

Each feature package must split code by responsibility instead of using one catch-all file. Use `model.go` for feature state/data methods, `navigation.go` or `update.go` for key and route decisions, `view.go` or `screen.go` for rendering, `commands.go` for slash commands, and `api.go` for the app-facing API interface. Feature packages should not import `internal/app`; the root app should call feature packages.

The tracked local config file is `make-a-change/.config/config.yaml`. No package under `internal/` may create or persist a `.config` directory. Config path resolution must anchor local config at the `make-a-change` module root, including when tests run from nested package directories.

## Model And Commands

The root Bubble Tea `Model` owns current screen, window size, command menu state, current project context, visible errors, and reusable component models. It should delegate screen-specific decisions to focused helpers rather than embedding full workflows in one method.

`Update` should only translate messages into state changes and `tea.Cmd` values. It should not perform HTTP requests, file writes, editor launches, or AI calls directly.

`tea.Cmd` functions should wrap asynchronous work and return typed messages. Backend API calls and long-running AI calls must be cancellable through `context.Context` where possible. A running AI call should update the UI through loading messages and then return either a structured result message or an error message.

`View` should render current state from model data only. Rendering must not mutate state, read files, call APIs, or start processes.

## Planning States

Future Change planning flows should use these states:

- `ready`: project context is valid and the app is ready for a planning command.
- `project selection`: no current project is selected or the saved project is invalid.
- `idea entry`: the user is entering or refining a Change idea.
- `AI running`: an async AI command is active and progress metadata is visible.
- `review`: generated requirement markdown is available for review.
- `save confirmation`: parsed output is ready to persist through backend APIs.
- `error`: recoverable failure with a visible reason and next action.
- `done`: the planned Change has been saved or the flow has exited cleanly.

Slash commands should be accepted only in states that define them. Unknown commands should leave user input intact and show a recoverable error.

## Navigation Shell

The navigation shell starts in `MainState` and renders deterministic screen titles so state transitions are easy to test. Initial render is `MainScreen - Title: Main`.

Top-level commands from `MainState`:

- `/new-change` opens `ChangeCreateState` and is the first command in the Main command list.
- `/changes` opens `ChangesListState` and renders `ChangesListScreen - Title: Changes List`.
- `/epics` opens `EpicsListState` and renders `EpicsListScreen - Title: Epics List`.
- `/projects` opens `ProjectsListState` and renders `ProjectsListScreen - Title: Projects List`.
- `/select-project` opens `SelectProjectDropDown`, loads projects through `POST /api/v1/project/list`, saves the selected current project in TUI state, writes its numeric `project_id` to `.config/config.yaml`, and returns to `MainState`.
- `/help` opens `MainHelpState`; `/find` opens `FindInput`; `/return` returns to `MainState`.
- `/quit` exits only from `MainState`.

Slash commands, list item selection, backend selectors, confirmations, and text search should use one shared dropdown or input interaction model where practical. Users can filter dropdown options, move the highlighted option with arrow keys, and confirm the highlighted option. Command dropdowns are overlays: opening the command list with `/` must preserve the active state and screen title while rendering commands below the page content. Selector dropdowns load options when opened, display recoverable errors when loading fails, and preserve the previous state on cancel.

Changes navigation includes `ChangesListState`, `ChangeDetailsState`, `RequirementDetailsState`, Change create and update states, Requirement create and update states, filter overlays, help, find input, and delete confirmation states. Create-state commands are context-specific: `/new-change`, `/new-requirement`, `/new-epic`, or `/new-project`; internal state names still keep CRUD-style `CreateState` suffixes. Update-state commands are named `/edit` even though internal state names keep CRUD-style `UpdateState` suffixes. Create-state screen titles use user-facing `New ...` wording, such as `ChangeCreateScreen - Title: New Change`; update-state screen titles use `Edit ...` wording, such as `ChangeUpdateScreen - Title: Edit Change`. `ChangesListState` exposes exactly `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order; `/phase-filter`, `/epic-filter`, `/type-filter`, and `/find-filter` remain inside `ChangesListState` and must not be modeled as separate states or screens. Phase, epic, and type filter option lists render normal options with a leading `-`, such as `-done`, and append `/clear` as the final item to remove only that field's filter. `ChangeDetailsState` exposes exactly `/new-requirement`, `/phase`, `/epic`, `/types`, `/edit`, `/delete`, and `/return` in that order. Requirement detail commands include new requirement, edit, delete, save, cancel, and return. Phase and type selectors load from `POST /api/v1/change/reference`; epic selectors and `/epic-filter` load from `POST /api/v1/epic/list` using the current project ID as a numeric JSON value.

Epics and Projects use the same state shape: list, detail, create, update, delete confirmation, help, find input, and return. List and detail screens may navigate to new epic or new project, edit, delete, help, find, and return states according to the commands available on each screen.

`ProjectsListState` loads fresh project data from `POST /api/v1/project/list` each time `/projects` opens from `MainState`. `ProjectsListScreen - Title: Projects List` renders projects as a selectable table with columns `id`, `Name`, `Changes`, `Created`, and `Modified`; ID and Changes values are right-aligned, and IDs display without a leading `#`. The created and modified timestamps are displayed in the current local timezone as `YYYY-MM-DD HH:mm`. Up and down arrows move the selected row within bounds, enter opens `ProjectDetailsState` with the selected project data, and this list selection must not update or persist the current project context. Pressing `/` from the list opens the command menu overlay while preserving the project list screen title underneath it.

`Esc` maps to the state-appropriate safe action: quit from `MainState`, return from returnable states, and cancel from create, update, dropdown, confirmation, loading selector, and input states. `/quit` outside `MainState` and unknown commands leave the current state unchanged and show a recoverable error.

Save, delete, filter, selector, and selection actions in the navigation shell are navigation-only until a later persistence Change. They must not write directly to the database.

## Backend And Persistence

Backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence. `mch` must not write application database tables directly.

Project-scoped commands should either use the saved current project context or require an explicit project option. When the saved project no longer exists, `mch` should clear or repair selection using the same behavior documented for current project context.

## Config

`mch` should load local config at startup, then apply command-line overrides such as backend URL for the current process. Config validation should reject missing or malformed backend URLs before project-scoped API calls.

`mch` loads `.config/config.yaml` at startup through `github.com/ridgelines/go-config`. The file owns `backend_url` and `project_id`; `project_id: 0` means no saved current project, and the tracked default config must stay at `project_id: 0`. When startup has no positive `project_id`, `mch` opens the same selector flow as `/select-project` from `MainState`. If the backend returns no projects, `mch` stays on `MainState` and shows `No projects to select from. Please create new project and select it on Main Screen.` It must not redirect to project creation. `/select-project` updates only this local config file and saves the selected project ID as a number. Product data must be saved only through backend APIs.

## Components

Reusable components should cover:

- prompt input
- command menu
- status/footer
- loading indicator
- error display
- output viewport
- confirmation prompt
- project selector

Components should accept width and state as inputs so narrow terminals do not produce overlapping text. When width is too small, content should truncate or stack before it clips important state.

## Style Tokens

The baseline style uses a dark terminal surface, full-width muted input band, compact monospace layout, cyan and purple accents, muted footer/status metadata, and minimal borders. This adapts the local Gemini CLI reference screenshots without copying Gemini branding, command names, or product copy.

Named Lip Gloss tokens:

- `Background`: dark terminal background
- `Foreground`: primary readable text
- `Muted`: secondary metadata text
- `InputBand`: full-width prompt and status band
- `Selection`: highlighted command or project selection
- `Error`: recoverable error text
- `Success`: completion text
- `AccentCyan`: primary interactive accent
- `AccentPurple`: secondary accent
- `Border`: low-contrast border color

UI text must remain product-specific to `mch`.

## Test Strategy

All future `mch` tests must use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require` for assertions. Do not add hand-written `if ... { t.Fatal... }` assertion blocks when a `testify` assertion can express the same expectation.

After every `make-a-change` code change, run `make lint` from `make-a-change` and fix all findings before handoff. Treat lint rewrites such as import formatting as part of the intentional code change.

Model tests should cover startup state, screen transitions, command parsing, async message handling, and cancellation paths.

Rendering tests should assert stable output for important strings, status bands, narrow widths, and no accidental `Make a Change` copy in regular UI.

API client tests should use HTTP test servers and must not inspect database tables directly.

Markdown parsing tests should cover valid generated requirements, invalid markdown, missing titles, unsupported type values, and editor round trips.

Config tests should cover missing files, malformed files, command-line overrides, saved backend URL, saved project ID, and invalid saved project repair.

## Follow-Up Work

After the Hello World baseline, add the real Change planning workflow, backend API integration, markdown validation, editor handoff, and a retirement plan for `cli-proto/`.
