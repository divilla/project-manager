# Bubble Tea `mch` Prototype For New Change Requirement Planning

Types: feature

## Problem Statement

Users need a minimal terminal app prototype that starts with a single `mch` command and supports interactive Codex-assisted requirement building for a new Change. The prototype should focus on the in-app interaction model, not command parsing or a broad CLI command tree.

The app starts, resolves the current Git repository root, lets the user select project context inside the terminal UI, and supports a small slash-command interface. The first supported planning command is `/change-new`, which starts a new interactive requirement-building session. `/cancel` exits the app.

## Primary Workflows

A user starts the app:

```bash
mch
```

The only supported startup flag is an optional backend URL override:

```bash
mch --backend-url http://localhost:8080
```

After startup, the app runs:

```bash
git rev-parse --show-toplevel
```

The returned path is used as the repository root for prompt lookup and all Codex invocations.

The app loads or initializes config from `cli-proto/.config`. If no backend URL is configured, it uses `http://localhost:8080` and saves it. If no current project is configured, the app prompts the user to select one from backend projects.

The app displays an input prompt that accepts slash commands.

The user enters:

```text
/change-new
```

The app starts a new requirement-building flow. It asks for the initial change idea, injects that idea into `agent/prompts/build-requirement-with-agent.md`, starts Codex with:

```bash
codex exec -C <resolved-repo-root> -
```

The app extracts the Codex session ID from the Codex response and displays Codex output.

The user can continue entering refinement prompts during the active requirement-building flow. The app resumes the same in-memory Codex session with:

```bash
codex exec -C <resolved-repo-root> resume <codex-session-id> -
```

When the user is satisfied, the app allows saving the final requirement as a backend Change after parsing and validating the final markdown output.

At any point, the user can enter:

```text
/cancel
```

The app exits. If no save occurred, no Change is created.

## Acceptance Criteria

- Prototype code is scoped to repository-root `cli-proto/`.
- Binary is named `mch`.
- App starts with `mch`.
- App does not require subcommands.
- App supports only one startup flag: `--backend-url`.
- Interactive terminal UI is built with Bubble Tea.
- Prototype does not use Cobra.
- App runs `git rev-parse --show-toplevel` after startup.
- App requires `git rev-parse --show-toplevel` to return a non-empty repository root path.
- App uses the resolved repository root to locate `agent/prompts/build-requirement-with-agent.md`.
- App always invokes Codex with `-C <resolved-repo-root>`.
- App uses the same resolved repository root for initial and resumed Codex sessions.
- App fails before invoking Codex if the repository root cannot be resolved.
- Config is stored under `cli-proto/.config`.
- Missing backend URL defaults to `http://localhost:8080` and saves that value.
- `--backend-url` overrides the configured backend URL for the running app.
- App selects project inside the interactive UI, not through startup commands.
- App stores selected current project in config.
- App shows an input prompt after startup.
- App recognizes `/change-new`.
- `/change-new` starts a new Codex-backed requirement planning flow.
- App asks for the initial change idea after `/change-new`.
- App uses `agent/prompts/build-requirement-with-agent.md` as the controlled prompt template.
- App starts a new Codex session for the initial planning prompt.
- App extracts `codex_session_id` from the session ID line in Codex output.
- App stores `codex_session_id` only in process memory until save or cancel.
- App accepts refinement prompts during an active planning flow.
- App resumes the same Codex session for refinement prompts.
- App displays Codex output in the terminal UI.
- App does not create or update backend Changes during refinement.
- `/cancel` exits the app.
- If `/cancel` occurs before save, no Change is created.
- Save requires final markdown output with a single H1 title.
- Save requires a valid `Types: <type-slugs>` metadata line.
- Save supports optional `Epic: <epic-name>` metadata line.
- Save validates type slugs against `POST /api/v1/change/reference`.
- Save validates epic name against `POST /api/v1/epic/list` for the selected project.
- Save removes H1, `Types:`, and optional `Epic:` lines before persisting requirement body.
- Save opens requirement body in `$EDITOR` before final confirmation.
- Save creates a Change only after explicit confirmation.
- Backend Change create accepts and persists nullable `codex_session_id`.
- Saved planned Changes always use `backlog` as `change_phase`.

## Edge Cases

- `mch` is started outside a Git repository.
- `git rev-parse --show-toplevel` fails.
- Resolved repository root is empty.
- Resolved repository root is not readable.
- Prompt template is missing under the resolved repository root.
- Codex invocation is attempted without `-C <resolved-repo-root>`.
- Backend URL config is missing or invalid.
- Backend API is unavailable.
- No projects exist.
- Configured project no longer exists.
- User enters unknown slash command.
- User enters `/change-new` while a planning flow is already active.
- Initial idea is empty.
- Codex command is unavailable.
- Codex command is not authenticated.
- Codex command fails or times out.
- Codex session ID line is missing.
- Codex output is conversational and not saveable.
- Markdown has no H1.
- Markdown has multiple H1 headings.
- `Types:` line is missing or malformed.
- `Types:` contains unknown slugs.
- `Epic:` names an unknown epic.
- `$EDITOR` is unset or exits unsuccessfully.
- User cancels before saving.
- Backend rejects Change create.

## Non-Goals

- No command tree after `mch`.
- No subcommands like `mch project list`.
- No flags except `--backend-url`.
- No Cobra.
- No production CLI packaging.
- No frontend planning page.
- No plan API group.
- No automatic DB save during planning.
- No PR body generation.
- No PR publishing.
- No GitHub integration.
- No `agent/changes/*.md` file export or sync.
- No repo-wide context indexing.
- No persistent planning-session table.
- No child requirement row generation.

## Dependencies And Risks

- Prototype depends on Bubble Tea.
- Prototype depends on authenticated local Codex CLI availability.
- Prototype depends on reachable backend API.
- Prototype depends on `git rev-parse --show-toplevel` resolving the repository root.
- Prototype depends on access to `agent/prompts/build-requirement-with-agent.md`.
- Prototype depends on Codex output including a stable session ID line.
- Bubble Tea async command handling must keep the UI responsive while Codex runs.
- In-memory session storage means interrupted planning sessions cannot be resumed.
- Codex may produce conversational output, so the UI must distinguish displayable output from saveable final output.

## Open Questions

None.
