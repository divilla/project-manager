# Pull Request

## Goal
What user/system outcome this PR delivers.

## Scope
What is included.

## Requirements
- Requirement 1
  - Acceptance: ...
- Requirement 2
  - Acceptance: ...

## Non-Goals
What this PR intentionally does not change.

## Specifications
- Relevant API contracts
- Relevant architecture decisions
- Any module boundaries touched

## Verification
- `make test`
- `make api-test`
- `pnpm --dir frontend test`
- Manual checks, if any

## Risks / Review Focus
- Areas where reviewer attention is useful

Example:

## Goal
Add dedicated frontend task workflows for task search, detail, creation, editing, deletion confirmation, requirements editing, and markdown-rendered task descriptions.

## Requirements Implemented
- Task board search supports name, type, and phase filters.
- Task detail route `/tasks/:id` renders task hierarchy, children, requirements, and markdown descriptions.
- Task create route `/tasks/create/:parentId` supports root and child task creation.
- Task edit route `/tasks/edit/:taskId` supports name, type, phase, and markdown description editing.
- Task and requirement deletion use persistent confirmation dialogs.
- Backend renders sanitized markdown into `description_html`.

## Non-Goals
- No schema changes.
- No changes to task type or task phase reference data.
- No backend auth changes.

## Design Notes
- Backend markdown rendering is isolated behind `backend/pkg/markdown`.
- `description` remains raw markdown; `description_html` is rendered/sanitized HTML.
- Project-scoped task data is cached in `taskCache.store`.

## Verification
- `make test`
- `make api-test`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`

## Review Focus
- Project-switch behavior on nested task routes.
- Create/edit route behavior when opened directly.
- Markdown sanitization and rendering boundaries.

The important practice is: don’t make the PR description just a change log. Make it traceable:

> Goal → requirements → implementation boundaries → verification.

For AI-assisted PRs especially, I’d add one more section:

## Specification References
- `agent/features/08-frontend-tasks.md`
- `agent/specs/...`

That lets reviewers compare the implementation against the intended behavior instead of reverse-engineering intent from the diff.



That reaction is correct. Real PRs often do not fit a clean single taxonomy.

A hard taxonomy like feature | fix | refactor | docs is useful for reporting, but weak as a model of actual work. Most meaningful PRs contain a bundle:

- feature behavior
- bug fixes discovered while implementing it
- refactors needed to make the change fit
- tests
- docs
- build/config adjustments

So instead of forcing one type, model the PR around primary intent plus change facets.

Example:

Work item
  primary_intent: feature
  facets:
    - backend
    - frontend
    - refactor
    - docs
    - tests

Or:

PR
  kind: feature
  includes:
    - fix
    - refactor
    - docs
    - test

The important distinction is:

> A PR has one reason for existing, but may contain many kinds of changes.

For example:

Primary intent:
Add task detail page.

Includes:
- Feature: task detail route
- Feature: requirement editing
- Fix: stale project refresh behavior
- Refactor: task cache store
- Docs: frontend architecture notes
- Tests: route and store tests

That is much more honest than pretending the PR is only feature.

A better model than hard taxonomy:

  Field           Meaning
━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Intent          Why this PR exists
──────────────  ───────────────────────────────
  Scope           What it changes
──────────────  ───────────────────────────────
  Facets          Categories of included work
──────────────  ───────────────────────────────
  Risk areas      What reviewers should inspect
──────────────  ───────────────────────────────
  Verification    How correctness was checked

For planning states, you can still use:

phase: backlog | in_progress | staging | closed

But for type, avoid making it too rigid. Use either:

intent: feature
facets: [fix, refactor, docs, tests]

or skip type entirely and use tags:

tags: [feature, frontend, backend, markdown, routing, tests]

My recommendation:

- Keep phase strict.
- Keep requirements structured.
- Keep type/tags flexible.

The core unit should not be “a feature PR” or “a fix PR.” It should be:
> a PR-sized change with explicit intent, requirements, and verification.

> A PR has one reason for existing, but may contain many kinds of changes.

For example:

Primary intent:
Add task detail page.

Includes:
- Feature: task detail route
- Feature: requirement editing
- Fix: stale project refresh behavior
- Refactor: task cache store
- Docs: frontend architecture notes
- Tests: route and store tests

That is much more honest than pretending the PR is only feature.

A better model than hard taxonomy:



# How to handle changes

Treat it as controlled evolution, not as noise and not as immutable scripture.

During development, a PR’s scope and docs can change. The important question is whether the change affects intent, requirements, acceptance criteria, or review risk.

I’d use this rule:

> If the change affects what reviewers must understand or what “done” means, record it. If it is just wording or local clarification, simply update it.

Practical policy:

  Change                                                     Treatment
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Typo, wording, formatting                                  Edit in place, no history note
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Clarifies existing requirement without changing meaning    Edit in place, maybe mention in PR comment if important
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Adds/removes requirement                                   Record in PR description changelog or decision log
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Changes acceptance criteria                                Record explicitly
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Changes non-goals/scope                                    Record explicitly
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Changes architecture/design direction                      Record as a decision
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Discovers follow-up work                                   Add follow-up item, do not silently expand PR unless necessary
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Fixes bug found during implementation                      Add to PR scope/facets if included
─────────────────────────────────────────────────────────  ────────────────────────────────────────────────────────────────
  Major pivot                                                Consider closing/superseding PR or splitting

For AI-assisted development, I’d keep a lightweight section in the PR:

## Scope Changes / Decisions

- 2026-06-25: Added backend markdown rendering because frontend-only parsing would duplicate sanitization rules.
- 2026-06-25: Deferred task reparenting UI to a follow-up PR.
- 2026-06-25: Included task cache store refactor because nested task routes needed shared project task state.

That gives you enough history without turning development into bureaucracy.

For repository docs/spec files, same idea:

- If the doc is the current specification, update it to match the current truth.
- If the change is a meaningful decision, record why somewhere durable.
- Avoid preserving every intermediate draft unless it explains a decision.

A useful split:

Current spec: what we now intend.
Decision log: why meaningful changes happened.
Git history: exact textual history if needed.

So I would not manually save every version. Git already gives you raw history. What you need manually is semantic history: the important decisions.

Best practice:

1. Update the PR description and docs to current truth.
2. Add a short decision/scope-change note for meaningful changes.
3. Keep rejected/deferred work as explicit follow-up items.
4. Do not preserve trivial edits.

In short:

> Change the docs freely, but record meaning-changing decisions.



## How to treat comments

Treat reshaping comments as change requests, not ordinary discussion.

A useful workflow:

1. Classify the comment
    - blocker: PR should not merge without it.
    - scope change: changes what the PR is supposed to deliver.
    - follow-up: valid, but belongs in another PR.
    - clarification: needs explanation, not code.
    - nit: optional cleanup.

2. Decide explicitly
    Do not let comments silently expand the PR. For each meaningful comment, decide:
    - apply in this PR
    - defer to follow-up
    - reject with reason
    - split PR
    - update docs/spec only

3. Update the PR contract
    If the comment changes behavior, scope, acceptance criteria, or architecture, update the PR description/spec.

4. Record the decision
    Add a short note:

    ### Review Decisions

    - Comment: Direct edit URLs can mutate tasks outside selected project.
      Decision: Fix in this PR because it violates the project-switch guardrail.
      Change: Add mismatch handling to edit route.

5. Link work to the comment
    When resolved, reply with something concrete:

    Fixed in commit abc123. Edit route now checks project mismatch before enabling save.
    Added regression test for pasted cross-project edit URLs.

For AI coding, the key rule is:

> A review comment may change the PR, but it should not silently change the PR’s definition of done.

If the comment reshapes the PR substantially, move the PR back from staging to in progress.

Your flow could be:

  Comment Impact             State Change
━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Minor fix/nit              stays in staging
─────────────────────────  ───────────────────────────────────────────
  Bug found                  staging → in progress until fixed
─────────────────────────  ───────────────────────────────────────────
  New requirement            staging → backlog decision or in progress
─────────────────────────  ───────────────────────────────────────────
  Architecture change        staging → in progress, update spec
─────────────────────────  ───────────────────────────────────────────
  Out-of-scope suggestion    create follow-up backlog item

The best practice is to maintain two sections:

## Current Scope
What this PR now promises.

## Review Decisions
Meaningful changes made during review.

That keeps the PR honest as it evolves.
