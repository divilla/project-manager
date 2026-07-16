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

`mch` owns the interactive AI-assisted `/new-change` flow from the Change list. The flow uses the committed repository-root `.mch/config.yaml` for `backend_url`, `temp_dir`, and numeric `project_id`, and uses `.mch/default` as the active Flow profile. Saved product data must still be created through supported backend APIs. Change Flow assignment, Run controls, claim reset, and branch reconciliation controls are not supported by the CLI until a dedicated CLI Change adds them; `mch` must not assign `ref`, `slug`, Flow snapshot fields, or Run state locally.

`mch` resolves the current Git repository root before loading configuration, so starting from the root or a nested directory uses the same `.mch` tree. The Main screen includes `/config`, which opens a read-only view of resolved in-memory CLI and Flow configuration without calling backend APIs, reading raw YAML for rendering, executing hooks, or saving local files.

## Reusable Flow Runtime
The CLI provides an uncomposed Flow runtime for later product Flows. A Flow definition is immutable static behavior that can be represented in YAML and names Steps, tasks, generic Screens, artifacts, commands, and typed destinations. A Flow context holds runtime-only state such as the configured temporary directory, active Change identity, originating navigation Screen, current Step and artifact, session data, and execution results. Composition supplies both values, the allowed terminal navigation Screens, external-command boundaries, and artifact persistence; the runtime does not depend on whether a validated definition was constructed in Go or loaded by a later configuration Change.

A Step operates on one of the supported artifact identifiers `idea`, `spec`, `pr`, `implement`, `review`, or `finalize` and uses Editor, Exec, Exec followed by Interactive, or Interactive tasks. Editor, Exec, Interactive, Preview, and Error are generic reusable Screens configured by definition and context rather than artifact-specific Screen implementations. User-facing commands map to typed destinations: a `step` destination starts another Step with a fresh artifact load, while a `screen` destination ends the Flow and navigates to a composition-approved terminal Screen.

Runtime artifacts are isolated under `<temp_dir>/<artifact>/`, where `temp_dir` is supplied from `.mch/config.yaml`. The runtime uses `session-id`, `input.md`, `output.md`, and `agent-output.md` resources in that directory, with `input.md` retained as the Step baseline and tasks editing only `output.md`. The generic persistence boundary loads and saves plain artifact bytes. Its Change API adapter loads through `POST /api/v1/change/get` and saves changed Idea, Spec, or PR bytes through the matching focused update endpoint; Editor-initiated saves send `agent_edit: false`. Other artifact identifiers require later persistence adapters.

This runtime does not compose the Idea Stage or any other product Flow. It does not connect to `/new-change`, Change-detail editing, or the existing `.mch/default/flow.yaml` loader, and it does not change `/config`. Those existing behaviors remain active until later Changes deliberately compose and route a product Flow.

Repository Change workflow automation uses `change/<change-slug>` branches, ideas under `agent/ideas/<change-slug>.md`, and Change specs under `specs/<change-slug>.md`. The spec structure template is `.mch/default/prompts/spec-file-structure.md`. This does not rename application routes, API paths, packages, or product data that use the Change concept.

## Current Project
For `mch`, current project selection is committed repository and branch configuration stored as numeric `project_id` in `.mch/config.yaml`. Missing `project_id` and `project_id: 0` mean no current project is selected. CLI commands that operate on project-scoped data must use a valid configured project ID or require an explicit project option.

## Agent Use
Agents can use CLI commands when commands are deterministic, documented, and return structured output. Prefer JSON output for commands intended for automation.

## Boundaries
The CLI must not become a parallel product implementation. Product rules remain owned by backend services, database constraints, and documented workflows.
