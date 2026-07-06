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

Repository Change workflow automation uses `change/<change-name>` branches and Change specs under `specs/<change-name>.md`. This does not rename application routes, API paths, packages, or product data that use the Change concept.

## Current Project
For `mch`, current project selection is committed repository and branch configuration stored as numeric `project_id` in `.mch/config.yaml`. Missing `project_id` and `project_id: 0` mean no current project is selected. CLI commands that operate on project-scoped data must use a valid configured project ID or require an explicit project option.

## Agent Use
Agents can use CLI commands when commands are deterministic, documented, and return structured output. Prefer JSON output for commands intended for automation.

## Boundaries
The CLI must not become a parallel product implementation. Product rules remain owned by backend services, database constraints, and documented workflows.
