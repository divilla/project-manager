# `mch` TUI Architecture

## Purpose

`mch` is the Go terminal UI for planning Changes. The formal app name is `Make a Change`, but product documentation, UI labels, command examples, specs, tests, and executable references use `mch` unless an approved about or version view explicitly needs the formal name.

The first executable version is `0.1`. The executable name is `mch`.

## Libraries

`mch` uses:

- Bubble Tea for the application loop, model updates, messages, and commands
- Bubbles for reusable terminal controls such as textarea-backed prompt input, viewport, spinner, and list behavior
- Lip Gloss for rendering styles and layout tokens
- `github.com/ridgelines/go-config` for loading repository YAML configuration

## Package Boundaries

Recommended layout:

```text
cli/
  cmd/mch/
  pkg/client/
  internal/dto/
  internal/app/
  internal/agent/
  internal/flow/
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
- `internal/app`: own startup wiring, config, version output, the root Bubble Tea model, product composition of reusable Flow behavior, and dispatch between feature packages. It must not own feature-specific table rendering or section navigation rules.
- `internal/agent`: own Main Screen commands, initial navigation state, and title rendering.
- `internal/flow`: own reusable Flow definitions, validation, runtime state and Screens, document canonicalization, artifact workspaces and persistence boundaries, and Editor, Preview, Diff, Exec, and Chat operations.
- `internal/projects`: own project list/detail/create/update/delete state, project commands, project navigation, project rendering, and a section-local `api.go` interface used by the app.
- `internal/changes`: own change list/detail/create/update/delete state, filters, change commands, change navigation, change rendering, and a section-local `api.go` interface.
- `internal/epics`: own epic list/detail/create/update/delete state, epic commands, epic navigation, epic rendering, and a section-local `api.go` interface.
- `internal/testcases`: own test case detail/create/update/delete state, test case commands, test case navigation, test case rendering, and a section-local `api.go` interface.
- `internal/help`: own help screen state, commands, navigation, and rendering.
- `internal/navigation`: own shared state names, screen titles, return/cancel/delete route targets, and cross-section command assembly.
- `internal/ui`: own reusable terminal UI primitives such as dropdowns, input bands, layout helpers, and generic table helpers.
- `internal/styles`: define Lip Gloss style tokens and shared components.

Each feature package must split code by responsibility instead of using one catch-all file. Use `model.go` for feature state/data methods, `navigation.go` or `update.go` for key and route decisions, `view.go` or `screen.go` for rendering, `commands.go` for slash commands, and `api.go` for the app-facing API interface. Feature packages should not import `internal/app`; the root app should call feature packages.

Runtime configuration is loaded from the current Git repository root. `mch` must resolve that root before loading `.mch/config.yaml` and `.mch/default`; starting from the repository root or any nested directory must use the same config files.

## Model And Commands

The root Bubble Tea `Model` owns current screen, window size, command menu state, current project context, visible errors, and reusable component models. It should delegate screen-specific decisions to focused helpers rather than embedding full workflows in one method.

`Update` should only translate messages into state changes and `tea.Cmd` values. It should not perform HTTP requests, file writes, editor launches, or AI calls directly.

`tea.Cmd` functions should wrap asynchronous work and return typed messages. Backend API calls and long-running AI calls must be cancellable through `context.Context` where possible. A running AI call should update the UI through loading messages and then return either a structured result message or an error message.

`View` should render current state from model data only. Rendering must not mutate state, read files, call APIs, or start processes.

## Reusable Flow Runtime

The reusable Flow runtime accepts a completed Flow definition and Flow context through a composition boundary. A Flow definition is immutable static behavior with YAML-representable identifiers and options for Steps, Tasks, Screens, artifacts, prompts or scripts, expected output, commands, and typed destinations. A Flow context owns runtime-only state: active Change identity, originating navigation Screen, current session, Step and Task position, Preview `from-step`, artifact, and execution result. Go conformance definitions use these same types and must return fresh independently mutable values.

Definitions support exactly the artifacts `idea`, `spec`, `pr`, `implement`, `review`, and `finalize`; generic Editor, Exec, Chat, Preview, and Error Screens; and validated ordered Task sequences. A Step is a non-executable container for `1..n` Tasks, and its Mode determines the allowed Task types and order: Editor Mode contains exactly `EditorTask`, Chat Mode contains exactly `ChatTask`, and Exec Mode contains exactly `ExecTask -> ChatTask`. Script Mode and additional Task shapes are not supported. A shared validator rejects missing or duplicate identifiers, unsupported kinds or sequences, inconsistent Step artifacts, invalid fields, and invalid destinations before execution. A typed Step destination activates the first ordered Task; a typed Screen destination names a terminal navigation Screen in the composition-supplied allowlist.

Editor Tasks require an artifact, Editor Screen, Preview completion, and Error failure destination. Exec Tasks additionally require a prompt, script, exact expected output, and the following same-Step Chat fallback. Chat Tasks require an artifact, Chat Screen, session-resume script, Editor destination, Preview completion, and Error destination. Preview records the `from-step` whose Task entered it. `/continue` follows its typed destination, `/edit` selects a same-artifact Editor Step and records the caller, and `/cancel` returns to `ChangeDetailsState`. An Exec- or Chat-mode `from-step` inserts mode-derived `/chat` between `/continue` and `/edit`; it activates that same Step's `ChatTask` without selecting or restarting a Step. Editor-mode Preview does not expose `/chat`.

Composition supplies the validated definition, Flow context, allowed terminal Screens, active Flow directory for relative paths, an `ArtifactStore`, and fakeable editor, preview, diff, execution, chat-session, and API boundaries. The current product composition allows terminal `MainState`, `ChangesListState`, and `ChangeDetailsState` and provides a fresh built-in Idea Stage definition for `IdeaEdit`, `IdeaRewriteExec`, and `IdeaReviewExec`. The Exec names identify Steps, not their first `ExecTask`; their second Tasks use `IdeaRewriteChat` and `IdeaReviewChat`. Product-specific destinations and expected output remain in the definition rather than generic runtime branches. Loading that definition from `.mch/default/flow.yaml`, Spec work, and later stages are not active.

### Artifact Workspace And Step Lifecycle

`TmpDir` is executable source code's sole temp-root value and equals `.mch/tmp`. Startup requires `<git-root>/.mch` and `<git-root>/.mch/tmp` to exist as directories and never creates those roots. `IdeaCreate` atomically creates `<git-root>/<temp-dir>/<attempt-uuid>/new-idea.md`; persisted Steps may create descendants under `<git-root>/<temp-dir>/<ref-uuid>/<artifact>`. The attempt UUID is runtime correlation identity and is not sent as `ref_uuid`. Editor-mode Steps isolate their draft pair in `<artifact>/editor/input.md` and `output.md`; `editor` is a workspace scope, not an artifact identity. Go passes `<temp-dir>` as `MCH_TEMP_DIR`, while execution and resume scripts require `MCH_TEMP_DIR`, `MCH_REF_UUID`, and `MCH_ARTIFACT` and compose the artifact root containing `session-id`, shared Preview/Exec/Chat `input.md` and `output.md`, `agent-output.md`, `events.jsonl`, and `error.log`.

Selecting an Exec or Chat Step loads exact persisted bytes into the artifact root's byte-identical `input.md` and `output.md`, then activates its first Task. Selecting an Editor Step loads those bytes under `<artifact>/editor`, leaving the caller's artifact-root comparison intact. Every Chat entry fresh-loads the artifact-root baseline while preserving the Step session and workspace. Each scoped `input.md` is immutable for that Task entry. Changed Editor completion saves first and publishes its Editor comparison pair to the artifact root before Preview; canonical byte-identical completion returns without touching the caller pair. Preview performs no API operation. Different Changes are isolated by `ref_uuid`; the runtime assumes at most one active Flow process per Change and adds no cross-process lock.

Every non-empty Idea, Spec, or PR submission normalizes CRLF to LF before validating a one-line `# <title>`, exact blank-line separators, optional artifact-local metadata, and a body containing at least one non-whitespace character. Title-only, whitespace-body, and metadata-only documents fail before option loading or persistence. A present `Types:` uses unique slugs validated against `POST /api/v1/options/change-types-list` and is emitted in API order as `Types: <type-slugs>`. A present `Epic:` resolves through the current project's `POST /api/v1/epic/list` response and is emitted as `Epic: <epic-title> #<epic-id>`. Canonical order is H1, Types, Epic, each followed by one blank line. Missing metadata remains missing.

Canonical output is compared byte-for-byte with the fresh input. A changed Editor submission saves before Preview with `agent_edit: false`; changed Exec or Chat output saves before routing with `agent_edit: true`. Canonical byte-identical Editor output makes no save and returns to the recorded calling terminal, Preview, or Chat Screen. Empty Editor output cancels to `ChangeDetailsState`; invalid non-empty Editor output exposes only `/fix` and `/cancel`. Artifact saves never mutate Change title, types, Epic, ref, or slug. Stopped, cancelled, failed, or unsuccessfully saved Tasks do not complete or persist.

The CLI Change artifact store loads the active Change through `POST /api/v1/change/get` and persists `idea`, `spec`, and `pr` through their matching focused update endpoints. Save provenance is supplied explicitly rather than inferred from artifact or Step names. Unsupported artifacts are rejected instead of being routed to an unrelated endpoint.

### Generic Flow Screens

- Editor opens `<artifact>/editor/output.md`. Changed canonical output saves and publishes its comparison pair to the artifact root before Preview; unchanged canonical output returns to the recorded caller without replacing its Preview/Diff files; empty output and `/cancel` return to details.
- Exec offers only `/stop` while running. It validates and saves output before evaluating readable `agent-output.md`. A matching last line skips the following Chat position; an empty or different last line advances to Chat inside the same Step. Missing or unreadable output enters Error. `/stop` returns to the origin without a pending save.
- Chat fresh-loads persisted artifact bytes and shows available agent output without starting a session. It offers `/chat`, `/edit`, and `/cancel`. `/chat` resumes through the configured command and non-empty `session-id`; the interactive transcript is terminal-only. `/edit` may select a configured same-artifact Editor Step and records the Chat caller; `/cancel` returns to the origin.
- Preview performs no load or save and supports prepared Idea, Spec, and PR files. It renders `output.md` with `bat -pp --theme 'Coldark-Dark'` and renders Diff with `git --no-pager diff --no-index --no-ext-diff --color=never` piped to `bat` in diff mode. Either horizontal arrow toggles modes. Git status `0` means identical, `1` means different, and greater than `1` is Error; the runtime must preserve Git's status rather than the piped renderer's. Ordinary commands map explicitly to `step` or `screen` destinations. `/continue` follows its configured typed destination: a `step` destination activates that Step's first Task, while a `screen` destination ends the Flow at the allowed terminal Screen. Mode-derived `/chat` instead activates the `ChatTask` owned by Preview's recorded `from-step`, and only an Exec- or Chat-mode `from-step` may expose it.
- Error displays the concrete validation, workspace, load, save, editor, execution, session, or rendering failure and offers only `/return`, which returns to the originating Screen without retrying, saving, or continuing.

External operations return typed messages to the Bubble Tea update loop and preserve terminal handoff, cancellation, process status, session resources, and configured working directories. `Update` must not directly perform filesystem, persistence, or process work.

## Planning States

The composed Idea Stage uses these state roles:

- `ready`: project context is valid and the app is ready for a planning command.
- `project selection`: no current project is selected or the saved project is invalid.
- `idea entry`: the user is entering or refining a Change idea.
- `idea processing`: Idea validation is in flight, a one-character spinner is visible, and `/cancel` is the only command.
- `AI running`: an async AI command is active and progress metadata is visible.
- `validation recovery`: non-empty Editor content is invalid and is waiting for `/fix` or `/cancel`.
- `create confirmation`: a parsed idea is waiting for `Create Change?` Yes or No.
- `rewrite` and `review`: the active Exec-mode Step is traversing its ordered `ExecTask -> ChatTask` sequence.
- `chat`: agent output is visible and the resumable session is waiting for `/chat`, `/edit`, or `/cancel`.
- `preview`: canonical saved output is waiting for its mode-aware next command.
- `error`: recoverable failure with a visible reason and next action.
- `done`: the planned Change has been saved or the flow has exited cleanly.

Slash commands should be accepted only in states that define them. Unknown commands should leave user input intact and show a recoverable error.

## Navigation Shell

The navigation shell starts in `MainState` and renders deterministic screen names as header context so state transitions are easy to test. Initial render shows `MainScreen` in the right side of the header.

Top-level commands from `MainState`:

- `/changes` opens `ChangesListState` and renders `ChangesListScreen` as the right header context.
- `/epics` opens `EpicsListState` and renders `EpicsListScreen` as the right header context.
- `/projects` opens `ProjectsListState` and renders `ProjectsListScreen` as the right header context.
- `/select-project` opens `SelectProjectDropDown`, loads projects through `POST /api/v1/project/list`, saves the selected current project in TUI state, writes its numeric `project_id` to `.mch/config.yaml`, and returns to `MainState`.
- `/config` opens a read-only `Config` view from resolved in-memory configuration and returns to `MainState` with `/return`, Esc, or Ctrl+C on an empty prompt.
- `/help` opens `MainHelpState`; `/find` opens `FindInput`; `/return` returns to `MainState`.
- `/quit` exits only from `MainState`.

Slash commands, list item selection, backend selectors, confirmations, and text search should use one shared dropdown or input interaction model where practical. Users can filter dropdown options, move the highlighted option with arrow keys, and confirm the highlighted option. Every screen renders a one-line header with `Make a change v<version>` left-aligned and screen context right-aligned. Screen shortcut hints render at the left of the footer before status, current project, and color cells. Confirmation dropdowns render `Are you sure?`, expose `/yes` and `/no`, show `<return> select | <esc> or <ctrl+c> cancel` in the footer, and treat Ctrl+C the same as `/no`. Command dropdowns are overlays: opening the command list with `/` must preserve the active state and screen context while rendering commands below the page content. Selector dropdowns load options when opened, display recoverable errors when loading fails, and preserve the previous state on cancel.

Changes navigation includes `ChangesListState`, `ChangeDetailsState`, generic Idea Flow Screens, `TestCaseDetailsState`, Change create and update states, Test Case create and update states, filter overlays, help, find input, and delete confirmation states. `/new-change` is available only from `ChangesListState`, not from `MainState`, and starts `IdeaCreate`; Return on the Idea detail row starts `IdeaEdit`. `ChangesListState` exposes exactly `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order. `ChangeDetailsState` exposes exactly `/new-testcase`, `/phase`, `/epic`, `/types`, `/edit-spec`, `/delete`, and `/return` in that order. It does not expose Flow assignment, Run controls, claim reset, or branch reconciliation commands. Filters remain list-local. Phase and type options come from their backend option endpoints, and Epic selectors use `POST /api/v1/epic/list` with numeric current `project_id`.

Change list and detail screens should use backend-provided `ref_uuid`, `ref`, and `slug` as read-only identity data when present and render unassigned `ref` or `slug` without deriving it locally. Change create and edit states must not prompt for, submit, or locally derive `ref_uuid`, `ref`, `slug`, or project reference counters.

Entering a backend-backed list or detail screen must refresh its data with the relevant backend API call instead of trusting stale local state. After any successful edit, create, delete, or focused field update, `mch` must reload the destination screen data from the backend before rendering the updated state.

`ChangesListState` loads Changes from `POST /api/v1/change/list` with the current numeric `project_id` every time the user opens `/changes` and displays rows in the backend response order. It renders a boxed, scrollable selectable table with columns `#Ref`, `Phase`, `Types`, `Epic`, `Title`, `Don`, `Tot`, `%`, and `Modified`, in that order. Numeric Change refs render as six digits with leading zeroes and no `#` in row values, such as `000003`; an unassigned ref renders as a neutral empty value. Empty `change_types` renders as an empty Types cell. `Phase` values render in a 10-character column using color metadata from the loaded phase option when present; missing color metadata uses the local fallback palette backlog `15`, staging `12`, progress `10`, rejected `9`, production `13`, and review `11`, with unknown phases neutral grey. `Types` values are at most 30 characters wide, `Epic` values are at most 20 characters wide, and `Title` values are at most 80 characters wide; longer values truncate at that position without a suffix. Title values render pure white. The table renders at its natural column width when the terminal is wide enough and shrinks columns only when the available terminal width is smaller than that natural table width. `Don`, `Tot`, and `%` show done test cases, total test cases, and completed percentage from the backend response. `%` values render bright cyan. `Modified` renders as `YYYY-MM-DD HH.MM`; missing or invalid timestamps render as `not a date`. `ChangesListScreen` renders the app title/header on the first line, then a second-line filter summary right-aligned to the table width, then the table with no blank line between filters and table. The filter summary is `/filter-phase <value>   /filter-type <value>   /filter-epic <value>   /filter-find <value>`; labels render muted and only values render pure white. Its footer hint is `<ctrl+n> new change | <return> view | </> command`; Ctrl+N behaves the same as selecting `/new-change` from the command menu. Ctrl+F behaves the same as selecting `/find-filter`. Up and down arrows move the selected row within bounds, PgUp and PgDown move by one visible page, and navigation keeps the selected row inside the visible table viewport. Enter or Return loads the selected Change through `POST /api/v1/change/get` before opening `ChangeDetailsState`. List load failures show a recoverable error.

Current-project selection is a navigation precondition. Without a valid positive numeric current project, the user remains in project selection and cannot reach `ChangesListState` or `ChangeDetailsState`. The Idea Stage therefore receives valid project context from either entry Screen and does not duplicate that guard inside the Flow.

Every `/new-change` generates a UUID v4, atomically creates `<git-root>/<temp-dir>/<attempt-uuid>` without a pre-check, creates zero-byte `new-idea.md`, and opens it in Editor. An existence collision retries with another UUID. It never offers `/resume`, `/clear`, or prior-draft recovery. Zero-byte output returns directly to `ChangesListScreen` without option or create API calls. Invalid non-empty content exposes `/fix` and `/cancel`; `/fix` reopens the same attempt and `/cancel` returns to the list. Valid content is canonicalized before `Create Change?`. No makes no create request; Yes creates from the displayed canonical bytes, initializes the Change title from this Idea H1, and starts `IdeaRewriteExec` with `ChangesListState` as origin. Validation and create replies must match the active attempt UUID. Completion, abandonment, or terminal failure removes `new-idea.md` and then removes only an empty attempt directory; a non-empty directory and its other files are preserved.

`IdeaRewriteExec` and `IdeaReviewExec` are Exec-mode Steps composed as `ExecTask -> ChatTask`. Rewrite expects `Done.` and Review expects `No questions or suggestions.` as the exact last line of readable `agent-output.md`. Matching output skips Chat; an empty or different readable line advances to the same Step's Chat Screen. Exec and Chat validate and save changed canonical Idea bytes with `agent_edit: true` before routing. Rewrite Preview `/continue` starts Review; Review Preview `/continue` temporarily exits to `MainState` without Spec work. Both agent Previews add `/chat` between `/continue` and `/edit`, and every `/cancel` returns to `ChangeDetailsState`.

`IdeaEdit` is the generic `DocEdit` configuration entered from the Idea row or from Preview or Chat `/edit`. It fresh-loads the persisted Idea into byte-identical `idea/editor/input.md` and `output.md`, records its calling Screen, and opens the Editor output without replacing the caller's artifact-root files. Changed valid canonical output saves with `agent_edit: false`, publishes its comparison pair to the artifact root, and enters Editor-mode Preview; only `/continue` starts Rewrite. Empty output returns to details. Canonical byte-identical output makes no save and restores the caller with its prior Preview/Diff comparison intact. Invalid non-empty output remains in `/fix` or `/cancel` recovery. Starting `IdeaEdit` from Preview or Chat abandons the unpersisted local proposal.

`ChangeDetailsState` renders `ChangeDetailsScreen` as the header context and details from the backend Change response. Details render as a scrollable two-column table with no header: labels are right-aligned in the first column and values render in the second column. ID and Ref UUID render as fixed selectable rows at the top. Scrollable rows appear in this order: Ref, Slug, Phase, Epic, Types, Title, Idea, Spec, linked test case rows, PR, PR URL, Agent Edit, Complete, Open, Created, and Modified. Detail rows are selectable. `ctrl+shift+c` or `ctrl+insert` copies the selected row value when the terminal passes the key through. Unassigned Ref and Slug render as neutral empty values. Phase renders with the same phase colors used by `ChangesListState`; the title value renders bright white. Idea, Spec, and PR values are rendered from `idea`, `spec`, and `pr`; Spec and PR markdown render from backend-sanitized HTML where available and truncate to 15 visible lines plus `...` when longer. Empty artifact strings and empty rendered artifact strings render as empty values, never as `null`. Test case rows render no section label: the first column shows U+2705 when `done` is true or U+274C when false, right-aligned, and the value column renders `<scenario> (#<id>)`. The literal label `Test Cases` must not appear on this screen. `Agent Edit` renders green U+2714 when true and red U+2718 when false. `Open` renders U+2705 when true and U+274C when false. Dividers separate Types from Title, Title from Idea, Spec and test cases from PR, and PR from PR URL. Created and Modified render as `YYYY-MM-DD HH.MM`.

Enter or Return on editable detail rows starts the focused edit flow for that field. `ChangeDetailsScreen` footer hint is `<ctrl+n> new testcase | <return> edit | <space> toggle | <del> delete | <ctrl+ins> copy | </> command`; Ctrl+N behaves like `/new-testcase`. Phase and Epic open selectors with the current value highlighted; Epic includes `@none` to clear the link. Types opens a backend-ordered toggle selector and persists changed selections through `POST /api/v1/change/update-change-types`; cancel never persists pending toggles. Title uses its explicit Change update flow. Spec and PR user saves call their focused endpoints with `agent_edit: false`, and PR URL uses its focused endpoint. Editing Idea starts the composed `IdeaEdit` Step described above rather than a specialized post-save path. Test case create, edit, delete, and done toggles use their focused endpoints. Open toggles through `POST /api/v1/change/update-open`. Successful focused saves reload destination data from the backend. Backend failures remain recoverable and must not display local-only success state. `/edit-spec` edits Spec content and metadata without allowing `ref`, `ref_uuid`, or `slug` edits.

`ChangeCreateState` renders `ChangeCreateScreen` as the header context and exposes `/save` and `/cancel` for non-agent create paths. Save parses an idea whose first line is exactly `# <title>` with a non-blank title, then creates the Change through `POST /api/v1/change/create` with numeric `project_id`, parsed title, and idea. It must not submit `ref`, `ref_uuid`, `slug`, project reference counters, `change_phase`, `change_types`, `epic_id`, `spec`, `pr`, `pr_url`, `agent_edit`, or `open`. Successful create reloads the created Change from the backend before opening details. Validation or backend failures keep the user in a recoverable create state with the edited content available to fix.

`ChangeUpdateState` renders `ChangeUpdateScreen` as the header context and exposes `/save` and `/cancel` for spec edits. Opening `/edit-spec` must open the external editor with the current Spec document. Save validates a non-blank H1 title and any present `Types:` and `Epic:` metadata, canonicalizes valid metadata, then persists changed Spec content through `POST /api/v1/change/update-spec` with `agent_edit` set to false. Parsed document metadata does not update the Change title, types, or linked Epic. Successful Spec update reloads the Change through `POST /api/v1/change/get` before refreshing details.

Change create and update allow no selected types. Type slugs are validated against the startup-loaded result from `POST /api/v1/options/change-types-list`; epic names are resolved from `POST /api/v1/epic/list` for the current project. An unparsable title or blank idea prevents any create or update API call. Missing or blank types persist as an empty `change_types` array.

Every Idea, Spec, and PR submission replaces every CRLF sequence with LF before validation, comparison, or persistence, so canonical artifact output uses LF throughout. It then validates and canonicalizes artifact-local metadata before saving. Types render as `Types: <type-slugs>` with API-backed slugs joined by `|` and no spaces. Epic renders as `Epic: <epic-title> #<epic-id>` using the canonical title and ID from `POST /api/v1/epic/list` for the current project. Artifact metadata may differ from the Change's Types and Epic and must not call `POST /api/v1/change/update-change-types` or `POST /api/v1/change/update-epic`; only pressing Return on the corresponding `ChangeDetailsScreen` row changes those Change fields. Validation or canonicalization failure must remain recoverable and must not persist the document or present processing as complete.

`/cancel` and `Esc` from Change create return to `ChangesListState` without creating a Change. `/cancel` and `Esc` from Change update return to `ChangeDetailsState` without calling update endpoints.

Branch reconciliation from Change identity is out of scope for `mch` until a dedicated CLI Change adds a supported control for the backend Flow assignment endpoint.

Changes filters are list-local. `/phase-filter` and `/type-filter` use the startup-loaded phase and type options from the backend; `/epic-filter` loads epics from `POST /api/v1/epic/list` with numeric `project_id`. Each filter overlay remains on `ChangesListState`, keeps the list title visible, and appends `/clear` as the final option to clear only that filter field. Reopening `/phase-filter` or `/type-filter` highlights the currently applied filter when one exists. `/find-filter` applies text filtering to the loaded list by title, `ref`, `slug`, phase, type, epic, and loaded idea or spec text when present; reopening it preloads the current find text for editing. `/clear-filters` clears phase, type, epic, and find filters, restores the unfiltered loaded list, and makes the next `/find-filter` prompt blank. A filter with no matches renders a no-results state.

Epics and Projects use the same state shape: list, detail, create, update, delete confirmation, help, find input, and return. List and detail screens may navigate to new epic or new project, edit, delete, help, find, and return states according to the commands available on each screen.

`ProjectsListState` loads fresh project data from `POST /api/v1/project/list` every time the user arrives at the screen; project rows are not served from a cache. `ProjectsListScreen` renders projects as a selectable table with columns `id`, `Name`, `Changes`, `Created`, and `Modified`; ID and Changes values are right-aligned, and IDs display without a leading `#`. The Name column width is derived from the longest rendered project name. Names longer than 80 characters are normalized to single spaces, trimmed by removing whole words from the right until the rendered name plus `...` is shorter than 78 characters, then rendered with the `...` suffix. The created and modified timestamps are displayed in the current local timezone as `YYYY-MM-DD HH:mm`. Up and down arrows move the selected row within bounds when the prompt is empty, enter opens `ProjectDetailsState` with the selected project data, and this list selection must not update or persist the current project context. Pressing `/` from the list opens the command menu overlay while preserving the project list screen context underneath it.

`/new-project` opens `ProjectCreateState` with a project name form and the placeholder `Write a Name`. `ProjectCreateState` and `ProjectUpdateState` expose `/editor`, `/save`, and `/cancel` in that order. Saving validates that the name is non-empty after trimming, then sends the name exactly as entered, including explicit newline characters. Create sends `POST /api/v1/project/create` with only the required `name` field, then reloads the created project through `POST /api/v1/project/get` before opening `ProjectDetailsState`. `/edit` from `ProjectDetailsState` opens `ProjectUpdateState` with the current project name prefilled. Update requires a valid positive numeric project ID, sends `POST /api/v1/project/update` with numeric `id` and string `name`, then reloads the updated project through `POST /api/v1/project/get` before refreshing `ProjectDetailsState`. Create and update failures leave the user in the form with a recoverable backend or validation error, and cancel actions return without an API call.

`ProjectDetailsState` reloads the selected project through `POST /api/v1/project/get` every time the user arrives at the screen; detail rows are not served from a cache. The detail view displays labels `#ID`, `Name`, `Changes`, `Created`, and `Modified`. Labels are shifted four spaces to the right, right-aligned, and values are left-aligned. Normal values render white, the `#ID` value renders light pink, the name value renders bright cyan, and created/modified values render grey between the label grey and white. The name value wraps at 80 characters without breaking words and preserves explicit newline characters. Created and modified timestamps render in the current local timezone as `YYYY-MM-DD HH:mm`, truncating seconds and sub-second precision. Missing or invalid timestamps render as `not a date`.

`Esc` maps to the state-appropriate safe action: quit from `MainState`, return from returnable states, and cancel from create, update, dropdown, confirmation, loading selector, and input states. `/quit` outside `MainState` and unknown commands leave the current state unchanged and show a recoverable error.

Save, delete, filter, selector, and selection actions in the navigation shell must use backend APIs for persistence when implemented. They must not write directly to the database. Project create and update forms persist only project data and must not update `.mch/config.yaml` or change the current project context.

## Backend And Persistence

Backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence. `mch` must not write application database tables directly.

Project-scoped commands should either use the saved current project context or require an explicit project option. When the saved project no longer exists, `mch` should clear or repair selection using the same behavior documented for current project context.

## Config

`mch` resolves the Git repository root at startup, then loads committed repository config from `.mch/config.yaml`. The file owns `backend_url` and numeric `project_id`; it has no Flow temp-root field. `project_id` is shared branch-scoped state for everyone using that repository branch. `.mch` is project configuration, not a secrets or user-local directory, and must remain eligible to be committed.

After resolving the Git root and before starting the TUI, `mch` requires both `<git-root>/.mch` and `<git-root>/.mch/tmp` to exist as directories. A missing path or non-directory produces a path-specific startup error; startup does not create either root. Runtime Flow files use the `TmpDir` constant `.mch/tmp`. Automated tests and verification commands may separately use disposable system temporary directories.

Missing `project_id` and `project_id: 0` are valid no-selection states. When startup has no positive `project_id`, `mch` opens the same selector flow as `/select-project` from `MainState`. If the backend returns no projects, `mch` stays on `MainState` and shows `No projects to select from. Please create new project and select it on Main Screen.` It must not redirect to project creation. `/select-project` updates only `.mch/config.yaml` and saves the selected project ID as a number. Product data must be saved only through backend APIs.

`mch` always loads the active Flow profile from `.mch/default`. It parses `flow.yaml` and `help.yaml` into typed in-memory config, including Flow metadata, ordered Step slugs, Step help text, Step mode, prompt paths, exact `entry`, `exec`, and `exit` hook command strings, stage modes, task statuses, and task steps. Flow Step slugs and modes are configuration data rather than a built-in vocabulary: Step slugs must be non-empty and unique, and Step modes must be non-empty. Flow help option groups are also configuration data: listed stage modes, task statuses, and task steps only require non-empty slugs, and missing groups, empty groups, custom slugs, and reordered slugs are valid.

Repository Change workflow automation uses `change/<change-slug>` branches, ideas under `agent/ideas/<change-slug>.md`, and Change specs under `specs/<change-slug>.md`. The canonical spec structure prompt is `.mch/default/prompts/spec-file-structure.md`; `mch` and agent workflows must not depend on legacy spec-structure prompts under `agent/prompts`.

The CLI must not implement Flow profile selection, config overrides, fallback layering, `.mch/local`, or compatibility loading from old config locations. Missing, unreadable, or malformed required configuration must produce a verbose path-specific startup error without substituting another source, then `mch` exits without starting the interactive app.

`/config` renders a read-only view from resolved in-memory configuration. It shows repository root, config path, `backend_url`, `project_id`, active Flow directory, Flow metadata, Step data, hook command strings, stage modes, task statuses, and task steps; it does not display a Flow temp-root setting. Opening or leaving `/config` must not call backend APIs, read raw YAML for rendering, execute Flow hooks, or save any config file.

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

Idea Stage tests should fake editor, backend, Exec, Chat, Preview, Diff, and scripts. They cover collision-safe UUID attempt allocation, isolated fresh `new-idea.md`, UUID-correlated stale-reply rejection, two-step cleanup with non-empty-directory preservation, zero-byte cancellation, validation and canonicalization before confirmation, create-before-Rewrite, generic `IdeaEdit`, caller-aware no-op return, ordered Rewrite and Review traversal, mode-aware Preview commands, terminal-only Chat output, provenance-sensitive saves, every cancellation and error route, and absence of legacy draft and Spec-generation paths.

Config tests should cover repository-root resolution from nested directories, required `.mch` and `.mch/tmp` directories, `.mch/config.yaml` parsing without a temp-root field, `.mch/default` Flow parsing, `/config` routing and rendering, no-selection `project_id` states, saved project ID persistence, and path-specific startup failures without fallback.

## Follow-Up Work

Retire `cli-proto/` when the reference `cli/` module fully covers the planning, backend API, markdown validation, editor handoff, and verification behavior.
