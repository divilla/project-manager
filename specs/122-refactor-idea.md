# Refactor Artifact Terminology and Retain Initial Flow Configuration

Types: feature|refactor|migration|docs

## Goal

- The Change artifact formerly named `idea` is named `def` in machine-facing contracts,
  `definition` in prose, and `define` when used as a verb.
- `def`, `spec`, and `pr` remain three separate artifacts, and all three are edited and written in
  the shared workflow stage named `artifact`.
- Database, backend, frontend, CLI, scripts, prompts, files, and current documentation expose one
  consistent artifact and stage naming contract without changing existing artifact behavior.
- The expanded initial Flow graph already committed on the branch remains available as
  configuration-only work without adding runtime support for its unfinished fields or resources.

## Scope

- Rename the definition artifact across active database, backend, frontend, and CLI contracts,
  implementations, fixtures, and tests.
- Rename the former `idea` stage to `artifact` across active CLI workspaces, environment values,
  Make targets, scripts, prompts, configuration, and tests.
- Include and reconcile the database, backend DTO, Flow, prompt, script, Spec, and directory work
  already committed on the branch.
- Retain the expanded default Flow step graph and metadata as declarative initial configuration.
- Remove the legacy default Flow prompt-selector script and its generated Make prompt targets;
  configured Flow step prompt paths remain the prompt-selection contract.
- Retain the new Change initializer behavior that creates or reuses a local branch and creates
  definition, Spec, and PR artifact files without publishing the branch.
- Rename the definition archive directory from `agent/ideas` to `agent/defs` while preserving its
  historical files and contents.
- Update current retained documentation and active agent prompts where they describe the renamed
  artifact, stage, or active definition paths.
- Add or update backend service tests, backend API tests, frontend tests, CLI model tests, CLI
  program integration tests, and applicable PTY tests for executable renamed behavior.

## Requirements

### Terminology contract

| Artifact slug | Full name | Verb | Editing and writing stage |
|---------------|-----------|------|---------------------------|
| `def` | `definition` | `define` | `artifact` |
| `spec` | `specification` | `specify` | `artifact` |
| `pr` | `pull-request` | `pull-request` | `artifact` |

- OTH - Every active use of `idea` that names the definition artifact must use the appropriate
  `def`, `definition`, or `define` term from the terminology contract.
- OTH - The `artifact` stage name must remain distinct from the `def`, `spec`, and `pr` artifact
  names; active stage paths, environment values, and assignments must use `artifact` when editing
  or writing any of the three artifacts.
- OTH - Ordinary uses of the English word "idea" that do not name the Change artifact must remain
  unchanged.
- OTH - Active runtime contracts must not retain `idea` fields, identifiers, routes, stages,
  commands, file names, labels, or errors for the renamed definition artifact.

### Database contract

- DB - The `public.change` definition column must be named `def`, and the change view must expose
  that column in the position consumed by backend scanners.
- DB - `public.fn_change_insert` must accept the definition as `_def`, persist it to `change.def`,
  and create the initial history row with `doc_type='def'` and the complete definition body.
- DB - The definition update procedure must be named `public.sp_change_def_update`, update
  `change.def`, preserve existing version and `agent_edit` behavior, and append a `def` history row.
- DB - Default Flow stage configuration must use `artifact` for the former `idea` stage and use
  definition wording in its help text.
- DB - Configured writable artifact types must remain `def`, `spec`, and `pr`; none of those
  artifact-type values may replace `artifact` when identifying their shared write stage.
- DB - Seed and demo setup must use the renamed database contracts without rewriting unrelated
  fixture prose that uses "idea" in its ordinary English sense.

### Backend contract

- BE - Change creation must accept a required, non-blank `def` JSON field and return the stored
  definition in the `def` response field while preserving current trimming and validation rules.
- BE - `POST /api/v1/change/update-def` must accept `id`, `def`, and `agent_edit`, persist the
  definition, and return the updated Change using the existing artifact update behavior.
- BE - The backend must no longer expose the `idea` JSON field or
  `POST /api/v1/change/update-idea` route.
- BE - Backend DTOs, API handlers, services, repositories, scanners, method names, comments, and
  errors must use `Def` or definition terminology where they represent the artifact.
- BE - Repository definition updates must call `public.sp_change_def_update`, and every Change
  scanner, including test-case projections, must scan `def` into the renamed DTO field.
- BE - Definition creation and updates must preserve current title, type, provenance, version,
  history, missing-record, and invalid-input behavior apart from the naming contract.
- BE - All backend endpoint changes must remain POST endpoints and must be covered by API tests
  using every affected request and response field.

### Frontend contract

- FE - Frontend Change and creation types must expose `def`, and API calls must send and receive
  `def` rather than `idea`.
- FE - The frontend definition update client must call
  `POST /api/v1/change/update-def` with `id`, `def`, and `agent_edit`.
- FE - Change create and edit forms must label the required field `Definition`, preserve existing
  trimming and required-field behavior, and save through the renamed API contract.
- FE - Change details must label and display the stored `Definition` value.
- FE - Frontend fixtures, mocks, component state, helper names, and tests must use definition
  terminology without changing unrelated Change form, navigation, or error behavior.

### CLI contract

- CLI - CLI DTOs, backend clients, models, state names, parsed values, detail fields, commands,
  errors, and rendered labels must use `Def` or definition terminology for the renamed artifact.
- CLI - CLI create and update requests must use the `def` JSON field and
  `POST /api/v1/change/update-def`; no active CLI client call may use the old field or route.
- CLI - New and existing definition edits must preserve current title and `Types:` parsing,
  backend save order, `agent_edit` provenance, reload behavior, and failure handling.
- CLI - Definition, Spec, and PR edit and write operations must use the same
  `.mch/tmp/<ref_uuid>/artifact` workspace and saved-session behavior previously provided by the
  shared `idea` workspace.
- CLI - CLI-launched definition, Spec, and PR processes must receive `MCH_STAGE=artifact`; active
  workspaces must not be created under an `idea`, `def`, `spec`, or `pr` stage directory for those
  operations.
- CLI - Definition write and review operations must use `def-write` and `def-review`; executable
  shared-stage chat must use `artifact-chat`, while `spec-write` and `pr-write` retain their names.
- CLI - Definition screens and previews must display `Definition`, preserve existing editing,
  cancellation, scrolling, and error behavior, and never present the artifact as `Idea`.
- CLI - Every executable CLI QA scenario in this Spec must have a named program integration test;
  affected model and PTY coverage must be updated in the same Change.

### Executable Flow, scripts, prompts, and files

- SP - The existing executable artifact operations must use this operation-to-stage mapping:

| Executable operation | Stage |
|----------------------|-------|
| `def-write` | `artifact` |
| `def-review` | `artifact` |
| `spec-write` | `artifact` |
| `pr-write` | `artifact` |
| `artifact-chat` | `artifact` |

- SP - `.mch/default/Makefile` must use `DEFS_DIR`, `def-prepare`, `def-init`, `def-write`,
  `def-review`, and `artifact-chat`; generated target lists must use the same executable names.
- SP - Definition preparation must read `agent/defs/<change-slug>.md`; definition, Spec, and PR
  writes must initialize and reuse the shared `artifact` workspace.
- SP - Default environment export must use `artifact` for the shared stage.
- SP - `.mch/default/flow.yaml` prompt declarations must remain the prompt-selection contract;
  `.mch/default/scripts/show-prompt.sh` and generated `*-prompt` Make targets must not exist.
- SP - Definition prompt files and text must use definition terminology, including input and
  output labels, review output, source authority, and resumed-session instructions.
- SP - `scripts/change-idea.pl` must be renamed to `scripts/change-def.pl`; its usage and commit
  wording must use definition terminology while preserving its add, commit, and push flow.
- SP - New Change initialization must create `agent/defs/<change-slug>.md`,
  `specs/<change-slug>.md`, and `agent/prs/<change-slug>.md` for the three artifact files.
- SP - When initialization creates a missing Change branch, it must leave the branch local without
  an upstream or remote branch after preserving the existing stage freshness and slug validation
  rules.
- SP - Initializing an existing Change branch must preserve its existing checkout and upstream
  configuration without creating another branch, pushing commits, or changing remote refs.
- SP - Remote publication must remain a separate explicit action; the existing PR publication
  workflow may push the Change branch before creating the pull request.
- SP - Active executable configuration must not reference the removed `agent/ideas` directory or
  `idea-*` definition operations.

### Initial Flow configuration

- SP - `.mch/default/flow.yaml` must retain the following configuration-only step declarations:

| Step | Mode | Declared stage |
|------|------|----------------|
| `def-create` | `editor` | `artifact` |
| `def-edit` | `editor` | `artifact` |
| `spec-edit` | `editor` | `artifact` |
| `pr-edit` | `editor` | `artifact` |
| `chat` | `chat` | `<inherit-from-previous-step>` |
| `def-write` | `exec` | `artifact` |
| `def-review` | `exec` | `artifact` |
| `branch-init` | `script` | `artifact` |
| `init-code-chat` | `chat` | `artifact` |
| `spec-write` | `exec` | `artifact` |
| `spec-review` | `exec` | `spec-review` |
| `spec-review-chat` | `chat` | `spec-review` |
| `code-implement` | `exec` | `code` |
| `code-chat` | `chat` | `code` |
| `pr-write` | `exec` | `artifact` |
| `pr-publish` | `script` | `artifact` |
| `code-review` | `exec` | `code-review` |
| `code-review-publish` | `script` | `code-review` |
| `code-review-chat` | `chat` | `code-review` |
| `pr-update` | `exec` | `artifact` |

- SP - The committed `next-step`, `commit`, `prompt`, `script`, `pre`, `post`, and
  `switch-output` declarations must remain configuration data, with terminology corrected only
  where they refer to the renamed definition artifact or stage.
- SP - Configuration help and descriptions must distinguish definition, Spec, and PR artifacts and
  must not describe Spec or PR edit steps as definition edits.
- CLI|SP - The CLI must continue loading the default Flow while unsupported keys remain inert; it
  must not execute configuration-only editor, script, routing, commit, hook, or output-switch
  declarations.
- SP - The established executable prompt remains `code-implement.md`; the configuration-only
  references do not require a `code-write.md` rename.

### Documentation and archive boundaries

- DOC - Current durable decisions and active repository prompts that describe the Change artifact
  or shared stage must use the new terminology and active paths.
- DOC - The final definition archive directory must remain `agent/defs`; `agent/prs` and `specs`
  must retain their current directory names.
- DOC - Historical files moved from `agent/ideas` to `agent/defs` must retain their names and
  contents, except where an active executable reference inside a retained file must stay valid.
- DOC - Historical Specs, PR artifacts, archived requirements, research prose, and prototype source
  must not be rewritten solely to replace past or ordinary uses of "idea".

### Test coverage

- BE - Service tests must cover definition trimming, required input, creation, updates, and
  repository delegation through the renamed DTO contract.
- BE - API tests must cover definition create, fetch, update, history and provenance, invalid
  payloads, missing records, and the absence of the old route and field.
- FE - Frontend tests must cover definition create, edit, details rendering, API payloads, and
  unchanged required-field and failure behavior.
- CLI - Model and client tests must cover renamed fields, methods, labels, operations, workspace
  paths, environment values, persistence calls, and error paths.
- CLI|SP - Program integration tests must cover executable new and existing definition, Spec, and
  PR writes in the shared `artifact` workspace, including session reuse, cancellation, and command
  failures.
- CLI - Applicable PTY tests must assert that the real terminal renders definition terminology and
  uses renamed executable operations without changing raw-mode or redraw behavior.
- CLI|SP - Configuration tests must preserve the declared initial Flow steps and prompt paths,
  confirm unsupported metadata does not become executable behavior in this Change, and confirm
  the removed prompt-selector script and generated Make targets are unavailable.
- SP - Script integration coverage must execute `scripts/change-def.pl` in a temporary Git
  repository, verify that it stages the pending definition change, creates a Definition commit, and
  pushes the Change branch, and verify that an invalid current branch is rejected without a commit
  or push.
- SP - Script integration coverage must verify new and existing Change initialization remains
  local, does not create or update remote branches, preserves existing upstream configuration, and
  creates the definition, Spec, and PR artifact files.

## Non-Goals

- Implementing a generic Flow engine for `editor`, `script`, `next-step`, inherited stages,
  commits, hooks, output switching, or `goto` routing declared in the initial Flow configuration.
- Creating the missing `spec-publish.sh`, `spec-save.sh`, `change-comment.pl`, or `pr-update.md`
  resources referenced by the configuration-only Flow graph.
- Renaming `code-implement.md` to `code-write.md`.
- Further renaming `agent/defs`, `agent/prs`, or `specs`.
- Renaming the current Change branch, Change slug, or historical filenames whose slugs contain
  `idea`.
- Rewriting historical artifact contents, archived prototypes, research prose, or ordinary English
  uses of "idea" that do not name the active definition artifact.
- Editing `AGENTS.md`.
- Adding compatibility aliases for old `idea` API fields, routes, executable operations, or paths.
- Changing artifact content rules, title or type parsing, UI navigation, or persistence behavior
  beyond the naming refactor.
- Changing the existing explicit PR publication workflow or other later workflow commands that
  intentionally push an already initialized Change branch.

## Design Notes

- `def`, `spec`, and `pr` identify three artifacts, while `artifact` identifies the shared stage
  that edits and writes all three.
- The executable operation table documents current behavior. The larger Flow table records the
  separate configuration-only graph and must not be treated as an implemented workflow.
- The user identified the expanded Flow graph as work in progress that must remain in configuration
  without adding its runtime engine or missing resources in this Change.
- The current CLI Flow decoder ignores unknown YAML fields. This Change preserves that boundary
  rather than adding models or execution logic for the new declarations.
- Existing branch work establishes the `def` database column, history type, backend DTO field,
  definition prompts, archive directory, and initial Flow graph, but each partial rename must still
  satisfy this Spec's final naming contract.
- The committed database Flow-stage value `def` conflicts with the clarified contract and must be
  `artifact`; the configured artifact type remains `def`.
- A Git push always targets a remote even when the remote name is omitted. Change initialization
  therefore performs no push; users can inspect local commits before explicitly publishing the
  branch through the later PR workflow.
- No workflow diagram is needed because executable transitions do not change, while the unexecuted
  initial graph is represented directly by its step table.
- Repository history and the PR preserve implementation progress. Final code and tests must present
  the desired naming contract without completion markers or compatibility shims.

## Verification

- `git diff --check`
- `(cd backend && make check)`
- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `(cd backend && make benchmark)`
- `(cd cli && make check)`
- `(cd cli && make test)`
- `(cd cli && make benchmark)`
- `(cd frontend && pnpm lint:check)`
- `(cd frontend && pnpm typecheck)`
- `(cd frontend && pnpm test)`
- `(cd frontend && pnpm build)`

## QA Test Cases

- BE - Create a Change with a non-blank `def`; verify the response, fetched Change, database row,
  and initial history row retain the complete definition under the renamed contract.
- BE - Update a definition with user and agent provenance; verify version, `agent_edit`, and history
  behavior remain unchanged, and invalid requests do not replace stored content.
- BE - Call the old `update-idea` route or send the old `idea` field and verify it is not accepted
  as the definition contract.
- FE - Create a Change from the web form, then open and edit it; verify every screen says
  `Definition`, the value round-trips through `def`, and existing validation and errors still work.
- CLI - Create and rewrite a definition; verify the CLI parses its title and types, uses the
  `artifact` workspace, calls the `def` backend contract, and renders `Definition` in details.
- CLI|SP - Write a definition, Spec, and PR for an existing Change; verify all three operations
  reuse the same `<ref_uuid>/artifact` workspace and session through their executable targets.
- CLI|SP - Cancel an artifact edit and force agent and backend failures; verify no unintended save
  occurs and existing error and persisted-file behavior remains intact under renamed paths.
- CLI|SP - Run executable definition review and shared chat; verify `def-review` uses the definition
  review prompt and `artifact-chat` resumes the shared artifact session.
- CLI|SP - Load the default Flow configuration; verify the expanded declarations and prompt paths
  remain available as data, unsupported fields do not trigger their hooks or routes, startup still
  succeeds, and the removed prompt-selector script and generated Make targets are unavailable.
- SP - Run `scripts/change-def.pl` in a temporary Git repository from a valid Change branch; verify
  the pending definition edit is committed with Definition wording and pushed to `origin`, then run
  it from an invalid branch and verify that it exits without committing or pushing.
- SP - Initialize a new Change in a temporary Git repository; verify its branch remains local
  without an upstream or remote ref and empty definition, Spec, and PR artifacts are created in
  `agent/defs`, `specs`, and `agent/prs`.
- SP - Initialize an existing local Change branch containing unpublished commits; verify it is
  reused without pushing or changing remote refs and its three artifact paths are present.
- DOC - Inspect current decisions, prompts, paths, and labels; verify active artifact terminology is
  updated while historical contents and excluded directory names remain unchanged.

## Review Focus

- Verify every executable layer distinguishes the `def`, `spec`, and `pr` artifacts from their
  shared `artifact` stage, especially database config, `MCH_STAGE`, workspace paths, and Make names.
- Trace the renamed `def` contract through schema, history, backend API, frontend, CLI, fixtures,
  and tests, with no stale old route or field.
- Confirm the expanded Flow graph remains intact as configuration data without introducing runtime
  semantics, missing resources, or a second executable workflow; prompt selection must use its
  configured paths rather than the removed script and Make targets.
- Confirm Change initialization never publishes new or existing local branches and leaves remote
  publication to the explicit later PR workflow.
- Check that partial backend changes call `sp_change_def_update`, scan `def` everywhere, and do not
  preserve mixed `Idea` method or error names.
- Confirm the `agent/ideas` to `agent/defs` move preserves historical content and does not expand
  into historical document rewrites or unrelated ordinary-language replacements.
- Separate the requested naming, Change initialization, and initial-config work from the unrelated
  `code-write.md` test references.

## Follow-Ups

- Implement and validate the generic Flow engine, missing scripts, missing prompts, and declared
  routing in a separate Change when the initial configuration is ready.
