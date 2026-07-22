# Spec File Structure

Use this structure exactly for every Spec file. Replace all placeholders and instructional text with concrete Spec content. Keep every section concise, specific, and testable.

A Spec begins as a provisional implementation guide and becomes a structured representation of the final PR during reconciliation. The PR is the de facto Change and single source of truth once published; before publication, the complete branch diff is the prospective PR. Use `Spec` or `spec` for the artifact and `Change` for the PR-backed delivery unit.

For the initial Spec, combine the Idea with existing code and every relevant branch change already applied manually, by an agent, or by another process. After implementation, reconcile the Spec with every material change in the final PR. The final Spec must follow accepted PR behavior and must not override it. Do not add implementation-progress statuses or a change diary; Git and the PR preserve that history.

Do not add, remove, rename, or reorder top-level `##` sections unless the user explicitly changes the Spec workflow. Optional `###` and `####` headings may organize details within a required section.

Use nested headings when a flat list would obscure screens, flows, states, commands, keybindings, configuration, persistence, error handling, glossary terms, naming conventions, relationships, navigation, or data movement.

Use a Markdown-native diagram when a non-trivial workflow, screen flow, state transition, relationship, or data flow would be clearer visually. Prefer Mermaid fenced blocks such as `flowchart`, `sequenceDiagram`, or `stateDiagram-v2` when supported; otherwise use a concise Markdown table or ASCII diagram. A diagram must clarify, not replace, testable prose. If a non-trivial flow has no diagram, explain why in `Design Notes`.

The generated Spec must begin with one H1 title. An optional `Types:` metadata line may appear
immediately after the title. When it is present, place exactly one blank line between the title
and metadata and exactly one blank line between the metadata and Spec body. When it is omitted,
place exactly one blank line between the title and Spec body. Do not place blank lines or other
content before the H1 title.

Format populated metadata as `Types: <type-slugs>`. Join multiple slugs with `|` and no spaces,
as in `Types: feature|fix|test`. `Types:` with an empty value is also structurally valid. Structural
validation checks only the line's placement and format; it does not restrict which slugs may
appear.

Use these subsystem tags where instructed:

- `DOC` - documentation
- `DB` - database
- `BE` - backend
- `FE` - frontend
- `CLI` - CLI
- `SP` - scripts and prompts
- `OTH` - work that does not fit another tag

Format a tagged item as `- <TAG> - <statement>`. When one item spans multiple subsystems, join its tags with `|` and no spaces, as in `- BE|CLI - <statement>`.

Do not wrap the generated Spec in a code block.

# <Spec Title>

Types: feature|fix|test

## Goal

Describe the observable end state or end states this Spec must define. Use multiple goal statements when the Spec has distinct required outcomes. Do not write a list of implementation tasks.

## Scope

- List the behavior, documentation, architecture, and implementation areas included in the provisional Change or final PR.
- Keep every item directly tied to the Goal.
- During final reconciliation, account for every material PR change and exclude adjacent work not present in the PR.

## Requirements

- State testable obligations the implementation must satisfy. During final reconciliation, every obligation must be supported by the PR.
- Include expected behavior, important boundaries, validation, and failure handling where relevant.
- Prefix each subsystem-specific requirement with all applicable subsystem tags. Leave an item untagged only when it is genuinely cross-cutting.
- Write behavioral requirements, not task summaries. For example:
  - `- CLI - The Change details view displays the Ref UUID label and value when a value is present.`
  - `- SP - The session restore script resumes the selected session using the configured default and temporary directories.`

## Non-Goals

- List related work that is intentionally outside the Spec.
- Include decisions that prevent accidental scope expansion.
- Move useful but non-essential work here or to `Follow-Ups` instead of expanding Scope.
- Use `- None.` when there are no non-goals.

## Design Notes

- Record implementation decisions already evidenced by code, plus constraints, data model assumptions, UX details, and workflow rules needed to guide remaining work. During final reconciliation, make these notes match the PR.
- Use nested headings for a glossary, naming conventions, state or flow models, screen design, data shapes, compatibility, or other structured detail when useful.
- Link to relevant documentation instead of repeating it, and ensure documentation follows accepted PR code.
- State assumptions that reviewers and future agents must preserve.

## Verification

- List the repository-supported commands needed to verify every affected area. During final reconciliation, make result claims match available PR evidence.
- Include relevant lint, unit test, race test, API-test, typecheck, and build commands.
- Do not invent commands or include commands unrelated to the Spec.

## QA Test Cases

- List manual or product-level scenarios, not implementation tasks or automated Verification commands.
- Cover applicable happy paths, validation failures, backend or external-command failures, cancellation or no-op behavior, persistence, and boundary cases.
- Prefix every QA test case with one or more subsystem tags. Use `OTH` when no more specific tag applies. For example:
  - `- CLI - Open a Change that has a Ref UUID and verify the details view shows the expected label and value.`
  - `- SP - Restore a saved session and verify the script uses the configured directories and resumes the selected session.`

## Review Focus

- Identify the riskiest or most subtle areas reviewers should inspect first.
- Highlight changed contracts, data flow, persistence, migrations, concurrency, security, generated artifacts, or workflow automation when relevant.

## Follow-Ups

- List useful future work that is outside the Spec.
- Use `- None.` when there are no known follow-ups.
