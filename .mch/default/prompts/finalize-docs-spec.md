Use this only for a post-code documentation pass. The implementation already exists; the job is to reconcile the Change contract, QA test cases, and `docs/` with the final accepted observable behavior. When docs and code diverge, code wins. When the Change spec and implementation diverge, ask the user how to resolve the divergence before editing the contract or docs.

## Preconditions

Extract `<change-slug>` from the current Git branch using `^change/([0-9]+-[a-z0-9-_]+)$`.

If the current branch does not match, stop with: `Branch must be change/<change-slug>.`

If `specs/<change-slug>.md` does not exist, stop with one concise error.

## Before Editing

1. Read `specs/<change-slug>.md`.
2. Read `docs/docs-rules.md`.
3. Inspect the branch diff against `origin/stage`.
4. Identify changed observable behavior, including public contracts, backend API payloads, frontend behavior, CLI behavior, persistence behavior, validation behavior, history behavior, and verification commands.
5. Use targeted search to find related `docs/` files.
6. Read only docs that materially affect this Change.
7. Inspect the existing `QA Test Cases` section in the Change spec and compare it to the implemented behavior.
8. Identify any divergence between the Change spec and implemented behavior before editing.

Useful commands:

- `git status --short --branch`
- `git diff --stat origin/stage...HEAD`
- `git diff --name-only origin/stage...HEAD`
- `git diff origin/stage...HEAD -- <targeted paths>`
- `rg "<relevant term>" docs specs`

## Reconciliation Rules

- Treat the Change spec as the intended contract.
- Treat the implemented diff as evidence of final observable behavior, not as automatic proof that the behavior is correct.
- Resolve conflicts between existing docs and the Change spec in favor of the Change spec.
- Resolve conflicts between existing docs and implemented code in favor of implemented code, as long as the code does not diverge from the Change spec.
- When docs are stale relative to implementation, update docs to match the implemented observable behavior.
- If the Change spec and implementation diverge, stop before editing and ask the user how to resolve it. Present the specific divergence, relevant files, and concrete options such as update the Change contract/docs to match implementation, treat implementation as a bug to fix elsewhere, or record a follow-up.
- Continue after a divergence only when the user has already given an explicit resolution for that divergence in the current task.
- When the user confirms that implemented behavior is the accepted final behavior, update the Change spec first so it remains the PR contract, then update docs to match.
- If implemented behavior appears wrong, incomplete, or outside the Change scope, do not document it as intended. Stop and report the conflict with the relevant files.
- If docs and the Change spec conflict in a way that cannot be resolved from the implemented diff, stop and report the ambiguity.
- Preserve established project vocabulary.
- Keep docs concise, product-focused, and testable.
- Do not describe implementation internals unless they are part of the observable product, API, CLI, or persistence contract.
- Do not create duplicate documentation if an existing doc is the right home for the behavior.
- Keep every doc under the repository line limit from `docs/docs-rules.md`.

## QA Test Cases

Always revisit the Change spec `QA Test Cases` section after inspecting the implementation.

- Add missing QA test cases for newly implemented external behavior.
- Update stale QA test cases when final behavior, request fields, response fields, validation rules, UI behavior, history behavior, or verification expectations changed.
- Remove or rewrite QA test cases that no longer match the accepted contract.
- Keep QA test cases behavior-focused and executable by a human or future agent.
- Include important negative, validation, edge, and regression scenarios when they are part of the Change.
- Keep QA test cases scoped to this Change; record unrelated test ideas as Follow-Ups.
- Do not edit executable test files.

## Editable Scope

Edit only:

- `specs/<change-slug>.md`
- files under `docs/`

Update the Change spec only to clarify accepted final behavior, acceptance criteria, verification notes, QA test cases, review focus, follow-ups, or unresolved ambiguity discovered after implementation.

Update docs enough that future work can align code, tests, frontend, CLI, and seed data with the intended behavior without relying on chat history.

## Hard Constraints

- Do not implement code.
- Do not edit database files.
- Do not edit backend code.
- Do not edit frontend code.
- Do not edit CLI code.
- Do not edit executable tests.
- Do not edit seed files.
- Do not edit generated artifacts.
- Do not run migrations or mutate any database.
- Do not stage, commit, or push changes unless the user explicitly asks.
- Do not broaden the Change to include unrelated behavior discovered while reading the diff.
- Record unrelated work as Follow-Ups instead of documenting it as part of this Change.

## Final Response

After successfully reconciling the Change contract, QA test cases, and documentation,
output exactly:

Done.
