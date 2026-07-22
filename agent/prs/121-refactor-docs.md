# Refactor Docs

## Summary
- Establish code as the single source of truth for current behavior and technical contracts, with
  Ideas and initial code originating Changes, Specs defining the desired future state, and PRs
  summarizing accepted changes and deferred instructions.
- Replace duplicated behavioral documentation with a concise repository entry point, durable
  architectural decisions, and a database recovery runbook while preserving research and Change
  artifacts as historical context.
- Remove documentation-writing and post-code reconciliation from the default Change workflow and
  update repository prompts to inspect code, tests, configuration, and branch diffs directly.

## Behavior
- The default Flow no longer exposes the removed documentation or synchronization stages, prompts,
  or helper script, and integration coverage rejects their stale prompt and target names.
- Default Codex session scripts and the default Makefile now require a non-empty `MCH_STAGE` without
  duplicating configured stage membership; unknown non-empty stages continue to normal
  path-specific resource validation.
- `spec-review` and `spec-review-*` targets route through the dedicated `spec-review` workspace
  instead of inheriting the broader Idea-stage assignment.
- Frontend lint rules enforce shared-layer import direction and restrict direct network calls to
  feature API modules and shared infrastructure, with Vitest coverage for allowed and rejected
  cases.
- CLI integration coverage enforces package direction and verifies Flow prompt resolution, removed
  stages, dynamic stage handling, missing-stage errors, and Spec-review workspace routing.

## Docs
- Rewrite `README.md` as a concise product, quick-start, repository-layout, and executable-command
  entry point.
- Add ADRs for code-first documentation authority, Flow-owned stage names, POST API actions, the
  no-foreign-key rule, and the Bubble Tea CLI choice.
- Retain only the database recovery runbook under `docs/operations/` and remove behavioral,
  architecture, obsolete planning, and duplicated verification prose.

## Verification
- `git diff --check origin/stage...HEAD -- . ':(exclude)agent/ideas/122-refactor-idea.md'` — passed
  for the intended Change 121 diff.
- `rg -n "docs-update|code-docs-spec-update|change-docs|change-docs-post-code" AGENTS.md .mch` —
  produced no matches.
- `find docs -type f -not -path "docs/research/*" -print | sort` — listed the five ADRs and the
  database recovery runbook as the complete surviving non-research `docs/` tree.
- `git diff --check origin/stage...HEAD` — failed on the explicitly excluded
  `agent/ideas/122-refactor-idea.md` with `new blank line at EOF`; remove that unrelated artifact
  from the branch before opening this PR.
- Backend, frontend, and CLI verification suites: Not run during PR drafting.

## References
- `specs/121-refactor-docs.md`
