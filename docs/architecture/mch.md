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
  internal/agent/
  internal/projects/
  internal/changes/
  internal/epics/
  internal/testcases/
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
- `internal/agent`: own AI-assisted Change planning flow state, commands, navigation, rendering, Codex process orchestration, temporary planning files, idea and spec artifact handling, future agent documentation-writing flows, future agent implementation flows, future review flows, and a section-local `api.go` interface.
- `internal/projects`: own project list/detail/create/update/delete state, project commands, project navigation, project rendering, and a section-local `api.go` interface used by the app.
- `internal/changes`: own change list/detail/create/update/delete state, filters, change commands, change navigation, change rendering, and a section-local `api.go` interface.
- `internal/epics`: own epic list/detail/create/update/delete state, epic commands, epic navigation, epic rendering, and a section-local `api.go` interface.
- `internal/testcases`: own test case detail/create/update/delete state, test case commands, test case navigation, test case rendering, and a section-local `api.go` interface.
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

AI-assisted Change planning flows should use these states:

- `ready`: project context is valid and the app is ready for a planning command.
- `project selection`: no current project is selected or the saved project is invalid.
- `idea entry`: the user is entering or refining a Change idea.
- `AI running`: an async AI command is active and progress metadata is visible.
- `parse recovery`: the edited idea could not produce `# <title>` and is waiting for `/edit` or `/cancel`.
- `create confirmation`: a parsed idea is waiting for `Create Change?` Yes or No.
- `rewrite`: Codex is rewriting a created or updated idea before the final agent-edit save.
- `error`: recoverable failure with a visible reason and next action.
- `done`: the planned Change has been saved or the flow has exited cleanly.

Slash commands should be accepted only in states that define them. Unknown commands should leave user input intact and show a recoverable error.

## Navigation Shell

The navigation shell starts in `MainState` and renders deterministic screen names as header context so state transitions are easy to test. Initial render shows `MainScreen` in the right side of the header.

Top-level commands from `MainState`:

- `/changes` opens `ChangesListState` and renders `ChangesListScreen` as the right header context.
- `/epics` opens `EpicsListState` and renders `EpicsListScreen` as the right header context.
- `/projects` opens `ProjectsListState` and renders `ProjectsListScreen` as the right header context.
- `/select-project` opens `SelectProjectDropDown`, loads projects through `POST /api/v1/project/list`, saves the selected current project in TUI state, writes its numeric `project_id` to `.config/config.yaml`, and returns to `MainState`.
- `/help` opens `MainHelpState`; `/find` opens `FindInput`; `/return` returns to `MainState`.
- `/quit` exits only from `MainState`.

Slash commands, list item selection, backend selectors, confirmations, and text search should use one shared dropdown or input interaction model where practical. Users can filter dropdown options, move the highlighted option with arrow keys, and confirm the highlighted option. Every screen renders a one-line header with `Make a change v<version>` left-aligned and screen context right-aligned. Screen shortcut hints render at the left of the footer before status, current project, and color cells. Confirmation dropdowns render `Are you sure?`, expose `/yes` and `/no`, show `<return> select | <esc> or <ctrl+c> cancel` in the footer, and treat Ctrl+C the same as `/no`. Command dropdowns are overlays: opening the command list with `/` must preserve the active state and screen context while rendering commands below the page content. Selector dropdowns load options when opened, display recoverable errors when loading fails, and preserve the previous state on cancel.

Changes navigation includes `ChangesListState`, `CreateIdeaState`, `UpdateIdeaState`, `RewriteIdeaState`, `ChangeDetailsState`, `TestCaseDetailsState`, Change create and update states, Test Case create and update states, filter overlays, help, find input, and delete confirmation states. Create-state commands are context-specific: `/new-change`, `/new-testcase`, `/new-epic`, or `/new-project`; internal state names still keep CRUD-style `CreateState` suffixes. `/new-change` is available only from `ChangesListState`, not from `MainState`, and starts the AI-assisted idea rewrite flow. The Change detail spec update command is named `/edit-spec` even though the internal state name keeps a CRUD-style `UpdateState` suffix. Create and update states use their screen name as the right-aligned header context, such as `ChangeCreateScreen` or `ChangeUpdateScreen`. `ChangesListState` exposes exactly `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order; `/phase-filter`, `/epic-filter`, `/type-filter`, and `/find-filter` remain inside `ChangesListState` and must not be modeled as separate states or screens. Phase, epic, and type filter option lists render normal options with a leading `-`, such as `-<option-slug>`, and append `/clear` as the final item to remove only that field's filter. `ChangeDetailsState` exposes exactly `/reference`, `/new-testcase`, `/phase`, `/epic`, `/types`, `/edit-spec`, `/delete`, and `/return` in that order. Test case detail commands include new test case, edit, delete, save, cancel, and return. Change phase selectors render as `SelectChangePhasesDropDownScreen`; Change type selectors render as `SelectChangeTypesDropDownScreen`. `mch` loads phase and type options at startup from `POST /api/v1/options/change-phases-list` and `POST /api/v1/options/change-types-list`; it must not hardcode allowed phase or type slugs. Epic selectors and `/epic-filter` load from `POST /api/v1/epic/list` using the current project ID as a numeric JSON value.

Change list and detail screens should use backend-provided `ref` and `slug` as read-only identity data when assigned and render unassigned identity without deriving it locally. Change create and edit states must not prompt for, submit, or locally derive `ref`, `slug`, or project reference counters.

Entering a backend-backed list or detail screen must refresh its data with the relevant backend API call instead of trusting stale local state. After any successful edit, create, delete, or focused field update, `mch` must reload the destination screen data from the backend before rendering the updated state.

`ChangesListState` loads Changes from `POST /api/v1/change/list` with the current numeric `project_id` every time the user opens `/changes` and displays rows in the backend response order. It renders a boxed, scrollable selectable table with columns `#Ref`, `Phase`, `Types`, `Epic`, `Title`, `Don`, `Tot`, `%`, and `Modified`, in that order. Numeric Change refs render as six digits with leading zeroes and no `#` in row values, such as `000003`; an unassigned ref renders as a neutral empty value. Empty `change_types` renders as an empty Types cell. `Phase` values render in a 10-character column using color metadata from the loaded phase option when present; missing color metadata uses the local fallback palette backlog `15`, staging `12`, progress `10`, rejected `9`, production `13`, and review `11`, with unknown phases neutral grey. `Types` values are at most 30 characters wide, `Epic` values are at most 20 characters wide, and `Title` values are at most 80 characters wide; longer values truncate at that position without a suffix. Title values render pure white. The table renders at its natural column width when the terminal is wide enough and shrinks columns only when the available terminal width is smaller than that natural table width. `Don`, `Tot`, and `%` show done test cases, total test cases, and completed percentage from the backend response. `%` values render bright cyan. `Modified` renders as `YYYY-MM-DD HH.MM`; missing or invalid timestamps render as `not a date`. `ChangesListScreen` renders the app title/header on the first line, then a second-line filter summary right-aligned to the table width, then the table with no blank line between filters and table. The filter summary is `/filter-phase <value>   /filter-type <value>   /filter-epic <value>   /filter-find <value>`; labels render muted and only values render pure white. Its footer hint is `<ctrl+n> new change | <return> view | </> command`; Ctrl+N behaves the same as selecting `/new-change` from the command menu. Ctrl+F behaves the same as selecting `/find-filter`. Up and down arrows move the selected row within bounds, PgUp and PgDown move by one visible page, and navigation keeps the selected row inside the visible table viewport. Enter or Return loads the selected Change through `POST /api/v1/change/get` before opening `ChangeDetailsState`. List load failures show a recoverable error.

When `/new-change` starts from `ChangesListState`, `mch` must require a valid numeric current project ID before opening any editor, running Codex, or creating a Change. The planning workspace is `/tmp/mch`: if that path is a regular file, `mch` removes it and creates a directory; if `/tmp/mch/initial-idea.md` exists and is non-empty, the user chooses `/resume` to edit the existing idea, `/clear` to replace it with an empty file, or `/cancel` to return to the list. This resume prompt is part of `CreateIdeaState`, so it must read `/tmp/mch/initial-idea.md` and render that idea markdown above `Resume idea?`. When the idea file does not exist or exists but is empty, the flow takes the `/clear` path without showing the resume menu. Idea entry and focused idea update use the same full-screen editor handoff as the rest of `mch`.

After the first idea editor exits, the first line must be exactly `# <title>` with a non-blank title. If the title cannot be parsed, `mch` shows the edited idea markdown first, then a blank line, then `error parsing title:` with `/edit` and `/cancel`; `/edit` reopens the idea editor with the edited content, and `/cancel` removes `/tmp/mch/initial-idea.md` and returns to `ChangesListScreen`. When the title is valid, `mch` shows the edited idea markdown first, then a blank line, then prompts `Create Change?` with Yes and No choices in the bottom prompt area. Selecting No removes `/tmp/mch/initial-idea.md` and returns to `ChangesListScreen` without calling create. Selecting Yes first calls `POST /api/v1/change/create` with numeric `project_id`, parsed title, and idea only, then enters `RewriteIdeaState`. The idea preview must keep the raw Markdown syntax intact and apply nano-style foreground syntax coloring without rendering Markdown into blocks, frames, or filled backgrounds.

`RewriteIdeaState` resolves the repository root with `git rev-parse --show-toplevel`, runs Codex from that root with JSON event output and the prompt `Use $change-idea-tmp.`, writes JSON events to `/tmp/mch/codex-run.jsonl`, and writes final text output to `/tmp/mch/codex-output.txt`. `AgentRunningScreen` shows an animated loader with elapsed seconds while Codex runs and displays available Codex command output above the prompt in a human-readable JSON format. The flow must extract a Codex session ID from `thread_id` on the first `thread.started` JSON event, equivalent to `jq -r 'select(.type=="thread.started") | .thread_id'`, with older `session_id`, `session.id`, and `id` shapes supported only as fallbacks. The final output must be exactly `Done.`. Missing session ID or any other final output shows a verbose recoverable error and keeps the formatted command output visible. After a successful rewrite, `mch` saves the rewritten idea through `POST /api/v1/change/update-idea-agent-edit`, removes `/tmp/mch/initial-idea.md`, reloads the Change through `POST /api/v1/change/get`, and routes to `ChangeDetailsState`. Backend failures show a recoverable error and must not display local-only success state.

`ChangeDetailsState` renders `ChangeDetailsScreen` as the header context and details from the backend Change response. Details render as a scrollable two-column table with no header: labels are right-aligned in the first column and values render in the second column. Rows appear in this order: Ref, Slug, Phase, Epic, Types, Title, Idea, Spec, linked test case rows, PR, PR URL, Agent Edit, Complete, Open, Created, and Modified. All rows are selectable. Unassigned Ref and Slug render as neutral empty values. Phase renders with the same phase colors used by `ChangesListState`; the title value renders bright white. Idea, Spec, and PR values are rendered from `idea`, `spec`, and `pr_body`; Spec and PR markdown render from backend-sanitized HTML where available and truncate to 15 visible lines plus `...` when longer. Test case rows render no section label: the first column shows U+2705 when `done` is true or U+274C when false, right-aligned, and the value column renders `<scenario> (#<id>)`. The literal label `Test Cases` must not appear on this screen. `Agent Edit` renders green U+2714 when true and red U+2718 when false. `Open` renders U+2705 when true and U+274C when false. Dividers separate Types from Title, Title from Idea, Spec and test cases from PR, and PR from PR URL. Created and Modified render as `YYYY-MM-DD HH.MM`.

Enter or Return on editable detail rows starts the focused edit flow for that field. `ChangeDetailsScreen` footer hint is `<ctrl+n> new testcase | <return> edit | <space> toggle | <del> delete | </> command`; Ctrl+N behaves the same as selecting `/new-testcase` from the command menu. `/reference` calls `POST /api/v1/change/reference`, refreshes details from `POST /api/v1/change/get`, and then reconciles the Git branch for the returned `changes/<slug>`. Phase and Epic open selectors with the current value highlighted; Phase selector options render with a leading `-`, and Epic appends `@none` to clear the epic. Types opens a toggle selector in backend option order with the hint `press <space> to change`, using `+` for unselected types and `-` for selected types. In the type selector, Space toggles the highlighted type in local pending selector state and keeps the selector open. Return leaves the selector; if the pending type set changed, `mch` sends the sorted type slug list through `POST /api/v1/change/update-change-types` and reloads the Change through `POST /api/v1/change/get`; an empty selected set is valid and persists as an empty array. If nothing changed, it returns to details without a mutation request. Canceling the type selector never persists pending toggles. Title opens `ChangeUpdateState` with a title prompt. Spec, PR, PR URL, and test case scenario prompts show `<return> save | <ctrl+c> delete prompt | <esc> cancel` in the footer and persist through `POST /api/v1/change/update-spec`, `POST /api/v1/change/update-pr-body`, `POST /api/v1/change/update-pr-url`, and test case endpoints. Editing Idea opens `UpdateIdeaState`; after the editor exits, `mch` first parses `# <title>`. If parsing fails, `mch` shows the edited idea markdown first, then a blank line, then `error parsing title:` with `/edit` and `/cancel`; `/edit` reopens the idea editor with the edited content, and `/cancel` removes `/tmp/mch/initial-idea.md` and returns to `ChangesListState`. If parsing succeeds, `mch` calls `POST /api/v1/change/update-idea`, enters `RewriteIdeaState`, runs the Codex rewrite, saves with `POST /api/v1/change/update-idea-agent-edit`, removes `/tmp/mch/initial-idea.md`, and routes to `ChangeDetailsState`. `/new-testcase` opens a scenario prompt and saves through `POST /api/v1/test-case/create`; Return on a test case row opens a scenario prompt and saves through `POST /api/v1/test-case/update`; Delete on a test case row opens `Are you sure?`, and `/yes` deletes through `POST /api/v1/test-case/delete`. Pressing Space on the Open row toggles `open` through `POST /api/v1/change/update-open`, reloads details, and updates the icon. Pressing Space on a test case row toggles `done` through `POST /api/v1/test-case/update-done`, refreshes current test case data, and rerenders the full detail screen. Successful focused saves reload the Change through `POST /api/v1/change/get`, return to `ChangeDetailsState`, and keep the same detail row selected. Backend failures show a recoverable error and must not display local-only success state. `/edit-spec` opens spec editing for title, spec, types, and epic without allowing `ref` or `slug` edits.

`ChangeCreateState` renders `ChangeCreateScreen` as the header context and exposes `/save` and `/cancel` for non-agent create paths. Save parses an idea whose first line is exactly `# <title>` with a non-blank title, then creates the Change through `POST /api/v1/change/create` with numeric `project_id`, parsed title, and idea. It must not submit `ref`, `slug`, project reference counters, `change_phase`, `change_types`, `epic_id`, `spec`, `pr_body`, `pr_url`, `agent_edit`, or `open`. Successful create reloads the created Change from the backend before opening details. Validation or backend failures keep the user in a recoverable create state with the edited content available to fix.

`ChangeUpdateState` renders `ChangeUpdateScreen` as the header context and exposes `/save` and `/cancel` for spec edits. Opening `/edit-spec` must open the external editor with the current title, optional spec, optional types, and optional epic. Save parses a non-blank `# <title>`, then persists changed title, spec, types, and epic through `POST /api/v1/change/update-title`, `POST /api/v1/change/update-spec`, `POST /api/v1/change/update-change-types`, and `POST /api/v1/change/update-epic`. Successful update reloads the Change through `POST /api/v1/change/get` before refreshing details. If only one field changed, only that field's endpoint is called.

Change create and update allow no selected types. Type slugs are validated against the startup-loaded result from `POST /api/v1/options/change-types-list`; epic names are resolved from `POST /api/v1/epic/list` for the current project. An unparsable title or blank idea prevents any create or update API call. Missing or blank types persist as an empty `change_types` array.

`/cancel` and `Esc` from Change create return to `ChangesListState` without creating a Change. `/cancel` and `Esc` from Change update return to `ChangeDetailsState` without calling update endpoints.

Before Git branch reconciliation, `/reference` must verify that `mch` is running inside a Git repository and show a recoverable error if it is not. Reconciliation first checks for a local `changes/<slug>` branch and checks it out when present. If absent, it checks for a local `changes/<ref>-*` branch, checks it out, and renames it to `changes/<slug>`. If no local branch matches, it checks for remote `changes/<slug>` and checks it out. If absent, it checks for remote `changes/<ref>-*`, checks it out, pushes `changes/<slug>`, and removes the old remote ref only after the new remote ref succeeds. If no branch matches, it creates and checks out `changes/<slug>`. Backend failures, non-Git directories, Git command failures, branch conflicts, and remote rename failures are recoverable and must not display stale local success state.

Changes filters are list-local. `/phase-filter` and `/type-filter` use the startup-loaded phase and type options from the backend; `/epic-filter` loads epics from `POST /api/v1/epic/list` with numeric `project_id`. Each filter overlay remains on `ChangesListState`, keeps the list title visible, and appends `/clear` as the final option to clear only that filter field. Reopening `/phase-filter` or `/type-filter` highlights the currently applied filter when one exists. `/find-filter` applies text filtering to the loaded list by title, `ref`, `slug`, phase, type, epic, and loaded idea or spec text when present; reopening it preloads the current find text for editing. `/clear-filters` clears phase, type, epic, and find filters, restores the unfiltered loaded list, and makes the next `/find-filter` prompt blank. A filter with no matches renders a no-results state.

Epics and Projects use the same state shape: list, detail, create, update, delete confirmation, help, find input, and return. List and detail screens may navigate to new epic or new project, edit, delete, help, find, and return states according to the commands available on each screen.

`ProjectsListState` loads fresh project data from `POST /api/v1/project/list` every time the user arrives at the screen; project rows are not served from a cache. `ProjectsListScreen` renders projects as a selectable table with columns `id`, `Name`, `Changes`, `Created`, and `Modified`; ID and Changes values are right-aligned, and IDs display without a leading `#`. The Name column width is derived from the longest rendered project name. Names longer than 80 characters are normalized to single spaces, trimmed by removing whole words from the right until the rendered name plus `...` is shorter than 78 characters, then rendered with the `...` suffix. The created and modified timestamps are displayed in the current local timezone as `YYYY-MM-DD HH:mm`. Up and down arrows move the selected row within bounds when the prompt is empty, enter opens `ProjectDetailsState` with the selected project data, and this list selection must not update or persist the current project context. Pressing `/` from the list opens the command menu overlay while preserving the project list screen context underneath it.

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

Markdown parsing tests should cover valid idea/spec edit content, invalid markdown, missing titles, unsupported type values, and editor round trips.

Planning tests should fake editor and Codex execution, cover `/tmp/mch` setup, `/resume`, `/clear`, and `/cancel` idea selection, unparsable idea titles with `error parsing title:`, `/edit`, and `/cancel`, idea previews before parse errors and confirmations, `Create Change?` No as a no-op, `Create Change?` Yes create-before-rewrite behavior, update-before-rewrite behavior, rewrite output validation, Codex session capture, `update-idea-agent-edit` saves, create and save failure recovery, `/reference` branch paths, and refreshed Change detail navigation.

Config tests should cover missing files, malformed files, command-line overrides, saved backend URL, saved project ID, and invalid saved project repair.

## Follow-Up Work

Retire `cli-proto/` when the reference `cli/` module fully covers the planning, backend API, markdown validation, editor handoff, and verification behavior.
