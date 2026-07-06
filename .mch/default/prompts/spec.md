Do not use any skills for this request.

Write an implementation-ready Change specification from the configured temp directory into clear Markdown.

## Preconditions

Read `.mch/config.yaml` and extract the top-level `temp_dir` value into `<temp-dir>`.

Extract `<change-name>` from the current Git branch using this regex: `^change/([0-9]+-[a-z0-9-_]+)$`.

Stop with one concise error if:

- `.mch/config.yaml` does not exist or cannot be read
- `temp_dir` is missing, empty, or not a single path value
- the current branch does not match `^change/([0-9]+-[a-z0-9-_]+)$`
- `<temp-dir>/input.md` does not exist or cannot be read
- `<temp-dir>/input.md` is empty or contains only whitespace
- `agent/prompts/change-file-structure.md` does not exist or cannot be read
- required Change type slugs cannot be determined from the active workflow context or repository-supported sources

## Instructions

Read `<temp-dir>/input.md`.

Read `agent/prompts/change-file-structure.md`.

Use targeted search to find relevant docs under `docs/`; do not broadly read unrelated docs. Read only docs that materially affect this Change.

Write a completed, implementation-ready Change specification to `<temp-dir>/output.md`. If `<temp-dir>/output.md` already exists, overwrite it intentionally.

- Use the template in `agent/prompts/change-file-structure.md` exactly. Do not add, remove, rename, or reorder sections.
- Treat documentation as the source of truth. If the input conflicts with docs, follow the docs and record important assumptions in `Design Notes`.
- Account for every meaningful item in the input, including headings, prose paragraphs, bullets, numbered steps, and fenced code blocks.
- Each meaningful input item must either be incorporated into the Change as a requirement, acceptance criterion, QA case, design note, non-goal, or follow-up, or it must trigger a clarifying question before writing.
- Preserve fenced code blocks exactly, including language tags and contents. Preserve their surrounding context as closely as possible.
- Keep the Change scoped to one coherent outcome. Convert vague notes into concise product requirements. Move related but non-essential ideas to `Non-Goals` or `Follow-Ups`.
- Do not invent product decisions, implementation details, API contracts, database behavior, verification steps, QA expectations, or new scope.
- Do not guess product decisions unless documentation directly supports the answer.
- Use `specs/<change-name>.md` as the generated Change spec path when the template needs that path.

Before writing, check for ambiguity in:

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

If clarification is required, stop and output only concise clarifying questions. Do not write `<temp-dir>/output.md`.

The final Change must satisfy:

- `Goal`: one observable end state.
- `Scope`: only directly required behavior and files.
- `Requirements`: required behavior, visible contracts, persistence expectations, and boundaries.
- `Acceptance Criteria`: observable success conditions.
- `Non-Goals`: related work intentionally outside scope, or `- None.`.
- `Design Notes`: important implementation constraints, assumptions, and workflow rules.
- `Relevant Specs`: `specs/<change-name>.md` and materially relevant docs.
- `Verification`: realistic repo-supported commands for every affected area.
- `QA Test Cases`: happy paths, validation failures, backend or command failures, cancellation/no-op paths, persistence behavior, and boundary cases.
- `Review Focus`: riskiest contracts to inspect.
- `Follow-Ups`: future work outside scope, or `- None.`.

Only write `<temp-dir>/output.md`. Do not modify any other file.

## Final Response

After successfully saving `<temp-dir>/output.md`, output exactly:

Done.
