# Change Into Spec

Types: docs|refactor|test

## Goal

The Change workflow uses `spec` terminology and the canonical `.mch/default/prompts/spec-file-structure.md` template everywhere, with no remaining repository or Codex skill references to the obsolete spec-structure prompt.

## Scope

- Replace the legacy Change spec structure prompt path with `.mch/default/prompts/spec-file-structure.md` across repository prompts, scripts, Makefiles, CLI workflow code, tests, and Codex skill instructions.
- Remove the obsolete spec-structure prompt file as the final implementation step.
- Clarify Change workflow terminology so Change is the overall delivery flow, Idea and Spec are artifacts in that flow, and `<change-slug>` names the branch/file slug placeholder.
- Update documentation and tests that materially depend on the old prompt path, old placeholder name, or ambiguous Change/Idea/Spec wording.

## Requirements

- The canonical Change spec structure template must live at `.mch/default/prompts/spec-file-structure.md`.
- The implementation must update every tracked reference to obsolete spec-structure template paths so they use `.mch/default/prompts/spec-file-structure.md`.
- The implementation must not leave the obsolete spec-structure prompt file present in the repository after all consumers use the canonical `.mch/default/prompts/spec-file-structure.md` path.
- Workflow wording must define a Change as the full flow made of artifacts such as branch, idea, spec, docs, code, and PR.
- Workflow wording must preserve that a Change has a title, one or more type classifications when required by generated spec metadata, and an optional epic according to product docs.
- Workflow wording must identify the Change slug as `<change-slug>` wherever the placeholder refers to branch and artifact filenames.
- Branch, idea, and spec path examples must use `change/<change-slug>`, `agent/ideas/<change-slug>.md`, and `specs/<change-slug>.md`.
- Change, Idea, and Spec must be described as sharing the same title.
- Workflow wording must state that changing the Idea title updates the Change title, and changing the Spec title updates the Idea and Change titles, where the active workflow supports those artifact updates.
- The Change must not rename the product concept `Change`, application Change data fields, API paths, database objects, or UI labels that intentionally refer to product Changes.
- The implementation must not edit `AGENTS.md` unless the user explicitly asks for that file in a later prompt.
- The implementation must not edit files under `db/**` or mutate any database state.

## Acceptance Criteria

- The obsolete spec-structure prompt file no longer exists after implementation.
- `.mch/default/prompts/spec-file-structure.md` exists and remains the template for generated Change specs.
- Repository search finds no tracked reference to obsolete spec-structure prompt paths.
- Repository search finds no obsolete slug placeholder wording in active Change workflow prompts, scripts, Makefiles, docs, or skills where the intended meaning is the slug used in branch, idea, or spec paths.
- Active Change workflow prompts and Codex skills instruct agents to read `.mch/default/prompts/spec-file-structure.md` for the spec structure template.
- Documentation that describes repository Change workflow branches and spec files uses `change/<change-slug>` and `specs/<change-slug>.md` when referring to the slug placeholder.
- Relevant automated tests, fixtures, or golden expectations that reference the old prompt path or placeholder are updated to the new contract.
- Existing CLI and backend behavior unrelated to prompt lookup, Change workflow wording, or spec artifact paths remains unchanged.

## Non-Goals

- Renaming the product entity `Change` to `Spec`.
- Renaming backend API routes, DTO fields, database tables, stored procedures, frontend routes, or CLI screens that intentionally model product Changes.
- Implementing new Change title synchronization behavior beyond documenting and wiring the existing workflow contract where it already belongs.
- Changing allowed backend Change type slugs, Change phase behavior, epic behavior, or test case completeness behavior.
- Adding Flow assignment, Run controls, branch reconciliation, or new non-interactive `mch` commands.
- Editing `AGENTS.md` or any file under `db/**`.

## Design Notes

- Documentation is the source of truth: `Change` remains the core product vocabulary, while `idea`, `spec`, `pr`, `pr_url`, and `agent_edit` are the artifact fields used by current docs.
- The idea mentions a pluralized legacy prompt path, but the active repository contains the singular legacy prompt path; implementation should remove the actual singular-path file and clean up both singular and plural stale references if present.
- The first non-blank line of generated Change specs remains the shared Change, Idea, and Spec title, followed by the `Types: <type-slugs>` metadata line.
- Valid backend Change type slugs are backend option data. The current repository seed includes `feature`, `fix`, `refactor`, `upgrade`, `chore`, `docs`, `test`, `ci`, `security`, `migration`, `revert`, and `spike`; this spec uses `docs|refactor|test`.
- `.mch/default` is the active Flow profile for `mch`; the CLI must continue loading prompts from that committed repository configuration without compatibility fallback to old prompt locations unless a future Change explicitly adds such behavior.
- Removal of the obsolete prompt file should happen after every consumer has been repointed so the repository is not left in an intermediate broken state.

## Relevant Specs

- `specs/116-spec-file-structure.md`
- `docs/concepts.md`
- `docs/docs-rules.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`

## Verification

- `rg -n "agent/prompts/change-file-structure[.]md|agents/prompts/change-file-structure[.]md" .`
- `rg -n "<change-name[>]" .mch agent docs specs cli backend frontend Makefile`
- `git diff --check`
- `(cd cli && make lint)`
- `(cd cli && go test ./...)`
- `(cd cli && go build -o /tmp/mch ./cmd/mch)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make lint)`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`

## QA Test Cases

- Run the reference search for obsolete spec-structure prompt paths; verify there are no tracked matches after implementation.
- Open or inspect every active `change-*` Codex skill and verify spec-generation instructions point at `.mch/default/prompts/spec-file-structure.md`.
- Start `mch` from the repository root and from a nested directory; verify it loads `.mch/config.yaml` and `.mch/default` without looking for the removed obsolete prompt file.
- Generate or inspect a Change spec workflow prompt; verify it uses `<change-slug>` for branch, idea, and spec file placeholders.
- Edit an Idea title through the documented workflow; verify the wording and implementation path preserve the shared Change, Idea, and Spec title contract where that workflow applies.
- Edit a Spec title through the documented workflow; verify the wording and implementation path preserve the shared Spec, Idea, and Change title contract where that workflow applies.
- Attempt to run any path that previously referenced the removed prompt; verify it either succeeds using `.mch/default/prompts/spec-file-structure.md` or reports a path-specific error that names the new canonical prompt location.
- Confirm no database files are edited and no database operation is required for this Change.

## Review Focus

- Search completeness across prompts, scripts, Makefiles, docs, tests, CLI code, and Codex skill files for stale prompt paths.
- Whether slug placeholder wording was replaced only where it meant the branch and artifact slug, without corrupting unrelated prose or active product terminology.
- Preservation of the `Change` product concept and current `idea`/`spec` artifact model from docs.
- CLI `.mch/default` prompt loading and tests that could silently continue using old prompt paths.
- The final deletion of the obsolete prompt file after all consumers have been migrated.

## Follow-Ups

- Consider a later Change that audits older archived specs and research notes for obsolete Change workflow wording if those documents need to stay current.
