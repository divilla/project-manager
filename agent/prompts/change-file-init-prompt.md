Extract `<change-slug>` from the current Git branch using this regex:

`^change/([0-9]+-[0-9A-Za-z_-]+)$`

If the current branch does not match, stop and output one concise error explaining that the branch must be named `change/<change-slug>`.

Use `agent/ideas/<change-slug>.md` to write a completed, implementation-ready Change spec to `specs/<change-slug>.md`. If that Change spec already exists, overwrite it intentionally.

Only write that Change spec. Do not edit any other file, including `AGENTS.md`. Do not run migrations, mutate databases, edit files under `db/**`, or perform unrelated cleanup.

Use the template in `.mch/default/prompts/spec-file-structure.md` exactly. Do not add, remove, rename, or reorder sections.

Process:
1. Read `agent/ideas/<change-slug>.md`.
2. Read `.mch/default/prompts/spec-file-structure.md`.
3. Inspect the initial code and tests that define current behavior and constraints.
4. Read a retained ADR or operational runbook only when the Idea or user explicitly cites it.
5. Write the completed Change spec to `specs/<change-slug>.md`.
6. Add and commit only `specs/<change-slug>.md`; do not include unrelated staged or unstaged changes. Use this commit message: `Write spec for <change-slug> by agent`. After the commit succeeds, output exactly `Done.`

Documentation is outside the default Change Flow. Do not inspect it by default or add documentation
work to the Spec unless the Idea or user explicitly requests it. Code is the source of truth for
current behavior; the Idea supplies the requested direction, and the Spec defines the desired
future state.

Account for every bullet in the idea file. Each bullet must either be incorporated into the Change as a requirement, acceptance criterion, QA case, design note, non-goal, or follow-up, or it must trigger a clarifying question before writing.

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

If clarification is required, stop and output only concise clarifying questions. Do not write or
commit the Change yet. Do not guess product decisions unless current code or an established project
convention resolves the answer without changing the requested outcome.

Keep the Change scoped to one coherent outcome. Convert vague notes into concise product requirements. Move related but non-essential ideas to `Non-Goals` or `Follow-Ups`. Do not expand scope.

The final Change must satisfy:
- `Goal`: one observable end state.
- `Scope`: only directly required behavior and files.
- `Requirements`: required behavior, visible contracts, persistence expectations, and boundaries.
- `Acceptance Criteria`: observable success conditions.
- `Verification`: realistic repo-supported commands for every affected area.
- `QA Test Cases`: happy paths, validation failures, backend or command failures, cancellation/no-op paths, persistence behavior, and boundary cases.
- `Review Focus`: riskiest contracts to inspect.
- `Follow-Ups`: future work outside scope, or `- None.`.

Stop without writing or committing if:
- the branch name does not match `^change/([0-9]+-[0-9A-Za-z_-]+)$`
- `agent/ideas/<change-slug>.md` does not exist
- `.mch/default/prompts/spec-file-structure.md` does not exist
- clarification is required

Final response must be exactly one of:
1. Concise clarifying questions, if clarification is required.
2. A concise error explaining why the Change could not be written.
3. `Done.`
