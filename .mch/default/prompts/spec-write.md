Do not use any skills for this request.

Write a provisional, implementation-ready Change Spec that incorporates the Idea and all relevant changes already applied to the branch.

## Files

- Input idea: `/stg-tmp-dir/input.md`
- Output spec: `/stg-tmp-dir/output.md`
- Change type options: `/def-dir/prompts/change-types.md`

## Instructions

Read, in this order:

1. `/stg-tmp-dir/input.md`
2. `/def-dir/prompts/spec-file-structure.md`
3. `/def-dir/prompts/change-types.md`
4. `AGENTS.md`, when present

### Change Type Options

Load the complete list of allowed Change type slugs from
`/def-dir/prompts/change-types.md`. Require the file to begin with `# Change Types`, followed by
one or more Markdown list items formatted as `- <slug>`. Treat those list-item values, in file
order, as the complete set of allowed Change type options.

Select one or more listed slugs that best describe the intended Change and format them according
to `/def-dir/prompts/spec-file-structure.md`. Always include a populated `Types:` line in the new
Spec. Use only slugs from the loaded file; do not use a hardcoded fallback or invent a slug.

If the file is missing, unreadable, malformed, or contains no slugs, use the Clarification Gate
instead of writing the Spec.

### Sources and Authority

Inspect existing code as the source of truth for current behavior. Inspect the complete branch state, including:

- the committed diff against the workflow base branch
- staged and unstaged diffs
- relevant untracked files
- tests and verification evidence relevant to the Change
- published PR metadata, when a PR exists

Before a PR exists, the Idea defines the initial intent and the complete branch diff is the prospective PR. Include relevant behavior already applied manually, by an agent, or by another process. Do not assume implementation has not started.

If a PR exists, treat it as authoritative and make the Spec represent its accepted behavior.

Use targeted search to inspect relevant code and tests. Read only documentation that materially affects understanding, and do not let stale docs override code or existing branch behavior.

### Writing Contract

- Use `/def-dir/prompts/spec-file-structure.md` exactly. Do not add, remove, rename, or reorder top-level sections.
- Account for every meaningful Idea item. Incorporate each item as a Goal, Scope item, Requirement, Non-Goal, Design Note, QA Test Case, Review Focus item, or Follow-Up. Ask for clarification before writing when an item cannot be classified confidently.
- Account for every relevant change already present in the branch, including work applied before the Spec was written.
- Describe intended final behavior in Requirements without adding implementation statuses, completion markers, or a change diary.
- Use existing implementation to make Requirements and Design Notes concrete, but do not assume that existing code is correct merely because it exists.
- Preserve fenced code blocks from the Idea when they remain part of the intended Change. Reformat them only as needed to meet the 100-character line limit without changing their meaning or behavior. Ask for clarification when that is not possible.
- Keep every line in the generated Spec within 100 characters.
- Do not invent product decisions, behavior, API contracts, database behavior, verification results, QA expectations, or new scope.
- Do not claim verification passed unless evidence shows it ran successfully.
- When the Idea conflicts with existing branch work, ask which behavior belongs in the prospective PR unless an accepted PR state already resolves the conflict.

### Clarification Gate

Before writing the Spec, stop and output only concise clarifying questions if:

- the scope or intended observable behavior has materially different interpretations
- existing branch work cannot be classified as part of or unrelated to the prospective PR
- the Idea and existing implementation conflict on a product decision
- API, persistence, CLI, validation, or failure behavior required for implementation is unclear
- valid Change type slugs cannot be determined
- any Idea statement is unclear, contradictory, or of uncertain intent

Use this format:

```text
Questions:
1. First question?
2. Second question?
```

Number every question and keep each question line within 100 characters.

### Quality Bar

The final Spec must satisfy:

- `Goal`: one or more observable end states intended for the Change.
- `Scope`: the intended Change plus relevant work already present in the prospective PR.
- `Requirements`: testable final behavior, visible contracts, persistence expectations, and boundaries.
- `Non-Goals`: intentionally excluded adjacent work, or `- None.`.
- `Design Notes`: established implementation decisions, constraints, and assumptions.
- `Verification`: realistic repository-supported commands with claims limited to available evidence.
- `QA Test Cases`: behavior-focused scenarios covering important paths and boundaries.
- `Review Focus`: the riskiest intended behavior and contracts.
- `Follow-Ups`: useful work outside the prospective PR, or `- None.`.

Write the completed Spec to `/stg-tmp-dir/output.md`, intentionally overwriting the file if it already exists. Do not modify any other file.

## Final Response

After successfully saving `/stg-tmp-dir/output.md`, output exactly:

Done.
