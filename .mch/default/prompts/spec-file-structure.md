# Spec File Structure

Use this structure exactly for every Spec file. Replace placeholders with concrete spec content. Keep sections concise, specific, and testable.

A Spec belongs to a Change, but when referring to the file or artifact, call it `Spec` or `spec`, not `Change`. Use `Change` only when referring to the parent workflow item, branch, lifecycle, or implementation effort.

Do not add, remove, rename, or reorder sections unless the user explicitly changes the Spec workflow.

The first non-blank line must be the Spec title as one H1. The first non-blank line after the title must be the type metadata line.

Select one or more backend type slugs that best describe the parent Change. Do not hardcode, invent, or assume allowed type slugs. Use the type options supplied by the active workflow context. If no current type options are supplied, retrieve them from `POST /api/v1/options/change-types-list` when the environment supports backend access. Otherwise, stop and ask for valid backend type slugs.

Format the metadata line exactly as `Types: <type-slugs>`, with selected backend slugs joined by `|` and no spaces.

Do not wrap the generated Spec in a code block.

# <Spec Title>

Types: <type-slugs>

## Goal

Describe the single outcome this Spec must define. Write this as the end state the user should observe, not as a list of implementation tasks.

## Scope

- List the behavior, documentation, architecture, or implementation areas included in this Spec.
- Keep every bullet directly tied to the Spec.
- Exclude adjacent work that is useful but not required for this Spec.

## Requirements

### Docs

### DB

### Backend

### Frontend

### CLI

### Other

- Divide requirements into Docs, DB, Backend, Frontend, CLI, and Other sections.
- Include only sections that have items.
- State testable requirements using product vocabulary from `docs`.
- Include expected behavior, important boundaries, and failure handling where relevant.
- Write requirements as obligations the implementation must satisfy.

## Acceptance Criteria

### Docs

### DB

### Backend

### Frontend

### CLI

### Other

- Divide acceptance criteria into Docs, DB, Backend, Frontend, CLI, and Other sections.
- Include only sections that have items.
- Define observable success conditions for this Spec.
- Include routes, commands, API behavior, UI states, persistence behavior, generated files, or workflow outcomes when relevant.
- Make each criterion verifiable by inspection, automated tests, or a concrete manual check.

## Non-Goals

- List related work that is intentionally out of scope.
- Include decisions that prevent accidental scope expansion.
- Move useful but non-essential ideas here or to Follow-Ups instead of expanding Scope.

## Design Notes

- Record important implementation constraints, data model assumptions, UX details, or workflow rules.
- Link to authoritative docs instead of repeating long explanations.
- Note assumptions that reviewers or future agents must preserve.

## Relevant Specs

- `specs/<change-slug>.md`
- `docs/<path>.md`

## Verification

- List every command needed to verify this Spec.
- Include backend, frontend, lint, typecheck, race, API-test, or build commands when the Spec touches those areas.
- Do not invent commands the repository cannot run.

## QA Test Cases

### Backend

### Frontend

### CLI

### Other

- Divide QA test cases into Backend, Frontend, CLI, and Other sections.
- Include only sections that have items.
- List the manual or product-level scenarios QA should test.
- Cover happy paths, validation failures, command or backend failures, cancellation or no-op paths, persistence behavior, and important boundary cases when relevant.
- Keep QA scenarios distinct from automated Verification commands.

## Review Focus

- Call out risky or subtle areas reviewers should inspect first.
- Highlight changed contracts, data flow, persistence, migrations, concurrency, security, generated artifacts, or workflow automation when relevant.

## Follow-Ups

- List useful future work that is outside this Spec.
- Use `- None.` when there are no known follow-ups.
