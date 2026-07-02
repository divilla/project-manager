Turn the rough notes in Change file `agent/changes/110-cli-improve-changes-and-agentic-workflow.md` into a complete, implementation-ready Change specification.

Rewrite only that Change file.

Hard constraints:
- Do not implement code, edit tests, edit docs outside the Change file, run migrations, mutate any database, or edit `AGENTS.md`.
- Use Template file `agent/prompts/change-file-structure.md` exactly. Do not add, remove, rename, or reorder sections.

First, read Change file and Template file

Then use targeted search, not exhaustive reading, to find relevant documentation under `docs/`. Prioritize docs covering `mch`, CLI behavior, backend API endpoints, Change lifecycle, test cases/completeness, persistence contracts, and verification.
Read only docs that materially affect this Change.

Documentation is the source of truth. If the draft conflicts with docs, follow the docs and record important assumptions in `Design Notes`.

Every existing bullet in the draft must be accounted for. For each bullet, either:
- incorporate it into the rewritten Change as a requirement, acceptance criterion, QA case, design note, non-goal, or follow-up; or
- ask a clarifying question before rewriting.

Before rewriting, check for ambiguity in scope, API contracts, database/persistence behavior, CLI behavior, keybindings, display labels, Unicode symbols, colors, verification, and QA expectations.

If clarification is needed, stop and output only concise clarifying questions. Do not rewrite yet. Do not guess missing product decisions unless documentation directly supports the answer.

Keep the Change scoped to one coherent outcome. Convert vague notes into concise product requirements. Move related but non-essential ideas to `Non-Goals` or `Follow-Ups`. Do not expand scope.

The final Change must satisfy:
- `Goal`: one observable end state.
- `Scope`: only directly required behavior/files.
- `Requirements`: required behavior, visible contracts, persistence expectations, and boundaries.
- `Acceptance Criteria`: observable success conditions.
- `Verification`: realistic repo-supported commands for every affected area.
- `QA Test Cases`: happy paths, validation failures, backend/command failures, cancellation/no-op paths, persistence behavior, and boundary cases.
- `Review Focus`: riskiest contracts to inspect.
- `Follow-Ups`: future work outside scope, or `- None.`.

Output either:
1. Concise clarifying questions, if required; or
2. The complete rewritten content for `agent/changes/110-cli-improve-changes-and-agentic-workflow.md`, ready to replace the file.
