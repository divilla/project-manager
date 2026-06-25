# CLI

## Purpose
The CLI is an optional automation surface for developers and agents. It should expose stable commands for project context, change workflow, and local verification without bypassing backend rules.

## Design Direction
The CLI should call supported backend APIs or documented local commands. It should not write application tables directly unless a command is explicitly an operations command and is documented as such.

## Current Project
Current project selection is user-specific application state. CLI commands that operate on project-scoped data should read the same user setting as the app or require an explicit project option.

## Agent Use
Agents can use CLI commands when commands are deterministic, documented, and return structured output. Prefer JSON output for commands intended for automation.

## Boundaries
The CLI must not become a parallel product implementation. Product rules remain owned by backend services, database constraints, and documented workflows.
