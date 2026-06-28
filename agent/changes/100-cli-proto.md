# CLI Prototype

## Goal

Deliver a minimal Bubble Tea `mch` prototype that lets a user start Codex-assisted Change requirement planning from a terminal, refine the requirement in one in-memory Codex session, and save the final result as a backend Change only after validation, editing, and explicit confirmation.

## Scope

- Add prototype code under repository-root `cli-proto/` with a binary named `mch`.
- Start the app with `mch` and support only the optional startup flag `--backend-url`.
- Resolve the Git repository root after startup and use it for prompt lookup and all Codex invocations.
- Store prototype config under `cli-proto/.config`, including backend URL and current project selection.
- Build an interactive Bubble Tea UI that supports project selection, `/change-new`, refinement prompts, save, and `/cancel`.
- Use `agent/prompts/build-requirement-with-agent.md` as the controlled prompt template for requirement generation.
- Invoke Codex for new and resumed sessions and keep `codex_session_id` only in memory until save or cancel.
- Parse, validate, edit, confirm, and persist the final markdown output as a backend Change.
- Extend backend Change create behavior to accept and persist nullable `codex_session_id`.
- Save CLI-created Changes in the `backlog` phase.

## Requirements

- `mch` must not require subcommands and must not use Cobra.
- The only startup flag must be `--backend-url`; when present, it overrides the configured backend URL for the running app.
- Startup must run `git rev-parse --show-toplevel` and fail before any Codex invocation when the repository root cannot be resolved to a non-empty readable path.
- The resolved repository root must be used to locate `agent/prompts/build-requirement-with-agent.md`.
- Every Codex invocation must include `-C <resolved-repo-root>`, including resumed sessions.
- Missing config must be initialized under `cli-proto/.config`; missing backend URL must default to `http://localhost:8080` and be saved.
- Project selection must happen inside the interactive UI and must be persisted as the current project in config.
- The app must show an input prompt after startup and recognize `/change-new` and `/cancel`.
- `/change-new` must ask for an initial change idea, inject it into the controlled prompt template, and start Codex with `codex exec -C <resolved-repo-root> -`.
- The app must extract `codex_session_id` from Codex output and display Codex output in the terminal UI.
- Refinement prompts during an active planning flow must resume the same Codex session with `codex exec -C <resolved-repo-root> resume <codex-session-id> -`.
- Refinement must not create or update backend Changes.
- `/cancel` must exit the app; if no save occurred, no Change may be created.
- Save must require final markdown output with exactly one H1 title and a valid `Types: <type-slugs>` metadata line.
- Save must support an optional `Epic: <epic-name>` metadata line.
- Type slugs must be validated against `POST /api/v1/change/reference`.
- Epic names must be validated against `POST /api/v1/epic/list` for the selected project.
- Save must remove the H1, `Types:`, and optional `Epic:` lines before persisting the Change body.
- Save must open the body in `$EDITOR` before final confirmation.
- Save must create the backend Change only after explicit confirmation.
- Backend Change create must accept and persist nullable `codex_session_id`.
- Saved planned Changes must use `backlog` as `change_phase`.

## Acceptance Criteria

- Running `mch` starts the interactive terminal UI without requiring a subcommand.
- Running `mch --backend-url http://localhost:8080` uses that backend URL for the running app.
- Starting outside a Git repository, with an empty repository root, or with an unreadable repository root fails before Codex is invoked.
- The prompt template is read from `<resolved-repo-root>/agent/prompts/build-requirement-with-agent.md`.
- Initial and resumed Codex commands both include the same resolved repository root in `-C`.
- Missing config creates `cli-proto/.config` with default backend URL when no override exists.
- A missing or stale current project causes the UI to ask the user to select from backend projects.
- Unknown slash commands and `/change-new` during an active planning flow produce recoverable UI errors.
- Empty initial ideas, missing Codex command, unauthenticated Codex, Codex failures, timeouts, and missing session ID are surfaced without saving a Change.
- Save rejects markdown with no H1, multiple H1 headings, missing or malformed `Types:`, unknown type slugs, or an unknown epic.
- `$EDITOR` failure aborts save before backend Change creation.
- Confirmed save calls backend Change create with `change_phase: "backlog"` and the in-memory `codex_session_id` when present.
- Cancel before save exits without creating or updating a backend Change.
- Backend create, get, and list behavior expose persisted `codex_session_id` for saved Changes.

## Non-Goals

- No command tree after `mch`.
- No subcommands such as `mch project list`.
- No startup flags except `--backend-url`.
- No production CLI packaging.
- No frontend planning page.
- No plan API group.
- No automatic backend save during refinement.
- No PR body generation, PR publishing, or GitHub integration.
- No `agent/changes/*.md` file export or sync.
- No repository-wide context indexing.
- No persistent planning-session table.
- No child requirement row generation.

## Design Notes

- The detailed requirement source is `agent/reqs/100-cli-proto.md`.
- The controlled prompt assets live in `agent/prompts/`.
- The CLI prototype is intentionally scoped to `cli-proto/` so it can evolve without disrupting the backend, frontend, or future production CLI packaging.
- Backend behavior remains authoritative for Change validation and persistence; the CLI validates references before save to give immediate feedback, then still relies on backend validation.
- `codex_session_id` is process memory in the CLI until save, then persisted on the created Change for traceability.
- Bubble Tea command handling should keep the UI responsive while Codex runs.

## Relevant Specs

- `docs/architecture/cli.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/current-project-context.md`
- `docs/functionality/agent-interaction.md`

## Verification

- From `cli-proto`: `go test ./...`
- From `cli-proto`: `go build -o /tmp/mch ./cmd/mch`
- From `backend`: `make test`
- From `backend`: `make api-test`

## Review Focus

- Verify Codex is always invoked with the resolved repository root and that the same root is used for resume.
- Verify save cannot create a backend Change before markdown, type, epic, editor, and confirmation checks pass.
- Verify backend `codex_session_id` persistence is nullable and does not break existing Change create callers.
- Verify Bubble Tea async commands cannot start overlapping planning sessions or save stale output.

## Follow-Ups

- Production CLI packaging and install instructions.
- Persistent planning-session resume across app restarts.
- Richer automation-friendly CLI commands once the prototype interaction model is validated.
- Fixed PR comment `IC_kwDOTA2Xls8AAAABH72nRg`: removed generated `cli-proto/.config/config.json` from the Change and ignored `cli-proto/.config/` so config remains local runtime state.
