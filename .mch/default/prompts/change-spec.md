---
name: change-spec
description: Create or rewrite an implementation-ready Change specification. Use when asked to initialize, write, generate, update, or rewrite a Change spec or Change spec from an idea file in the repository Change workflow; formerly change-init.
---

Extract `<change-slug>` from the current Git branch using this regex: `^change/([0-9]+-[a-z0-9-_]+)$`.

If the current branch does not match, stop and output one concise error explaining that the branch must be named `change/<change-slug>`.

Use `agent/ideas/<change-slug>.md` to write a completed, implementation-ready Change specification to `specs/<change-slug>.md` like a very senior dev. If that Change spec already exists, overwrite it intentionally.

Only write `specs/<change-slug>.md`. Do not edit any other file, including `AGENTS.md`. Do not run migrations, mutate databases, edit files under `db/**`, or perform unrelated cleanup.

Use the template in `.mch/default/prompts/spec-file-structure.md` exactly. Do not add, remove, rename, or reorder sections.

## Process

1. Read `agent/ideas/<change-slug>.md`.
2. Read `.mch/default/prompts/spec-file-structure.md`.
3. Use targeted search to find relevant docs under `docs/`; do not broadly read unrelated docs.
4. Read only docs that materially affect this Change.
5. Check whether the idea, docs, and template provide enough information to write a specific, testable Change.
6. Write the completed Change specification to `specs/<change-slug>.md`.
7. Add and commit only `specs/<change-slug>.md`; do not include unrelated staged or unstaged changes.
8. Use this commit message: `Write spec for <change-slug> by agent`.
9. After the commit succeeds, output exactly `Done.`

## Writing Rules

Treat documentation as the source of truth. If the idea file conflicts with docs, follow the docs and record important assumptions in `Design Notes`.

Account for every meaningful item in the idea file, including headings, prose paragraphs, bullets, numbered steps, and fenced code blocks. Each item must either be incorporated into the Change as a requirement, acceptance criterion, QA case, design
note, non-goal, or follow-up, or it must trigger a clarifying question before writing.

If the idea file contains fenced code blocks, copy them into the Change spec unchanged. Preserve their surrounding context as closely as possible. If a fenced code block appears under a bullet, numbered step, or paragraph, place it under the generated
Change item that represents that original context.

Keep the Change scoped to one coherent outcome. Convert vague notes into concise product requirements. Move related but non-essential ideas to `Non-Goals` or `Follow-Ups`. Do not expand scope.

Before writing the Change, check for ambiguity in:

- scope
- API contracts
- database or persistence behavior
- CLI behavior
- keybindings
- display labels
- Unicode symbols
- colors
- verification
- QA expectations
- valid Change type slugs

If clarification is required, stop and output only concise clarifying questions. Do not write or commit the Change yet. Do not guess product decisions unless documentation directly supports the answer.

## Final Change Requirements

The final Change must satisfy:

- `Goal`: one observable end state.
- `Scope`: only directly required behavior and files.
- `Requirements`: required behavior, visible contracts, persistence expectations, and boundaries.
- `Acceptance Criteria`: observable success conditions.
- `Non-Goals`: related work intentionally outside scope, or `- None.`.
- `Design Notes`: important implementation constraints, assumptions, and workflow rules.
- `Relevant Specs`: the generated Change spec path and materially relevant docs.
- `Verification`: realistic repo-supported commands for every affected area.
- `QA Test Cases`: happy paths, validation failures, backend or command failures, cancellation/no-op paths, persistence behavior, and boundary cases.
- `Review Focus`: riskiest contracts to inspect.
- `Follow-Ups`: future work outside scope, or `- None.`.

## Stop Conditions

Stop without writing or committing if:

- the branch name does not match `^change/([0-9]+-[a-z0-9-_]+)$`
- `agent/ideas/<change-slug>.md` does not exist or cannot be read
- `.mch/default/prompts/spec-file-structure.md` does not exist or cannot be read
- required Change type slugs cannot be determined from the active workflow context or repository-supported sources
- clarification is required

## Final Response

Final response must be exactly one of:

1. Concise clarifying questions, if clarification is required.
2. A concise error explaining why the Change could not be written or committed.
3. `Done.`, after the Change spec is written and committed successfully.
