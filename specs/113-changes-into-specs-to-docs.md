# Changes Into Specs To Docs

Types: docs|refactor|test

## Goal

The repository Change workflow uses `change/<change-slug>` branches and stores implementation contracts as Change specs under repo-root `specs/`, while preserving `Change` as the core product and workflow concept.

## Scope

- Rename the repository branch namespace used by Change workflow automation from `changes/<change-slug>` to `change/<change-slug>`.
- Move Change contract Markdown artifacts from `agent/changes/` to repo-root `specs/`.
- Update active workflow prompts, scripts, docs, CLI branch reconciliation code, and tests that refer to Change spec paths or Change workflow branch names.
- Update `AGENTS.md` repository agent guidance for the new Change spec path, `change/<change-slug>` branch namespace, and artifact terminology.
- Update visible and internal wording from `Change file` to `Change spec` only when the text refers to the repository Markdown contract artifact.
- Preserve product terminology where `Change` means the product object, workflow unit, API resource, route, package, or domain concept.

## Requirements

- Change workflow branches must use `change/<change-slug>` everywhere the workflow creates, validates, checks out, renames, pushes, or documents a branch for a Change.
- `changes/<change-slug>` must no longer be accepted as the canonical workflow branch name by active agent prompts, workflow scripts, or `mch` `/reference` branch reconciliation.
- Repository Change specs must live at `specs/<change-slug>.md`.
- Active workflow code must read, write, stage, commit, inspect, and reference Change specs from repo-root `specs/` instead of `agent/changes/`.
- The old `agent/changes/` Change spec files must be moved to matching `specs/` paths without changing their contract content except for path or terminology updates required by this Change.
- Active documentation must describe the repository artifact as a `Change spec`, not a `Change file`.
- `AGENTS.md` must describe Change specs under repo-root `specs/`, `change/<change-slug>` workflow branches, and `Change spec` terminology for the repository Markdown contract artifact.
- Active documentation must continue to use `Change` for the product and workflow object; the implementation must not rename the product concept to `Spec`.
- Existing frontend application routes such as `/changes`, package names such as `changes`, API paths such as `/api/v1/change/*`, and database/domain names for Changes must remain unchanged unless they specifically refer to Git workflow branch names or repository spec files.
- `mch` `/reference` must reconcile Git branches using `change/<slug>` and `change/<ref>-*` for local and remote branch detection, checkout, rename, push, and delete behavior.
- Workflow scripts must fail fast with clear errors when run outside a `change/<change-slug>` branch.
- Any helper that migrates existing local or remote branches must operate from `changes/` to `change/`, avoid deleting a remote branch before the replacement branch exists, and be covered by dry-run or testable behavior when feasible.
- Tests and documentation checks must cover the renamed branch namespace, new Change spec path, and corrected terminology.

## Acceptance Criteria

- All repository Change specs that previously lived under `agent/changes/` are tracked under repo-root `specs/` with matching file names.
- Active workflow prompts and scripts use `specs/<change-slug>.md` as the Change spec path.
- Active workflow prompts and scripts validate the current branch as `change/<change-slug>` and reject `changes/<change-slug>` with a clear error.
- `mch` `/reference` checks for, checks out, renames, creates, pushes, and deletes `change/<slug>` or `change/<ref>-*` branches instead of `changes/<slug>` or `changes/<ref>-*`.
- Product docs that describe PR integration, agent interaction, local verification, and `mch` branch reconciliation refer to `change/<change-slug>` branches and `specs/<change-slug>.md` specs.
- Active docs and prompts refer to the Markdown contract artifact as a `Change spec`.
- `AGENTS.md` refers to `specs/<change-slug>.md`, `change/<change-slug>` branches, and Change specs for active repository workflow guidance.
- Product, API, frontend route, package, and database terminology still refer to the domain object as a `Change`.
- Search results for `agent/changes/`, `Change file`, and workflow-branch `changes/<...>` references are either removed from active workflow surfaces or intentionally limited to historical/research context that is not active product behavior.
- Existing frontend `/changes` routes and backend `/api/v1/change/*` endpoints continue to work and are not renamed as part of this Change.
- Automated tests fail before the branch/path rename behavior is implemented and pass after implementation.

## Non-Goals

- Renaming the product concept `Change`.
- Renaming backend API routes, frontend application routes, Go package names, database tables, DTO names, or domain fields that represent Changes.
- Changing Change lifecycle phases, type slugs, epic behavior, test case behavior, or project selection behavior.
- Changing the standard Change spec section structure beyond terminology and path references.
- Running database migrations, changing files under `db/**`, or mutating any database state.

## Design Notes

- Documentation is the source of truth; active docs should describe current behavior, not compatibility with the old `changes/` branch namespace.
- The idea explicitly distinguishes the product object from the repository artifact: `Change` remains the cornerstone of the app and process; only `Change file` becomes `Change spec`.
- The current `mch` architecture docs still describe `/reference` branch reconciliation with `changes/<slug>`; this Change updates that contract to `change/<slug>`.
- `docs/docs-rules.md` requires product vocabulary to keep `change`, `epic`, `test case`, and `history` as core terms and to use `title`, `idea`, `spec`, `pr`, `pr_url`, and `agent_edit` for Change artifacts.
- `docs/functionality/pr-integration.md` is an active behavior doc and must move from `changes/<change-slug>` plus `agent/changes/<change-slug>.md` to `change/<change-slug>` plus `specs/<change-slug>.md`.
- Historical PR notes and research may mention old paths when they are clearly historical, but active workflow prompts, scripts, and product behavior docs must not rely on them.
- Valid type slugs are sourced from repository seed data; this spec uses `docs|refactor|test`.

## Relevant Specs

- `specs/113-changes-into-specs-to-docs.md`
- `agent/ideas/113-changes-into-specs-to-docs.md`
- `.mch/default/prompts/spec-file-structure.md`
- `docs/docs-rules.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/pr-integration.md`
- `docs/operations/verification.md`
- `specs/112-ideas-and-change-refactor.md`

## Verification

- `git diff --name-status origin/stage...HEAD`
- `rg -n "agent/changes|Change file|changes/<|changes/\\$|changes/" agent docs scripts cli backend frontend`
- `(cd cli && make lint)`
- `(cd cli && go test ./...)`
- `(cd cli && go build -o /tmp/mch ./cmd/mch)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`

## QA Test Cases

- Run a workflow script from a `change/113-changes-into-specs-to-docs` branch and verify it derives `113-changes-into-specs-to-docs` and reads the Change spec from `specs/113-changes-into-specs-to-docs.md`.
- Run the same workflow script from a `changes/113-changes-into-specs-to-docs` branch and verify it fails before making changes with a clear branch-name error.
- Generate or update a Change spec through the active planning workflow and verify only `specs/<change-slug>.md` is written and committed.
- Draft a PR body through the active workflow and verify the PR references the Change spec under `specs/`.
- Run `mch` `/reference` when local `change/<slug>` already exists and verify it checks out that branch.
- Run `mch` `/reference` when only local `change/<ref>-*` exists and verify it checks out that branch, renames it to `change/<slug>`, and refreshes Change details.
- Run `mch` `/reference` when only remote `change/<slug>` exists and verify it fetches/checks out the remote branch.
- Run `mch` `/reference` when only remote `change/<ref>-*` exists and verify it creates `change/<slug>` remotely before deleting the old remote branch.
- Run `mch` `/reference` when no matching local or remote branch exists and verify it creates local `change/<slug>`.
- Simulate Git checkout, rename, push, and delete failures during `/reference`; verify each failure remains recoverable and does not display stale success.
- Search active docs, prompts, scripts, CLI code, backend code, and frontend code for old `agent/changes/`, `Change file`, and workflow-branch `changes/` references; verify remaining matches are either intentionally historical or unrelated product routes/packages.
- Open the frontend `/changes` routes and verify they still work after the workflow branch rename because application routing is out of scope.
- Call representative backend `/api/v1/change/*` endpoints or run backend tests and verify API paths remain unchanged.

## Review Focus

- Branch namespace changes in workflow scripts and `mch` Git reconciliation, especially local/remote rename and delete ordering.
- Correct isolation between workflow branch names (`change/<slug>`) and product/frontend route names (`/changes`).
- Complete movement of Change spec artifacts from `agent/changes/` to repo-root `specs/` without losing existing spec content.
- Active docs and prompts using `Change spec` for the artifact while preserving `Change` as the domain object.
- Tests proving old branch names are rejected and new branch names are accepted.
- Verification search results for stale `agent/changes/`, `Change file`, and workflow `changes/` references.

## Follow-Ups

- Consider renaming `agent/prompts/change-file-*` prompt file names to `change-spec-*` if a later cleanup wants filenames to match the new terminology.
