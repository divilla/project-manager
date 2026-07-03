# Change file structure

Use this structure exactly for every Change file. Replace placeholders with content for the active Change. Keep sections concise, specific, and testable. Do not add, remove, rename, or reorder sections unless the user explicitly changes the Change workflow.

The first non-blank line must be the Change title as one H1. The first non-blank line after the title must be the type metadata line.

Select one or more backend type slugs that best describe the Change. Do not hardcode, invent, or assume allowed type slugs. Use the type options supplied by the active workflow context; if no current type options are supplied, retrieve them from `POST /api/v1/options/change-types-list` when the environment supports backend access, otherwise stop and ask for valid backend type slugs.

Format the metadata line exactly as `Types: <type-slugs>`, with selected backend slugs joined by `|` and no spaces.

Do not wrap the generated Change file in a code block.

# <Change Title>

Types: <type-slugs>

## Goal

Describe the single outcome this Change must deliver. Write this as the end state the user should observe, not as a list of implementation tasks.

## Scope

- List the behavior, documentation, architecture, or implementation areas included in this Change.
- Keep every bullet directly tied to the active Change.
- Exclude adjacent work that is useful but not required for this Change.

## Requirements

- State testable requirements using product vocabulary from `docs`.
- Include expected behavior, important boundaries, and failure handling where relevant.
- Write requirements as obligations the implementation must satisfy.

## Acceptance Criteria

- Define observable success conditions for this Change.
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

- `agent/changes/<change-name>.md`
- `docs/<path>.md`

## Verification

- List every command needed to verify this Change.
- Include backend, frontend, lint, typecheck, race, API-test, or build commands when the Change touches those areas.
- Do not invent commands the repository cannot run.

## QA Test Cases

- List the manual or product-level scenarios QA should test.
- Cover happy paths, validation failures, command or backend failures, cancellation or no-op paths, persistence behavior, and important boundary cases when relevant.
- Keep QA scenarios distinct from automated Verification commands.

## Review Focus

- Call out risky or subtle areas reviewers should inspect first.
- Highlight changed contracts, data flow, persistence, migrations, concurrency, security, generated artifacts, or workflow automation when relevant.

## Follow-Ups

- List useful future work that is outside this Change.
- Use `- None.` when there are no known follow-ups.
