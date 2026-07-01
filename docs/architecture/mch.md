# `mch` TUI Architecture

## Purpose

`mch` is the Go terminal UI for planning Changes. The formal app name is `Make a Change`, but product documentation, UI labels, command examples, specs, tests, and executable references use `mch` unless an approved about or version view explicitly needs the formal name.

The first executable version is `0.1`. The executable name is `mch`.

## Libraries

`mch` uses:

- Bubble Tea for the application loop, model updates, messages, and commands
- Bubbles for reusable terminal controls such as textarea-backed prompt input, viewport, spinner, and list behavior
- Lip Gloss for rendering styles and layout tokens
- `github.com/ridgelines/go-config` for loading local YAML configuration

## Package Boundaries

Recommended layout:

```text
cli/
  cmd/mch/
  pkg/client/
  internal/dto/
  internal/app/
  internal/projects/
  internal/changes/
  internal/epics/
  internal/testcases/
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
- `internal/testcases`: own test case detail/create/update/delete state, test case commands, test case navigation, test case rendering, and a section-local `api.go` interface.
- `internal/planning`: own future AI-assisted planning flow state, commands, navigation, rendering, and a section-local `api.go` interface.
- `internal/help`: own help screen state, commands, navigation, and rendering.
- `internal/navigation`: own shared state names, screen titles, return/cancel/delete route targets, and cross-section command assembly.
- `internal/ui`: own reusable terminal UI primitives such as dropdowns, input bands, layout helpers, and generic table helpers.
- `internal/styles`: define Lip Gloss style tokens and shared components.

Each feature package must split code by responsibility instead of using one catch-all file. Use `model.go` for feature state/data methods, `navigation.go` or `update.go` for key and route decisions, `view.go` or `screen.go` for rendering, `commands.go` for slash commands, and `api.go` for the app-facing API interface. Feature packages should not import `internal/app`; the root app should call feature packages.

The local config file is `cli/.config/config.yaml`. No package under `internal/` may create or persist a `.config` directory. Config path resolution must anchor local config at the `cli` module root, including when tests run from nested package directories.

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
- `review`: generated test case markdown is available for review.
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

Changes navigation includes `ChangesListState`, `ChangeDetailsState`, `TestCaseDetailsState`, Change create and update states, Test Case create and update states, filter overlays, help, find input, and delete confirmation states. Create-state commands are context-specific: `/new-change`, `/new-test-case`, `/new-epic`, or `/new-project`; internal state names still keep CRUD-style `CreateState` suffixes. Update-state commands are named `/edit` even though internal state names keep CRUD-style `UpdateState` suffixes. Create-state screen titles use user-facing `New ...` wording, such as `ChangeCreateScreen - Title: New Change`; update-state screen titles use `Edit ...` wording, such as `ChangeUpdateScreen - Title: Edit Change`. `ChangesListState` exposes exactly `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order; `/phase-filter`, `/epic-filter`, `/type-filter`, and `/find-filter` remain inside `ChangesListState` and must not be modeled as separate states or screens. Phase, epic, and type filter option lists render normal options with a leading `-`, such as `-done`, and append `/clear` as the final item to remove only that field's filter. `ChangeDetailsState` exposes exactly `/new-test-case`, `/phase`, `/epic`, `/types`, `/edit`, `/delete`, and `/return` in that order. Test case detail commands include new test case, edit, delete, save, cancel, and return. Phase and type selectors load from `POST /api/v1/change/reference`; epic selectors and `/epic-filter` load from `POST /api/v1/epic/list` using the current project ID as a numeric JSON value.

Change list and detail screens should use backend-provided `ref` and `slug` as read-only identity data. Change create and edit states must not prompt for, submit, or locally derive `ref`, `slug`, or project reference counters.

`ChangesListState` loads Changes from `POST /api/v1/change/list` with the current numeric `project_id` every time the user opens `/changes` and displays rows in the backend response order. It renders a boxed, scrollable selectable table with columns `#Ref`, `Phase`, `Types`, `Epic`, `Title`, `Don`, `Tot`, `%`, and `Modified`, in that order. Numeric Change refs render as six digits with leading zeroes and no `#` in row values, such as `000003`. `Phase` values render in a 10-character column with bright phase-specific colors; backlog is white and progress is bright cyan. `Types` values are at most 30 characters wide, `Epic` values are at most 20 characters wide, and `Title` values are at most 80 characters wide; longer values truncate at that position without a suffix. Title values render pure white. The table renders at its natural column width when the terminal is wide enough and shrinks columns only when the available terminal width is smaller than that natural table width. `Don`, `Tot`, and `%` show done test cases, total test cases, and completed percentage from the backend response. `%` values render bright cyan. `Modified` renders as `YYYY-MM-DD HH.MM`; missing or invalid timestamps render as `not a date`. Up and down arrows move the selected row within bounds, PgUp and PgDown move by one visible page, and navigation keeps the selected row inside the visible table viewport. Enter or Return loads the selected Change through `POST /api/v1/change/get` before opening `ChangeDetailsState`. List load failures show a recoverable error.

`ChangeDetailsState` renders `ChangeDetailsScreen - Title: Change Details` from the backend Change response. Details show title, phase, types, epic, closed state, requirement content, and read-only `ref` and `slug`. `/edit` opens `ChangeUpdateState` with the current Change available for editor-based requirement markdown editing. Dedicated detail commands remain available for focused phase, epic, and type edits; editor save also repopulates title, types, epic, and `requirement_body` from the markdown metadata.

`ChangeCreateState` renders `ChangeCreateScreen - Title: New Change` and exposes `/save` and `/cancel`. Opening Change create must open the external editor; the only editable input is the requirement markdown body. Save parses title, types, and optional epic from that body, validates the extracted fields, then creates the Change through `POST /api/v1/change/create` with numeric `project_id`, title, full preserved `requirement_body`, one or more `change_types`, and optional `epic_id`. Successful create opens details for the created Change. Validation or backend failures keep the user in a recoverable create state with the editor content available to fix.

`ChangeUpdateState` renders `ChangeUpdateScreen - Title: Edit Change` and exposes `/save` and `/cancel`. Opening Change edit must open the external editor with the current Change represented in the same requirement markdown format. Save parses title, types, and optional epic from the edited body, validates the extracted fields, and persists changed title, requirement text, types, and epic through `POST /api/v1/change/update-title`, `POST /api/v1/change/update-requirement-body`, `POST /api/v1/change/update-change-types`, and `POST /api/v1/change/update-epic`. Successful update reloads the Change through `POST /api/v1/change/get` before refreshing details. If only one field changed, only that field's endpoint is called.

Change create and update use the strict requirement body contract from requirement planning. The first non-blank line must be an H1 title and becomes the Change `title`. The first non-blank line after the H1 must be exactly `Types: <type-slugs>`, where slugs are backend type slugs joined by `|` with no spaces; missing or blank `Types:` is a validation error. The next non-blank line may be `Epic: <epic-name>`; omitted or blank epic means no epic. Type slugs are validated against `POST /api/v1/change/reference`; epic names are resolved from `POST /api/v1/epic/list` for the current project. Missing title or missing types prevents any create or update API call. The full editor markdown, including the H1, `Types:`, optional `Epic:`, and all following sections, is always preserved as `requirement_body`.

`/cancel` and `Esc` from Change create return to `ChangesListState` without creating a Change. `/cancel` and `Esc` from Change update return to `ChangeDetailsState` without calling update endpoints.

Changes filters are list-local. `/phase-filter` and `/type-filter` load options from `POST /api/v1/change/reference`; `/epic-filter` loads epics from `POST /api/v1/epic/list` with numeric `project_id`. Each filter overlay remains on `ChangesListState`, keeps the list title visible, and appends `/clear` as the final option to clear only that filter field. `/find-filter` applies text filtering to the loaded list by title, `ref`, `slug`, phase, type, epic, and loaded requirement text when present. `/clear-filters` clears phase, type, epic, and find filters and restores the unfiltered loaded list. A filter with no matches renders a no-results state.

Epics and Projects use the same state shape: list, detail, create, update, delete confirmation, help, find input, and return. List and detail screens may navigate to new epic or new project, edit, delete, help, find, and return states according to the commands available on each screen.

`ProjectsListState` loads fresh project data from `POST /api/v1/project/list` every time the user arrives at the screen; project rows are not served from a cache. `ProjectsListScreen - Title: Projects List` renders projects as a selectable table with columns `id`, `Name`, `Changes`, `Created`, and `Modified`; ID and Changes values are right-aligned, and IDs display without a leading `#`. The Name column width is derived from the longest rendered project name. Names longer than 80 characters are normalized to single spaces, trimmed by removing whole words from the right until the rendered name plus `...` is shorter than 78 characters, then rendered with the `...` suffix. The created and modified timestamps are displayed in the current local timezone as `YYYY-MM-DD HH:mm`. Up and down arrows move the selected row within bounds when the prompt is empty, enter opens `ProjectDetailsState` with the selected project data, and this list selection must not update or persist the current project context. Pressing `/` from the list opens the command menu overlay while preserving the project list screen title underneath it.

`/new-project` opens `ProjectCreateState` with a project name form and the placeholder `Write a Name`. `ProjectCreateState` and `ProjectUpdateState` expose `/editor`, `/save`, and `/cancel` in that order. Saving validates that the name is non-empty after trimming, then sends the name exactly as entered, including explicit newline characters. Create sends `POST /api/v1/project/create` with only the required `name` field, then reloads the created project through `POST /api/v1/project/get` before opening `ProjectDetailsState`. `/edit` from `ProjectDetailsState` opens `ProjectUpdateState` with the current project name prefilled. Update requires a valid positive numeric project ID, sends `POST /api/v1/project/update` with numeric `id` and string `name`, then reloads the updated project through `POST /api/v1/project/get` before refreshing `ProjectDetailsState`. Create and update failures leave the user in the form with a recoverable backend or validation error, and cancel actions return without an API call.

`ProjectDetailsState` reloads the selected project through `POST /api/v1/project/get` every time the user arrives at the screen; detail rows are not served from a cache. The detail view displays labels `#ID`, `Name`, `Changes`, `Created`, and `Modified`. Labels are shifted four spaces to the right, right-aligned, and values are left-aligned. Normal values render white, the `#ID` value renders light pink, the name value renders bright cyan, and created/modified values render grey between the label grey and white. The name value wraps at 80 characters without breaking words and preserves explicit newline characters. Created and modified timestamps render in the current local timezone as `YYYY-MM-DD HH:mm`, truncating seconds and sub-second precision. Missing or invalid timestamps render as `not a date`.

`Esc` maps to the state-appropriate safe action: quit from `MainState`, return from returnable states, and cancel from create, update, dropdown, confirmation, loading selector, and input states. `/quit` outside `MainState` and unknown commands leave the current state unchanged and show a recoverable error.

Save, delete, filter, selector, and selection actions in the navigation shell must use backend APIs for persistence when implemented. They must not write directly to the database. Project create and update forms persist only project data and must not update `.config/config.yaml` or change the current project context.

## Backend And Persistence

Backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence. `mch` must not write application database tables directly.

Project-scoped commands should either use the saved current project context or require an explicit project option. When the saved project no longer exists, `mch` should clear or repair selection using the same behavior documented for current project context.

## Config

`mch` should load local config at startup, then apply command-line overrides such as backend URL for the current process. Config validation should reject missing or malformed backend URLs before project-scoped API calls.

`mch` loads `.config/config.yaml` at startup through `github.com/ridgelines/go-config`. The file owns `backend_url` and `project_id`; `project_id: 0` means no saved current project, and newly created local config must default to `project_id: 0`. When startup has no positive `project_id`, `mch` opens the same selector flow as `/select-project` from `MainState`. If the backend returns no projects, `mch` stays on `MainState` and shows `No projects to select from. Please create new project and select it on Main Screen.` It must not redirect to project creation. `/select-project` updates only this local config file and saves the selected project ID as a number. Product data must be saved only through backend APIs.

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

The shared prompt uses a textarea-backed model but renders with app-owned input-band styles so the visible prompt stays stable. Enter submits `/save` on screens that expose `/save`; otherwise it submits slash commands or list selection according to the current state. Shift+Enter inserts a newline, grows the prompt vertically, and preserves the entered text. Current terminal input can deliver Shift+Enter as the leaked `Esc O M` sequence; `mch` treats that sequence as newline and must not insert literal `OM`. Ctrl+E opens `$EDITOR`, falling back to `nano`, with a temporary `.md` file. On editor exit, the content is submitted immediately through the same flow as Enter and the terminal is cleared. Ctrl+C clears non-empty prompt text first; Ctrl+C on an empty prompt runs `/cancel`, `/return`, or `/quit` according to the active screen. Up/down move the prompt cursor within multiline prompt text, except in selectable lists when the prompt is empty.

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

After every `cli` code change, run `make lint` from `cli` and fix all findings before handoff. Treat lint rewrites such as import formatting as part of the intentional code change.

Model tests should cover startup state, screen transitions, command parsing, async message handling, and cancellation paths.

Rendering tests should assert stable output for important strings, status bands, narrow widths, and no accidental `Make a Change` copy in regular UI.

API client tests should use HTTP test servers and must not inspect database tables directly.

Markdown parsing tests should cover valid generated test cases, invalid markdown, missing titles, unsupported type values, and editor round trips.

Config tests should cover missing files, malformed files, command-line overrides, saved backend URL, saved project ID, and invalid saved project repair.

## Follow-Up Work

After the Hello World baseline, add the real Change planning workflow, backend API integration, markdown validation, editor handoff, and a retirement plan for `cli-proto/`.
