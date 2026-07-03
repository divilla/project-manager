Extract `<change-name>` from the current Git branch using this regex: 

`^changes/([0-9]+-[a-z0-9-]+)$`

If the current branch does not match, stop and output one concise error explaining that the branch must be named `changes/<change-name>`.

Use `agent/ideas/<change-name>.md` to write a completed, implementation-ready Change specification to `agent/changes/<change-name>.md`. If that Change file already exists, overwrite it intentionally.

Only write that Change file. Do not edit any other file, including `AGENTS.md`. Do not run migrations, mutate databases, edit files under `db/**`, or perform unrelated cleanup.

Use the template in `agent/prompts/change-file-structure.md` exactly. Do not add, remove, rename, or reorder sections.

Process:
1. Read `agent/ideas/<change-name>.md`.
2. Read `agent/prompts/change-file-structure.md`.
3. Use targeted search to find relevant docs under `docs/`; do not broadly read unrelated docs.
4. Read only docs that materially affect this Change.
5. Write the completed Change specification to `agent/changes/<change-name>.md`.
6. Add and commit only `agent/changes/<change-name>.md`; do not include unrelated staged or unstaged changes. Use this commit message: `Init change <change-name> by agent`. After the commit succeeds, output exactly `Done.`

Prioritize docs covering:
- `mch`
- CLI behavior
- backend API endpoints
- Change lifecycle
- test case completeness
- persistence contracts
- verification

Treat documentation as the source of truth. If the idea file conflicts with docs, follow the docs and record important assumptions in `Design Notes`.

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

If clarification is required, stop and output only concise clarifying questions. Do not write or commit the Change yet. Do not guess product decisions unless documentation directly supports the answer.

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
- the branch name does not match `^changes/([0-9]+-[a-z0-9-]+)$`
- `agent/ideas/<change-name>.md` does not exist
- `agent/prompts/change-file-structure.md` does not exist
- clarification is required

Final response must be exactly one of:
1. Concise clarifying questions, if clarification is required.
2. A concise error explaining why the Change could not be written.
3. `Done.`
