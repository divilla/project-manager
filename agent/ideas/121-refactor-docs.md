# Refactor Docs

Types: refactor|docs

## Goal

- Code is the repository's single source of truth. The Idea and initial code originate a Change,
  the Spec expresses its desired future state and strictly instructs what needs to change and how
  the Change must be conducted, and the PR summarizes what changed and why.
- Behavioral contracts live in executable sources instead of duplicated prose, while durable
  decisions, genuine operational runbooks, research, and Change artifacts remain available.

## Scope

- OTH - Update repository agent instructions with the code-only authority model and remove
  product-behavior prose that duplicates executable sources.
- DOC - Replace the behavioral documentation tree with a concise README, durable ADRs, and only
  genuine operational runbooks.
- SP - Remove documentation and post-code documentation stages and prompts from the default Change
  workflow, and make implementation and review inspect code, tests, configuration, and the branch
  diff directly.
- OTH - Preserve every Idea, Spec, PR artifact, and research document as historical context rather
  than current behavioral authority.
- OTH - Replace important rules currently carried only by deleted prose with tests, lint rules,
  typed configuration, command help, or other executable enforcement at the owning source.
- SP - Preserve the applied removal of the hardcoded stage allowlist from the Codex resume script;
  configured Flow stages, rather than a duplicate shell list, define valid stage names.

## Requirements

### Authority model

- OTH - Code is the single source of truth for current behavior and technical contracts throughout
  the Change lifecycle.
- OTH - The Idea and the code that exists when the Change begins are the sources of the Change. The
  Idea supplies desired direction, and the initial code supplies the starting behavior and
  constraints.
- OTH - The active Spec is the Change specification. It expresses the desired future state and
  gives strict, implementation-ready instructions for what needs to change and how the Change must
  be conducted.
- OTH - The Spec directs the intended Change but does not override code as the description of
  current behavior. A difference between the Spec and initial code is an instruction to change the
  code, not evidence that the current code already behaves as specified.
- OTH - Implementation must follow every Spec instruction or explicitly resolve the discrepancy;
  omitted or deferred instructions must not be presented as implemented.
- OTH - Every delivered behavior and technical contract must be represented in code; a statement in
  an Idea, Spec, PR, test, or document cannot establish current behavior that the code does not
  provide.
- OTH - The PR is a Change summary and reflection. It summarizes what the code changed, why it
  changed, verification evidence, and any Spec instructions that were intentionally deferred.
- OTH - The PR, Idea, and Spec never override code as the source of current behavior.
- OTH - Merged Ideas, Specs, and PRs remain historical artifacts and are not reconciled continuously
  after merge.
- DOC - Behavioral documentation is not a source of truth. When retained prose conflicts with code,
  the code wins.

### Repository instructions

- OTH - `AGENTS.md` retains safety boundaries, coding rules, testing rules, and agent workflow
  instructions.
- OTH - `AGENTS.md` states the authority hierarchy above without describing current product screens,
  endpoints, payloads, or other behavior already represented in code and tests.
- OTH - `AGENTS.md` requires agents to implement the Spec's desired future state according to its
  strict instructions, verify the resulting code, and make the PR summary explain what changed and
  why.
- DOC|OTH - Agents may update documentation only when the user explicitly requests documentation
  work or when an active Change explicitly includes the documentation edit.
- OTH - Repository instructions no longer require a post-code pass that rewrites Specs or docs to
  mirror implementation. Any Spec/code discrepancy must instead be resolved explicitly or recorded
  in the PR as a deferred instruction.

### Surviving documentation

- DOC - `README.md` contains only the product purpose, a minimal quick start, a concise repository
  layout, and links to executable setup, run, and verification commands.
- DOC - `README.md` does not duplicate API payloads, screen behavior, workflow state machines, or
  other contracts available from executable sources.
- DOC - `docs/decisions/` contains only ADR-style records that explain durable choices and
  trade-offs which code cannot explain by itself.
- DOC - ADR coverage includes the repository-wide POST mutation convention, the prohibition on
  database foreign keys, and the choice of Bubble Tea for the CLI.
- DOC - Files retained under `docs/operations/` cover only deployment, recovery, credentials,
  external dependencies, or unusual failure handling that cannot be encoded safely in Makefiles,
  configuration, command help, or CI.
- DOC - All files under `docs/research/` remain unchanged as non-authoritative research history.
- OTH - All files under `agent/ideas/`, `agent/prs/`, and `specs/` remain preserved as historical
  Change artifacts.

### Behavioral documentation removal

- DOC - Remove these manually maintained behavioral references after their required rules have been
  moved to an owning executable source or durable ADR:
  - `docs/architecture/backend-api.md`
  - `docs/architecture/frontend-spa.md`
  - `docs/architecture/cli.md`
  - `docs/architecture/mch.md`
  - `docs/architecture/system-architecture.md`
  - every file under `docs/functionality/`
  - `docs/concepts.md`
  - `docs/product-overview.md`
  - every file under `docs/plans/`
  - `docs/docs-rules.md`
  - `docs/operations/verification.md`
- DOC - Move the short product vision from `docs/product-overview.md` into `README.md` before
  removing the source file.
- DOC - Replace executable local-development and verification prose with links to Make targets,
  command help, or CI; retain only qualifying operational material under `docs/operations/`.
- DOC|FE - Remove `frontend/ARCHITECTURE.md` only after every enforceable rule it uniquely carries
  has an owning lint rule, dependency check, or automated test.
- DOC - Do not replace deleted behavioral documents with renamed documents that restate the same
  code-derived contracts.

### Contracts at their executable sources

- BE - Backend routes, request and response shapes, validation, and persistence behavior are owned
  by handlers, DTOs, services, and repositories; API integration tests verify that code.
- CLI|SP - CLI commands, stage membership, prompt selection, and workflow behavior are owned by Go
  code, `.mch/default/flow.yaml`, `.mch/default/help.yaml`, prompts, and program integration tests.
- FE - Frontend behavior is owned by components, stores, and router configuration; component tests
  verify that code.
- OTH - Package-boundary rules are enforced through lint or dependency checks rather than prose.
- OTH - Verification behavior is exposed through Makefiles and CI rather than a duplicated command
  catalog in documentation.
- OTH - Configuration contracts are owned by typed structures, defaults, validation, and `--help`
  or equivalent executable output.
- BE - Domain rules are owned by service code; named behavioral tests verify those rules.
- DB|BE - Database contracts are represented by database and repository code and verified by
  integration tests while all database work remains subject to the repository safety boundary.
- OTH - Comments explain non-obvious invariants and reasons; they do not narrate code behavior.
- OTH - Before deleting prose that carries an important rule, add the narrowest practical
  executable enforcement and a named regression test for that rule.

### Change workflow and prompts

- SP - The default Flow has no documentation-writing or post-code documentation-reconciliation
  stage, step, hook, or active prompt reference.
- SP - Remove default prompts used only by those deleted stages, including documentation update and
  post-code Spec/docs reconciliation prompts.
- SP - Implementation prompts treat the Spec as the desired future state and as strict instructions
  for what needs to change and how to conduct the Change, inspect the initial code and tests before
  editing, and preserve code as the single source of truth.
- SP - Review prompts compare the final code diff and tests with the Spec instructions, report
  omitted instructions or unintended behavior, and use code to determine current behavior.
- SP - PR-writing prompts summarize and reflect what the complete accepted code diff changed, why
  it changed, verification evidence, and deferred Spec instructions without presenting the PR as
  behavioral authority.
- SP - Spec-writing and Spec-review prompts derive the Change from the Idea plus initial code and
  express the desired future state as strict instructions for what needs to change and how to
  conduct the Change.
- SP - Prompt and Flow integration tests fail when an active Flow entry references a removed prompt
  or documentation stage.
- SP - `codex-resume-session.sh` requires a non-empty `MCH_STAGE` but does not maintain a hardcoded
  allowlist; the configured Flow owns stage-name validity.

## Non-Goals

- Changing backend, frontend, or CLI product behavior while relocating or enforcing its existing
  contract.
- Deleting Ideas, Specs, PR artifacts, or research documents.
- Treating tests, comments, ADRs, README content, runbooks, Ideas, Specs, or PR descriptions as a
  replacement for code as the single source of current behavior.
- Continuously updating historical Specs after their Change has merged.
- Moving active Change planning into `docs/plans/`; Changes and Epics remain the planning system.
- Modifying external installed skills as part of the repository diff.

## Design Notes

### Change artifact roles

| Artifact | Role |
| --- | --- |
| Code | The single source of truth for current behavior and technical contracts. |
| Idea + initial code | Originate the Change by supplying its direction and starting state. |
| Spec | Expresses the desired future state and strictly instructs what to change and how. |
| PR | Summarizes and reflects what changed and why. |
| Merged artifacts | Preserve historical context without overriding code. |

- “Single source of truth” refers only to code. Idea, Spec, and PR describe the Change from
  different perspectives but never compete with code as the authority for current behavior.
- A Spec instruction that is intentionally deferred remains visible in the PR summary rather than
  being silently removed or represented as implemented.

### Documentation disposition

- Keep prose only when it explains why a durable decision exists or provides operational knowledge
  that is unsafe or impractical to encode.
- A deleted behavioral document may be consulted while locating contracts, but its wording does not
  override executable behavior.
- `README.md` is an entry point, not a comprehensive product contract.
- The existing documentation tree may be removed in one Change because preservation of research and
  artifacts is explicit and each behavioral rule must be checked for executable ownership first.

### Workflow boundary

- Repository-local default prompts are part of this Change because they ship in the prospective PR.
  Installed skills outside the repository require a separate distribution/update operation and are
  tracked as a Follow-Up.
- The removed hardcoded stage allowlist is accepted branch behavior because stage definitions are
  configuration data. The script still rejects a missing stage value and fails naturally when the
  resulting workspace or session resource does not exist.

## Verification

- `git diff --check`
- `rg -n "docs-update|code-docs-spec-update|change-docs|change-docs-post-code" AGENTS.md .mch`
- `rg -n "source of truth|authoritative|PR contract|documentation.*contract" README.md AGENTS.md \
  .mch/default`
- `find docs -type f -not -path "docs/research/*" -print | sort`
- `(cd backend && make check)`
- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`
- `(cd cli && make test)`
- `(cd cli && make race)`
- `(cd cli && make integration-test)`
- No successful verification run is claimed by this provisional Spec.

## QA Test Cases

- OTH - Build a Spec from an Idea and initial code and verify it expresses the desired future state
  as strict instructions for what needs to change and how to conduct the Change, without presenting
  itself as current behavior.
- OTH - After implementation, compare the accepted code diff, tests, and PR and verify the PR
  accurately summarizes what changed, why it changed, and any deferred instructions without
  overriding code.
- OTH - Inspect a merged Idea, Spec, and PR and verify each remains available as historical context
  but repository instructions do not treat it as current behavior.
- DOC - Inspect the surviving non-research documentation and verify it contains only the concise
  README, durable ADRs, and qualifying operational runbooks.
- DOC - Verify every listed behavioral document is removed only after its important enforceable
  rules have an executable owner or durable rationale has moved to an ADR.
- DOC - Verify all research, Idea, Spec, and PR artifacts remain present and unchanged except for
  files explicitly belonging to this active Change.
- DOC - Follow README setup, run, and verification links and verify they resolve to executable
  commands without duplicating their implementation details.
- OTH - Request a documentation edit without explicitly including it in the task or active Change
  and verify repository instructions prevent the agent from changing documentation.
- SP - Load the default Flow and verify it contains no documentation or post-code reconciliation
  stage and every remaining prompt reference resolves to an existing file.
- SP - Run the Flow prompt integration coverage and verify removed documentation prompt names and
  steps are rejected instead of silently resolving to stale files.
- SP - Resume a session with a configured non-empty stage name that is not in the former shell
  allowlist and verify the script uses the configured stage workspace rather than rejecting its
  name before resource validation.
- FE - Verify `frontend/ARCHITECTURE.md` is absent only when its unique enforceable rules have
  matching lint, dependency, or automated-test coverage.

## Review Focus

- Confirm every workflow and repository instruction identifies code as the single source of truth,
  the Idea plus initial code as the sources of the Change, the Spec as the desired future state and
  strict Change instructions, and the PR as a summary of what changed and why.
- Confirm important safety, coding, testing, and workflow rules remain enforceable after behavioral
  prose is removed from `AGENTS.md` and `docs/`.
- Audit every deleted document for unique rationale, operational knowledge, or unenforced rules
  before accepting its removal.
- Confirm README and ADR additions do not recreate the behavioral-documentation pipeline under new
  filenames.
- Confirm default Flow and prompt references contain no removed documentation stages or stale
  source-of-truth language.
- Confirm the dynamic stage contract has executable coverage and no script duplicates stage names
  already owned by Flow configuration.
- Confirm no application behavior changes are hidden inside documentation cleanup or enforcement
  work.

## Follow-Ups

- Update installed Change workflow skills outside this repository to use the same authority and
  documentation model.
- Decide whether completed Specs should eventually stop being committed after their content is
  captured in PR history; this Change preserves the existing artifact archive.
