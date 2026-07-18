# CLI

## Purpose
`cli/` is the repository directory for terminal-facing tools. Today it contains the `mch` Bubble Tea terminal UI, so references to the "CLI module" in this repository usually mean the interactive TUI application rather than a broad subcommand-oriented automation CLI.

Future non-interactive CLI commands may be added for developers and agents. They should expose stable commands for project context, change workflow, and local verification without bypassing backend rules.

## Design Direction
The CLI should call supported backend APIs or documented local commands. It should not write application tables directly unless a command is explicitly an operations command and is documented as such.

## Prototype
The `cli-proto/` directory may contain experimental terminal prototypes. The first prototype binary is `mch`, a Bubble Tea app for Codex-assisted Change test case planning. It starts without subcommands, accepts only `--backend-url`, resolves the current Git repository root with `git rev-parse --show-toplevel`, and uses that root for prompt lookup and Codex execution.

The prototype stores its local config under `cli-proto/.config`. It may persist backend URL and current project selection there, but it must save product data only through supported backend APIs.

## Reference TUI
The `cli/` module contains the reference `mch` TUI. Its architecture, package boundaries, style tokens, state model, and test strategy are documented in `docs/architecture/mch.md`. The executable remains `mch`; `cli/` is only the source directory name.

`mch` owns the interactive AI-assisted `/new-change` flow from the Change list. The committed repository-root `.mch/config.yaml` provides `backend_url` and numeric `project_id`; the Flow temp root is the hardcoded `.mch/tmp`, not configuration. `.mch/default` remains the active Flow profile for repository Flow metadata, while the current Idea Stage uses a built-in Go definition. Saved product data must still be created through supported backend APIs. Change Flow assignment, Run controls, claim reset, and branch reconciliation controls are not supported by the CLI until a dedicated CLI Change adds them; `mch` must not assign `ref`, `slug`, Flow snapshot fields, or Run state locally.

Multiple `mch` instances may run simultaneous `/new-change` attempts. Each invocation generates a UUID v4 and atomically creates `<git-root>/<temp-dir>/<attempt-uuid>` without checking existence first; a collision retries with another UUID. Its isolated `new-idea.md` prevents an older validation from rewriting a later draft. The UUID is also the runtime correlation token for validation and backend-create replies, but it is not sent as the backend-owned `ref_uuid`. This isolation needs no locking or process coordination.

`mch` resolves the current Git repository root before loading configuration, so starting from the root or a nested directory uses the same `.mch` tree. Startup requires both `<git-root>/.mch` and `<git-root>/.mch/tmp` to exist as directories and does not create either root. The Main screen includes `/config`, which opens a read-only view of resolved in-memory CLI and Flow configuration without displaying a temp-root setting, calling backend APIs, reading raw YAML for rendering, executing hooks, or saving local files.

## Reusable Flow Runtime
The CLI provides a reusable Flow runtime and composes the current Idea Stage with a fresh built-in Go definition. A Flow definition is immutable static behavior that can be represented in YAML and names Steps, Tasks, generic Screens, artifacts, commands, and typed destinations. A Flow context holds runtime-only state such as active Change identity, originating navigation Screen, active Step and Task position, Preview `from-step`, artifact, session data, and execution results. Composition supplies both values, allowed terminal navigation Screens, external-command boundaries, and artifact persistence; the generic runtime does not branch on Idea-specific Step names or commands. Loading this product definition from `.mch/default/flow.yaml` is not yet active.

A Step is a non-executable container for an ordered sequence of `1..n` Tasks operating on one artifact. Its Mode determines the accepted Task types and order: Editor Mode is exactly `EditorTask`, Chat Mode is exactly `ChatTask`, and Exec Mode is exactly `ExecTask -> ChatTask`; Script Mode is not supported. Selecting a typed Step destination activates its first Task, and runtime traversal advances by Task position. An expected Exec final line skips the following Chat position; an empty or different line advances to Chat without leaving the Step. Both paths reach Preview only because the current Exec shape has no later Task. The ordered cursor is the extension boundary for future Mode contracts, although additional shapes remain invalid now.

`IdeaRewriteExec` and `IdeaReviewExec` are composed Step names: `idea-rewrite` or `idea-review` plus Exec Mode. Neither name denotes their first `ExecTask`. The corresponding `IdeaRewriteChat` and `IdeaReviewChat` Screens run the second Task in the same Step and preserve its session and workspace. Chat displays available agent output without resuming automatically; `/chat` starts the resumable terminal session, whose interactive transcript remains terminal-only.

Preview records the `from-step` whose completed Task entered it. Every Preview provides `/continue`, `/edit`, and `/cancel` in that order. An Exec- or Chat-mode `from-step` inserts `/chat` after `/continue`; that command activates the same Step's `ChatTask` without selecting or restarting a Step, even when static configuration omits it. Editor-mode Preview has no `/chat`. `/edit` selects the configured same-artifact Editor Step and records the calling Preview; `/cancel` returns to `ChangeDetailsState`.

The Go `TmpDir` constant is the only executable temp-root source and has value `.mch/tmp`. Pre-persistence creation uses `<git-root>/<temp-dir>/<attempt-uuid>/new-idea.md`; persisted artifacts use `<git-root>/<temp-dir>/<ref-uuid>/<artifact>`, and Editor-mode Steps isolate their draft pair under `<artifact>/editor`. The `editor` component is a workspace scope, not an artifact identity. Go passes `<temp-dir>` as `MCH_TEMP_DIR`, and execution and resume scripts combine it with required `MCH_REF_UUID` and `MCH_ARTIFACT` to use the artifact root. That root holds session and agent files; Editor holds only its scoped `input.md` and `output.md`. Different Changes are isolated by `ref_uuid`; one Change may have at most one active Flow process.

Starting Exec or Chat fresh-loads persisted artifact bytes into the artifact root's byte-identical `input.md` and `output.md`; starting Editor loads the same baseline into `<artifact>/editor/input.md` and `output.md` without replacing a calling Preview's files. Changed Editor output is saved and its comparison pair is published to the artifact root before Preview, while canonical byte-identical return leaves the caller workspace intact. Preview performs no API operation. Every non-empty Idea, Spec, or PR submission normalizes CRLF to LF, validates a one-line H1, exact blank-line separators, and a body containing at least one non-whitespace character, then canonicalizes optional artifact-local metadata before comparison or persistence. Title-only, whitespace-body, and metadata-only documents are invalid. `Types: <type-slugs>` uses unique API-valid slugs in backend option order with `|` separators. `Epic: <epic-title> #<epic-id>` uses the current-project API title and numeric ID. When both exist, `Types:` precedes `Epic:`. Absent metadata remains absent.

The persistence boundary saves only canonical bytes that differ from the fresh input through the matching focused artifact endpoint. Editor saves use `agent_edit: false`; Exec and Chat saves use `agent_edit: true`. Artifact H1, Types, and Epic values never mutate the Change title, type set, linked Epic, ref, or slug. Empty Editor output cancels without validation recovery; invalid non-empty Editor output offers `/fix` and `/cancel`; canonical byte-identical output performs no save and returns to the Screen that opened Editor.

## Composed Idea Stage

Each `/new-change` atomically allocates `<git-root>/<temp-dir>/<attempt-uuid>/new-idea.md` and opens the empty file without resume or clear commands. Empty output returns to `ChangesListState` without an API request. Non-empty output is validated and canonicalized before `Create Change?`; No returns without creation, while Yes persists the canonical Idea and initializes the Change title from its H1 before starting `IdeaRewriteExec` with list origin. Finishing or abandoning the attempt removes `new-idea.md` first, then removes the UUID directory only when empty; a non-empty result preserves any other files that have appeared under the same UUID, while other cleanup failures enter Error. `/fix` retains the attempt workspace.

Return on the Idea row in `ChangeDetailsState` starts the `IdeaEdit` generic `DocEdit` Step. Changed valid canonical output saves as a user edit and enters Editor-mode Preview; `/continue` starts `IdeaRewriteExec`. Empty output or `/cancel` returns to details. Canonical byte-identical output returns to the exact recorded caller: Change details, the calling Preview, or the calling Chat Screen. Preview or Chat `/edit` abandons any unpersisted proposal and starts `IdeaEdit` from a fresh persisted baseline.

Rewrite expects final agent output `Done.` and Review expects `No questions or suggestions.`. A matching line skips Chat and enters the corresponding Preview after any agent save; empty or different readable output advances to the same Step's Chat Task. Rewrite Preview `/continue` starts `IdeaReviewExec`. Review Preview `/continue` temporarily exits to `MainState`; it does not start Spec work. Exec `/stop`, Chat `/cancel`, and Error `/return` use the recorded origin, while Preview and validation cancellation return to `ChangeDetailsState`.

Repository Change workflow automation uses `change/<change-slug>` branches, ideas under `agent/ideas/<change-slug>.md`, and Change specs under `specs/<change-slug>.md`. The spec structure template is `.mch/default/prompts/spec-file-structure.md`. This does not rename application routes, API paths, packages, or product data that use the Change concept.

## Current Project
For `mch`, current project selection is committed repository and branch configuration stored as numeric `project_id` in `.mch/config.yaml`. Missing `project_id` and `project_id: 0` mean no current project is selected. CLI commands that operate on project-scoped data must use a valid configured project ID or require an explicit project option.

## Agent Use
Agents can use CLI commands when commands are deterministic, documented, and return structured output. Prefer JSON output for commands intended for automation.

## Boundaries
The CLI must not become a parallel product implementation. Product rules remain owned by backend services, database constraints, and documented workflows.
