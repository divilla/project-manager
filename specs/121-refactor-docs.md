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
- SP - Remove hardcoded stage allowlists from every default Codex session script and the default
  Makefile; configured Flow stages define valid stage names.
- SP - Correct the default Makefile stage export so Spec review targets use the `spec-review`
  workspace instead of inheriting the `idea` stage from the broader Spec target pattern.

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
- DOC - ADR coverage includes the code-first documentation authority model, Flow-owned stage names,
  the repository-wide POST mutation convention, the prohibition on database foreign keys, and the
  choice of Bubble Tea for the CLI.
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
- OTH - The contract disposition inventory in Design Notes is exhaustive for normative rules in
  every file scheduled for deletion. Each contract family must reach its listed final owner and
  coverage before its source is removed.
- OTH - If implementation discovers a material rule that does not fit an inventory entry, preserve
  its source and report the missing disposition instead of deciding its importance or destination
  during implementation.

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
- SP - Every default Codex session script and `.mch/default/Makefile` requires a non-empty
  `MCH_STAGE` but does not maintain a hardcoded stage allowlist; the configured Flow owns stage-name
  validity.
- SP - A non-empty stage outside the former allowlist proceeds to normal workspace, artifact,
  prompt, or session validation. Missing resources fail with the existing path-specific error
  instead of an `invalid MCH_STAGE` error.
- SP - The exact `spec-review` target and every `spec-review-*` target export
  `MCH_STAGE=spec-review`; other `spec-*` targets continue to export `MCH_STAGE=idea`.
- SP - Spec review commands create, read, and resume only the
  `.mch/tmp/<ref-uuid>/spec-review` workspace and never use the Idea-stage workspace because of
  overlapping Make target patterns.

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

### Contract disposition inventory

The entries below classify every normative contract family in the documents scheduled for removal.
Descriptive examples and inventories inherit the disposition of their contract family.

#### Authority, workflow, and artifact rules

- Sources: `docs/docs-rules.md`, `docs/functionality/agent-interaction.md`, and
  `docs/functionality/pr-integration.md`.
- Final owner: `AGENTS.md`, `.mch/default` Flow configuration, scripts, and active prompts.
- Durable rationale: `docs/decisions/code-first-documentation.md` and
  `docs/decisions/flow-stage-ownership.md`.
- Coverage: `TestDefaultFlowPromptAndStageIntegration` and the workflow branch-grammar integration
  tests verify prompt resolution and removal of documentation and post-code reconciliation stages.

#### Backend API, domain, history, and persistence behavior

- Sources: `docs/architecture/backend-api.md`, `docs/architecture/system-architecture.md`,
  `docs/concepts.md`, `docs/functionality/change-lifecycle.md`,
  `docs/functionality/history.md`, and `docs/functionality/requirements-and-acceptance.md`.
- Final owner: backend handlers, DTOs, services, repositories, database source, and configuration.
- Coverage: backend service tests and API integration tests for projects, epics, changes, test
  cases, health, options, validation, history transactions, completeness, and error responses.
- Disposition: delete the prose without recreating it; code determines behavior when prose is stale.

#### Frontend product behavior

- Sources: `docs/architecture/frontend-spa.md`,
  `docs/functionality/current-project-context.md`, and the Projects Area and Change Specs sections
  of `frontend/ARCHITECTURE.md`.
- Final owner: frontend components, stores, composables, router configuration, and API modules.
- Coverage: existing component tests plus `projectSelection.store.test.ts`,
  `projectChangeRedirect.test.ts`, `useProjectsPage.test.ts`, and the direct-entry Change page
  tests; add missing cases only when the audit identifies an executable branch without coverage.
- Disposition: delete product-behavior prose after those tests pass; do not change product behavior.

#### Frontend dependency and network boundaries

- Source: the Boundaries section of `frontend/ARCHITECTURE.md`.
- Final owner: `frontend/eslint.config.js` restrictions for import direction and direct network use.
- Coverage: `frontend/src/architecture.test.ts` exercises the ESLint configuration against allowed
  and forbidden in-memory files, proving the import and direct-network restrictions reject invalid
  code rather than merely observing that the current tree contains no violation.
- Allowed boundary: feature API modules and shared HTTP infrastructure may perform network calls;
  pages may import features and shared code, and features may import shared code.

#### CLI behavior, configuration, and rendering

- Sources: `docs/architecture/cli.md`, `docs/architecture/mch.md`, and CLI-specific material in
  `docs/functionality/current-project-context.md`.
- Final owner: CLI Go code, typed configuration, command help, and `.mch/default` resources.
- Coverage: CLI model, client, program integration, and PTY terminal tests cover applicable state,
  command, persistence, external-process, configuration, rendering, and failure behavior.
- Disposition: delete the prose without changing behavior or recreating a CLI reference manual.

#### CLI package boundaries

- Source: the Package Boundaries section of `docs/architecture/mch.md`.
- Final owner: Go `internal` visibility plus an automated dependency check for repository-specific
  package direction and feature-to-app import prohibitions.
- Coverage: `TestCLIPackageBoundaries` in `cli/integration/architecture_test.go` inspects Go imports
  and fails when a feature package imports `internal/app` or code crosses another prohibited
  package boundary from the deleted rules.

#### Routine setup and verification

- Sources: `docs/operations/verification.md`, routine commands in
  `docs/operations/local-development.md`, and testing instructions in
  `frontend/ARCHITECTURE.md` and `docs/architecture/mch.md`.
- Final owner: repository, backend, CLI, and frontend Makefiles; frontend package scripts; CI; and
  command help.
- Coverage: README links resolve to those executable entry points, and repository checks exercise
  every affected test tier without copying their command catalogs into prose.

#### Database safety and recovery

- Sources: database safety and backup material in `docs/operations/local-development.md` and
  `docs/functionality/agent-interaction.md`.
- Final owner: the `AGENTS.md` database hard boundary and the no-foreign-key ADR.
- Operational destination: move destructive backup, restore, target-selection, and recovery
  knowledge that cannot be encoded safely into `docs/operations/database-recovery.md`.
- Coverage: inspect the surviving runbook for explicit disposable-target and no-production-restore
  warnings; this Change must not edit database files or execute database operations.

#### Product vision and durable decisions

- Sources: `docs/product-overview.md`, the Flow-owned stage-name rule, the POST mutation convention
  in architecture prose, the no-foreign-key rule, and the Bubble Tea choice in CLI architecture
  prose.
- Final owner: the concise product-purpose section in `README.md` and focused ADRs under
  `docs/decisions/` for documentation authority, Flow stage ownership, POST mutations, foreign
  keys, and Bubble Tea.
- Coverage: inspect README and the five ADRs for the required purpose or rationale without copied
  payloads, screens, state machines, or other behavioral contracts.

#### Obsolete planning and descriptive prose

- Sources: every file under `docs/plans/` plus screen inventories, examples, current coverage lists,
  prototype descriptions, and future-work statements in the deleted behavioral documents.
- Final owner: active Ideas, Specs, Epics, code, tests, or no owner when the statement is obsolete.
- Coverage: confirm no active requirement or operational instruction depends on the removed prose;
  do not migrate obsolete plans into README, ADRs, runbooks, or renamed behavioral documents.

### Workflow boundary

- Repository-local default prompts are part of this Change because they ship in the prospective PR.
  Installed skills outside the repository require a separate distribution/update operation and are
  tracked as a Follow-Up.
- Flow stage names are configuration data. Every default Codex session script and the default
  Makefile treats `MCH_STAGE` as an opaque, required name and does not duplicate Flow membership.
- The scripts still reject a missing stage value and fail naturally when the resulting workspace,
  artifact, prompt, or session resource does not exist.
- The explicit `spec-review` and `spec-review-*` Make target assignments take precedence over the
  broader `spec-*` assignment. This target selection is routing to a configured stage, not a stage
  membership allowlist.

## Verification

- `git diff --check`
- `rg -n "docs-update|code-docs-spec-update|change-docs|change-docs-post-code" AGENTS.md .mch`
- `rg -n "source of truth|authoritative|PR contract|documentation.*contract" README.md AGENTS.md \
  .mch/default`
- `find docs -type f -not -path "docs/research/*" -print | sort`
- `(cd backend && make check)`
- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `pnpm --dir frontend lint:check`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`
- `(cd cli && make check)`
- `(cd cli && make test)`
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
- SP - Exercise every default Codex session entry point and the default Makefile with a configured
  non-empty stage outside the former allowlist and verify each reaches its normal stage-specific
  workspace or resource behavior without returning `invalid MCH_STAGE`.
- SP - Exercise every default Codex session entry point and the default Makefile without
  `MCH_STAGE` and verify each rejects the missing value; use a non-empty stage with missing
  resources and verify the scripts return their existing path-specific resource errors.
- SP - Run the exact `spec-review` target and a prefixed target such as `spec-review-chat` and
  verify both export `MCH_STAGE=spec-review` and use the Spec review workspace; run a non-review
  `spec-*` target and verify it continues to use `MCH_STAGE=idea`.
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
- Confirm the dynamic stage contract has executable coverage and no default session script or
  Makefile duplicates stage names already owned by Flow configuration.
- Confirm Make target-specific variable precedence cannot route `spec-review` or
  `spec-review-*` through the broader Idea-stage assignment.
- Confirm no application behavior changes are hidden inside documentation cleanup or enforcement
  work.

## Follow-Ups

- Update installed Change workflow skills outside this repository to use the same authority and
  documentation model.
- Decide whether completed Specs should eventually stop being committed after their content is
  captured in PR history; this Change preserves the existing artifact archive.
