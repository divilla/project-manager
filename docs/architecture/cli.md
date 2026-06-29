# CLI

## Purpose
The CLI is an optional automation surface for developers and agents. It should expose stable commands for project context, change workflow, and local verification without bypassing backend rules.

## Design Direction
The CLI should call supported backend APIs or documented local commands. It should not write application tables directly unless a command is explicitly an operations command and is documented as such.

## Prototype
The `cli-proto/` directory may contain experimental terminal prototypes. The first prototype binary is `mch`, a Bubble Tea app for Codex-assisted Change requirement planning. It starts without subcommands, accepts only `--backend-url`, resolves the current Git repository root with `git rev-parse --show-toplevel`, and uses that root for prompt lookup and Codex execution.

The prototype stores its local config under `cli-proto/.config`. It may persist backend URL and current project selection there, but it must save product data only through supported backend APIs.

## Reference TUI
The `make-a-change/` module contains the reference `mch` TUI baseline. Its architecture, package boundaries, style tokens, state model, and test strategy are documented in `docs/architecture/mch.md`.

## Current Project
Current project selection is user-specific application state. CLI commands that operate on project-scoped data should read the same user setting as the app or require an explicit project option.

## Agent Use
Agents can use CLI commands when commands are deterministic, documented, and return structured output. Prefer JSON output for commands intended for automation.

## Boundaries
The CLI must not become a parallel product implementation. Product rules remain owned by backend services, database constraints, and documented workflows.
