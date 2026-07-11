# CLI Agent Integration

Types: feature|test|docs

## Goal

Make `/new-change` in `mch` run an AI-assisted planning flow that turns a user's Markdown idea into a standard Change file and creates the corresponding backend Change for the current project.

## Scope

- Implement the `/new-change` agent planning flow for `mch` from `ChangesListState`.
- Manage the temporary planning workspace at `/tmp/mch`, including `initial-idea.md`, Codex JSON logs, Codex output, and `initial-change.md`.
- Reuse the existing full-screen editor handoff for idea entry and review.
- Invoke Codex CLI rewrite and Change-generation prompts for the documented loop.
- Parse `/tmp/mch/initial-change.md` into a backend Change create payload.
- Create the Change and one backend test case for every generated QA Test Case through supported backend APIs, then open refreshed Change details on success.
- Add or update focused `mch` documentation and tests for the new planning flow.

## Requirements

- Selecting `/new-change` from `ChangesListState` must start the AI-assisted planning flow instead of the manual Change markdown create flow.
- The flow must use the current saved project context and must not create a Change when no valid numeric current project ID is available.
- Before idea entry, `mch` must ensure `/tmp/mch` is a directory. If `/tmp/mch` exists as a regular file, `mch` must remove that file and create the directory.
- The idea file path must be `/tmp/mch/initial-idea.md`.
- If `/tmp/mch/initial-idea.md` already exists and is non-empty, `mch` must prompt with a command menu containing `/resume`, `/clear`, and `/cancel`.
- Selecting `/resume` must open the existing idea file in the full-screen editor and must not run Codex if the user exits without changing the file.
- Selecting `/clear` must create or replace `/tmp/mch/initial-idea.md` with an empty file and open it in the full-screen editor.
- Selecting `/cancel` must return to `ChangesListScreen` without opening an editor, invoking Codex, or calling a backend create endpoint.
- If `/tmp/mch/initial-idea.md` does not exist or exists but is empty, `mch` must take the `/clear` path without showing the `/resume` menu.
- The idea editor must use the same full-screen editor mechanism as the existing CLI editor integration.
- If `/tmp/mch/initial-idea.md` is empty after the user exits the first editor, `mch` must return to `ChangesListScreen` without invoking Codex and without calling a backend create endpoint.
- For a non-empty idea, `mch` must resolve the repository root with `git rev-parse --show-toplevel`.
- The initial rewrite run must execute Codex from the repository root with JSON event output, write the final text output to `/tmp/mch/codex-output.txt`, and write the JSON event log to `/tmp/mch/codex-run.jsonl`.
- The initial rewrite run must invoke the prompt `Use $change-idea-tmp.`.
- While Codex is running, `AgentRunningScreen` must show an animated loader, elapsed seconds, and human-readable formatted Codex command output above the prompt when output is available.
- `mch` must extract the Codex session ID from `thread_id` on the first `thread.started` JSON event, matching `jq -r 'select(.type=="thread.started") | .thread_id'`, and may fall back to older `session_id`, `session.id`, or `id` shapes.
- If no Codex session ID can be extracted, `mch` must show the recoverable error `something went wrong - please try again`.
- If `/tmp/mch/codex-output.txt` is not exactly `Done.` after a rewrite run, `mch` must show the recoverable error `something went wrong - please try again`.
- After a successful rewrite run, `mch` must reopen `/tmp/mch/initial-idea.md` in the full-screen editor for user review.
- If the user saves changed content during review, `mch` must resume the same Codex exec session with `Use $change-idea-tmp.`, update `/tmp/mch/codex-output.txt`, require the output to be exactly `Done.`, and continue the review loop.
- If the user exits the review editor without saving changes, `mch` must start interactive Codex with `codex resume <session_id> 'Use $change-spec-tmp.'` using the same terminal process handoff pattern as the full-screen editor.
- After the interactive Codex process exits, `mch` must parse `/tmp/mch/initial-change.md`.
- Parsing must require the first non-blank line to be an H1 title and use that value as the Change `title`.
- Parsing must require the first non-blank line after the H1 title to be exactly `Types: <type-slugs>`, with one or more slugs joined by `|` and no spaces.
- Parsed Change types must be sent as backend `change_types`.
- The full contents of `/tmp/mch/initial-change.md` must be preserved as the backend Change `body`.
- Change create must call `POST /api/v1/change/create` with numeric `project_id`, parsed `title`, full `body`, and parsed `change_types`.
- Change create must not send `ref`, `slug`, `change_phase`, `pr`, `pr_url`, `agent_edit`, or `open`.
- After a successful Change create, `mch` must parse every scenario listed under `## QA Test Cases` in the generated Change body and call `POST /api/v1/test-case/create` with the created numeric `change_id` and each scenario as `scenario`.
- Test case creation must preserve QA Test Case order and must complete before reloading created Change details.
- If any generated QA Test Case create request fails, `mch` must show a recoverable save error and must not open a local-only Change details success state.
- Successful create must open `ChangeDetailsState` for the created Change and reload backend detail data before rendering it.
- Parse, Codex command, editor, file-system, and backend failures must leave the user in a recoverable flow state with a visible error and no local-only success state.
- The implementation must not write directly to application database tables or mutate files under `db/**`.

## Acceptance Criteria

- Selecting `/new-change` on `ChangesListScreen` prepares `/tmp/mch`, opens `/tmp/mch/initial-idea.md`, and does not enter the previous manual Change create editor flow.
- When `/tmp/mch` exists as a regular file, starting `/new-change` replaces it with a directory before opening the idea file.
- When `/tmp/mch/initial-idea.md` already exists and is non-empty, the user sees `/resume`, `/clear`, and `/cancel`; `/resume` preserves and opens the existing file, `/clear` opens a blank idea file, and `/cancel` returns to the list.
- Exiting a resumed idea editor without changes returns to `ChangesListScreen` without running Codex or calling `POST /api/v1/change/create`.
- Exiting a resumed idea editor after changing the file runs the normal rewrite flow with `Use $change-idea-tmp.`.
- Exiting the first idea editor with an empty file returns to `ChangesListScreen` without running Codex or calling `POST /api/v1/change/create`.
- Exiting the first idea editor with non-empty content runs Codex with `Use $change-idea-tmp.`, writes `/tmp/mch/codex-run.jsonl` and `/tmp/mch/codex-output.txt`, extracts a session ID from `thread.started.thread_id`, and requires output `Done.`.
- `AgentRunningScreen` displays an animated loader with elapsed seconds and formats JSON command output into readable indented blocks.
- A rewrite output other than `Done.` shows `something went wrong - please try again` and does not advance to Change creation.
- Saving edited idea content during review resumes the same Codex session with `Use $change-idea-tmp.` and reopens the review editor after successful output.
- Exiting the review editor without saving starts interactive `codex resume` with `Use $change-spec-tmp.`.
- After interactive Codex exits, a valid `/tmp/mch/initial-change.md` with an H1 title and `Types:` line is parsed into `title`, `body`, and `change_types`.
- A generated Change file missing an H1 title, missing `Types:`, using a blank type list, or using malformed type separators shows a parse error and does not call the backend.
- Successful parsing sends `POST /api/v1/change/create` with numeric `project_id`, `title`, full `body`, and `change_types`, creates one backend test case for each generated QA Test Case, then opens refreshed details for the created Change.
- A generated Change with multiple QA Test Case bullets creates matching backend test cases in the same order before details reload.
- A test case create failure after Change create shows a recoverable save error and does not show created details.
- Backend create failure shows a recoverable error and does not show the Change as created.
- The implementation and docs use `/api/v1/change/create`, not `/api/v3/change/create`.

## Non-Goals

- No direct database writes, migrations, seed changes, or edits under `db/**`.
- No new backend planning endpoints.
- No frontend SPA changes.
- No automatic epic selection or `epic_id` persistence from generated Change files.
- No changes to Change detail focused edit behavior, test case mutation behavior, or existing backend Change persistence rules.
- No support for running this flow from `MainState`; `/new-change` remains a `ChangesListState` command.

## Design Notes

- `docs/architecture/mch.md` defines AI-assisted workflow ownership under `cli/internal/agent`; use that package boundary for this planning flow and future agent documentation-writing, implementation, and review flows instead of `internal/planning`.
- `docs/architecture/backend-api.md` is authoritative for the backend route and payload; use `POST /api/v1/change/create`, not the idea draft's `/api/v3/change/create`.
- The generated Change file's `Types:` line comes from `.mch/default/prompts/spec-file-structure.md` and supplies the backend `change_types` required by Change create.
- The backend owns `ref`, `slug`, default phase, `agent_edit`, `open`, and project reference allocation.
- The existing `mch` editor behavior should be reused so terminal process handling, editor fallback, and screen restoration stay consistent.
- Codex process execution should be abstracted enough for tests to use fake command runners instead of launching real Codex.
- Detecting whether the review editor was saved should compare file contents or file metadata before and after editor exit; unchanged content advances to final Change generation.

## Relevant Specs

- `.mch/default/prompts/spec-file-structure.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/concepts.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/current-project-context.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/operations/verification.md`

## Verification

- From `cli`: `make lint`
- From `cli`: `go test ./...`
- From `cli`: `go build -o /tmp/mch ./cmd/mch`
- From the repository root: `wc -l docs/architecture/mch.md docs/architecture/cli.md docs/functionality/agent-interaction.md docs/functionality/change-lifecycle.md docs/functionality/current-project-context.md docs/functionality/requirements-and-acceptance.md docs/operations/verification.md`
- From the repository root: `rg -n "api/v3|internal/planning" docs/architecture/mch.md docs/architecture/cli.md docs/functionality/agent-interaction.md cli`

## QA Test Cases

- Start `/new-change` with no valid current project and confirm the flow stops with a recoverable project-context error and no Codex or create request.
- Start `/new-change` when `/tmp/mch` is absent and confirm the directory and blank `initial-idea.md` are created.
- Start `/new-change` when `/tmp/mch` is a regular file and confirm it is replaced by a directory.
- Start `/new-change` when non-empty `initial-idea.md` already exists, choose `/resume`, and confirm the existing contents open in the editor.
- Start `/new-change` when non-empty `initial-idea.md` already exists, choose `/clear`, and confirm a blank idea file opens.
- Start `/new-change` when non-empty `initial-idea.md` already exists, choose `/cancel`, and confirm the flow returns to the list without opening an editor.
- Resume an existing idea, exit without changing it, and confirm the flow returns to the list without Codex or backend calls.
- Resume an existing idea, change it, and confirm the rewrite prompt `Use $change-idea-tmp.` runs.
- Exit the first editor with an empty idea and confirm `ChangesListScreen` returns without Codex or backend calls.
- Enter a non-empty idea and confirm the initial Codex rewrite writes the JSON log, output file, and session ID.
- Simulate Codex rewrite output other than `Done.` and confirm a verbose error is shown with formatted Codex output still visible on `AgentRunningScreen`.
- Save changes during review and confirm the same Codex session is resumed for another rewrite pass.
- Exit review without saving and confirm interactive Codex resumes with `Use $change-spec-tmp.`.
- Provide a valid generated Change file and confirm the backend create payload includes numeric `project_id`, parsed `title`, full `body`, and parsed `change_types`.
- Provide a valid generated Change file with multiple QA Test Case bullets and confirm `mch` creates one backend test case for each scenario before reloading details.
- Simulate a generated QA Test Case create failure and confirm the error remains visible and no created details screen is shown.
- Provide a generated Change file without a title and confirm no backend create request is sent.
- Provide a generated Change file without `Types:` or with malformed types and confirm no backend create request is sent.
- Simulate a backend create failure and confirm the error remains visible and no created details screen is shown.
- Complete a successful create and confirm `mch` reloads and renders the created Change detail using backend data.

## Review Focus

- State transitions around `/new-change`, editor exit, saved-versus-unchanged review detection, and recovery from failures.
- Correct file handling under `/tmp/mch` without touching repository files outside the intended docs and CLI implementation.
- Codex session ID extraction from `thread.started.thread_id` JSON events and exact `Done.` output validation.
- Strict parsing of the generated Change title and `Types:` line into the backend create payload.
- Extraction of generated QA Test Cases into ordered backend test case create requests.
- Use of `/api/v1/change/create` and backend-owned fields without direct database writes.
- Test fakes for editor and Codex execution so automated tests remain deterministic.

## Follow-Ups

- Add optional epic selection or generated epic parsing for planned Changes.
- Consider replacing direct Codex process orchestration with backend-mediated planning endpoints after those endpoints support the full workflow.
