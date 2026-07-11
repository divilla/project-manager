# Ideas And Change Refactor

Types: feature|refactor|docs|test|migration

## Goal

Change creation and reference assignment are refactored so a Change starts from a required title and idea, can be rewritten by an agent through explicit API flow, then receives optional spec, PR, type, epic, ref, and slug artifacts through explicit backend and `mch` flows.

## Scope

- Update the Change artifact model around shared `title`, required `idea`, optional `spec`, optional `pr`, and optional `pr_url`.
- Add `change.idea`, empty `change_types` defaults, optional ref/slug state, and the new reference-assignment persistence path to canonical initialization data.
- Add a backend Change reference endpoint that assigns missing `ref` and `slug` through `sp_change_ref_update` without depending on update flows.
- Use the edited `internal/dto/change.go` structs as the reference contract, then update backend services, repositories, API handlers, validation, and history behavior to match those DTOs.
- Update frontend Change create, detail, edit, and display behavior to support idea, spec, empty type arrays, optional PR fields, and unassigned ref/slug state.
- Update `mch` `/new-change`, `CreateIdeaState`, `UpdateIdeaState`, `RewriteIdeaState`, `ChangeDetailsScreen`, and `/reference` behavior for the new idea-first and reference flows.
- Preserve existing agent execution code paths needed for later spec-writing work.
- Update documentation and automated tests for backend, frontend, and CLI behavior affected by the artifact and reference refactor.

## Requirements

- A Change must require only `project_id`, `title`, and `idea` at creation time.
- `idea` must store the user idea at creation time, store the rewritten idea after a successful agent rewrite save, and be returned by Change detail APIs and clients.
- `spec`, replacing the former Change body requirement role, must be optional and default to null until generated or edited.
- `pr` and `pr_url` must be optional and default to null until supplied.
- `change_phase` must remain mandatory and must default to `backlog` when clients do not provide a phase.
- `change_types` must remain part of the Change contract, but create and update flows must allow no selected types; omitted types must default to an empty array at the API, client, and storage boundaries.
- `epic_id` must be optional and default to null when no epic is selected.
- `ref` and `slug` must be optional/unassigned on newly created Changes until the reference flow assigns them.
- `ref` and `slug` must remain backend-owned identity fields; clients must not create, edit, or overwrite their values directly.
- `fn_change_insert` must accept only `project_id`, `title`, and `idea`; it must not assign `ref` or `slug`.
- Reference assignment must move out of `sp_change_insert` and into a separate `sp_change_ref_update` procedure.
- `sp_change_ref_update` must be independent of Change update flow, so Change updates work both before and after the user runs `/reference`.
- The backend must expose `POST /api/v1/change/reference` to assign or return a Change reference for an existing Change and then return refreshed Change data.
- Calling `POST /api/v1/change/reference` for a Change that already has a `ref` must not allocate a new `ref`; it may refresh `slug` from the current title for the explicit `/reference` branch-reconciliation flow.
- Backend create and update paths must reject missing or blank required title or idea values and must accept omitted optional artifact fields.
- Backend list and detail responses must expose empty `change_types` arrays and optional/unassigned ref and slug without stale field names from the former body-only Change contract.
- Frontend and CLI clients must use backend APIs for all Change persistence and reference assignment and must not write directly to the database.
- `mch` `/new-change` must start from `ChangesListState`, require a valid numeric current project, open an editor for the initial idea, and then prompt `Create Change?` with Yes and No choices after the editor exits.
- `CreateIdeaState` and `UpdateIdeaState` must parse `# <title>` after editor exit before any create or update API call; parse failures must show the idea markdown first, then `error parsing title:` with `/edit` and `/cancel`.
- In `CreateIdeaState` and `UpdateIdeaState`, `/edit` must reopen the idea editor with the current content, and `/cancel` must remove `/tmp/mch/initial-idea.md` and return to `ChangesListScreen`.
- In `CreateIdeaState` and `UpdateIdeaState`, prompt rendering must read `/tmp/mch/initial-idea.md` and show the idea markdown first, then one blank line, then the question or error and prompt options, including the initial `Resume idea?` prompt.
- Idea prompt previews must keep raw Markdown syntax visible and apply nano-style foreground syntax coloring without rendering Markdown into blocks, frames, or filled backgrounds.
- Selecting No from `Create Change?` must return to `ChangesListScreen` without calling the backend create endpoint.
- Selecting Yes from `Create Change?` must create the Change through `POST /api/v1/change/create`, then start a new agent session in `RewriteIdeaState` to rewrite the idea.
- `UpdateIdeaState` must save the edited idea through `POST /api/v1/change/update-idea` before entering `RewriteIdeaState`.
- Every successful rewrite save must call `POST /api/v1/change/update-idea-agent-edit`, remove `/tmp/mch/initial-idea.md`, reload Change data from the backend API, and route to `ChangeDetailsScreen`.
- `ChangeDetailsScreen` must add `/reference` as the first command before existing detail commands and expose spec editing as `/edit-spec`.
- `mch` `/reference` must call `POST /api/v1/change/reference`, refresh `ChangeDetailsScreen` from the backend, and then reconcile the Git branch for the Change.
- Before Git branch reconciliation, `mch` must verify that the app is running inside a Git repository and show a recoverable error if it is not.
- Git branch reconciliation must first check whether local `changes/<slug>` exists with `git branch --list`; if it exists, checkout that branch.
- If local `changes/<slug>` does not exist, Git branch reconciliation must check for a local `changes/<ref>-*` branch; if found, checkout it and rename it to `changes/<slug>` with `git branch -m changes/<slug>`.
- If no local branch matches, Git branch reconciliation must check whether remote `changes/<slug>` exists; if found, checkout the remote branch.
- If remote `changes/<slug>` does not exist, Git branch reconciliation must check for remote `changes/<ref>-*`; if found, checkout it and rename it both locally and on the remote to `changes/<slug>`.
- If no local or remote branch matches, Git branch reconciliation must create and checkout `changes/<slug>` with `git checkout -b changes/<slug>`.
- All backend, frontend, and CLI behavior changed by this artifact and reference model must have automated test coverage appropriate to the affected layer.

## Acceptance Criteria

- Creating a Change through the backend succeeds with only a valid `project_id`, `title`, and `idea`, and the response shows default `backlog` phase with null optional fields where omitted.
- Creating a Change with missing or blank `title` or `idea` returns a validation error and does not persist a Change.
- Creating a Change with omitted `change_types`, `epic_id`, `spec`, `pr`, or `pr_url` persists an empty `change_types` array and null optional artifacts, then returns them consistently in detail responses.
- New Changes can exist before reference assignment with unassigned ref and slug values, and list/detail clients render that state without deriving identity locally.
- `POST /api/v1/change/reference` assigns ref and slug to an unreferenced Change, returns refreshed Change data, and preserves all existing artifact fields.
- Repeating `POST /api/v1/change/reference` for the same Change does not advance the project reference counter; if the title changed, the refreshed slug is returned and branch reconciliation renames or checks out the matching branch.
- Updating title, idea, spec, PR fields, phase, types, epic, agent edit, open state, and test cases works before and after reference assignment.
- Backend DTOs and JSON contracts use the new artifact field names from `internal/dto/change.go` and no create path requires the former body/spec value.
- Existing Change list, get, focused update, rendered body/spec, option, and test case workflows continue to work with Changes whose types are empty arrays and whose artifacts, ref, or slug are null.
- Frontend Change create and detail flows support title plus idea creation and render optional spec, PR artifacts, empty type arrays, and unassigned ref/slug state without crashes or placeholder-only success states.
- `mch` `/new-change` opens idea entry from `ChangesListState`, handles No from `Create Change?` as a no-op return, and handles Yes by creating the Change before running the idea rewrite agent path.
- `CreateIdeaState` and `UpdateIdeaState` parse invalid title input before any API write, show the idea markdown before `error parsing title:`, expose `/edit` and `/cancel`, and remove `/tmp/mch/initial-idea.md` on cancel.
- After a rewritten idea save, `mch` reloads the persisted Change from the backend and opens `ChangeDetailsScreen`.
- `ChangeDetailsScreen` command ordering shows `/reference` first and exposes `/edit-spec` for spec edits.
- `mch` `/reference` calls the backend reference endpoint, refreshes the Change details, and performs the documented local/remote branch checkout, rename, or create path for `changes/<slug>`.
- `mch` `/reference` shows recoverable errors for backend failures, non-Git directories, Git command failures, branch conflicts, and remote rename failures without displaying stale local success state.
- The existing agent execution code for future spec writing remains present and covered enough that a later Change can re-enable the Yes path without reconstructing it.
- Repository docs describe the new artifact model, empty `change_types` behavior, backend API payloads, reference endpoint, Git branch reconciliation, and CLI idea rewrite routing.
- Backend, frontend, and CLI tests cover successful creation, validation failures, null optional fields, reference assignment, ref-preserving reference calls, Git branch paths, backend failures, cancellation/no-op paths, and persisted reload behavior.

## Non-Goals

- Completing the future `Write Spec with Agent?` Yes path that generates and saves a spec.
- Renaming repository `agent/changes/` files or moving Change specifications to another folder.
- Changing PR publication behavior beyond keeping `pr` and `pr_url` optional artifacts.
- Allowing clients to submit arbitrary `ref` or `slug` values.
- Adding foreign keys or introducing new project-wide locking protocols.
- Reworking unrelated Change lifecycle phases, option ordering, epics, test case completeness, or project selection behavior.

## Design Notes

- The idea says the agent should "rewrite the skill"; this Change treats that as rewriting the user's idea because the surrounding flow saves and displays a rewritten idea.
- The idea says types remain mandatory while defaulting to an empty array; this Change treats an empty `change_types` array as the valid no-selection state across API, client, and storage layers.
- Documentation currently defines `body` and required non-empty `change_types`; this Change intentionally changes that contract to `idea`, optional `spec`, and empty `change_types` array support, and must update docs in the same implementation.
- Documentation currently says `ref` and `slug` are assigned by Change insert persistence; this Change moves assignment to `sp_change_ref_update` while preserving backend ownership and read-only client behavior.
- `spec` is the new product term for the former Change body/requirement artifact, but the implementation may need compatibility mapping where existing API, storage, or rendering fields are renamed in phases.
- `POST /api/v1/change/reference` should identify the Change by numeric `id`; `ref` and `slug` are outputs of the endpoint, not request inputs.
- `/reference` is an explicit user command, so refreshing an existing slug from the current title is allowed when the command also reconciles the Git branch to that refreshed slug.
- Branch reconciliation should treat `origin` as the default remote unless repository configuration clearly supplies a different tracked remote.
- Local branch checks should use exact `changes/<slug>` matching before prefix `changes/<ref>-*` matching to avoid renaming the canonical branch.
- Remote `changes/<ref>-*` rename requires creating/pushing `changes/<slug>` and removing the old remote ref only after the new remote ref succeeds.
- The backend DTO layer, especially `internal/dto/change.go`, is the reference contract for this implementation; DTO-only edits are not complete until backend, frontend, and CLI behavior align with them.
- Agents may read `db/**` for context, but implementation must only edit database files when explicitly authorized by the active Change scope and must not run database mutation commands without explicit user approval.
- Database initialization changes must not introduce foreign keys.
- `mch` must keep using backend APIs for persistence and must not write application database tables directly.
- The future `Write Spec with Agent?` branch is out of this implementation; reviewers should verify the agent execution code for later spec writing is preserved rather than deleted.

## Relevant Specs

- `agent/changes/112-ideas-and-change-refactor.md`
- `docs/concepts.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/history.md`
- `docs/operations/local-development.md`
- `docs/operations/verification.md`

## Verification

- `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make lint)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make api-test)`
- `(cd backend && GOCACHE=/tmp/project-manager-go-build make race)`
- `(cd cli && make lint)`
- `(cd cli && go test ./...)`
- `(cd cli && go build -o /tmp/mch ./cmd/mch)`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`

## QA Test Cases

- Create a Change through the backend with only project ID, title, and idea; verify the detail response includes the saved idea, default `backlog` phase, null optional fields, and unassigned ref and slug.
- Try to create a Change with blank title and with blank idea; verify validation errors are returned and no Change is created.
- Create Changes with omitted types, empty selected types, and selected types; verify create responses, list views, and detail views use empty arrays for no selected types.
- Create a Change without an epic, spec, PR body, or PR URL; verify all omitted optional artifacts remain null after reload.
- Call `POST /api/v1/change/reference` for an unreferenced Change; verify ref and slug are assigned, existing artifacts are preserved, and the refreshed detail response contains the new identity.
- Call `POST /api/v1/change/reference` twice for the same Change without changing the title; verify the second call returns the same ref and slug and does not advance the project reference counter.
- Change a referenced Change title and run `/reference`; verify the existing ref is preserved, slug refreshes from the title, and branch reconciliation renames or checks out the expected branch.
- Update title, idea, spec, phase, types, epic, PR body, PR URL, agent edit, open state, and test cases before reference assignment; verify each update persists and survives later reference assignment.
- Update title, idea, spec, phase, types, epic, PR body, PR URL, agent edit, open state, and test cases after reference assignment; verify ref and slug remain unchanged.
- Use the frontend to create a title plus idea Change; verify detail display, reload, and optional artifact rendering work when types, spec, ref, and slug are empty.
- Start `mch` `/new-change` with no valid current project; verify the flow stops before editor, agent, or backend create work.
- Start `mch` `/new-change`, enter an idea without a `# <title>`, and verify the rendered prompt shows the idea markdown before `error parsing title:` with `/edit` and `/cancel`.
- Edit an existing Change idea without a `# <title>` and verify no `update-idea` call is made, the rendered prompt shows the idea markdown before `error parsing title:`, `/edit` reopens the idea editor, and `/cancel` removes `/tmp/mch/initial-idea.md` and returns to `ChangesListScreen`.
- Start `mch` `/new-change`, enter an idea, choose No at `Create Change?`, and verify the app returns to `ChangesListScreen` without creating a Change.
- Start `mch` `/new-change`, enter an idea, choose Yes at `Create Change?`, and verify create runs before the idea rewrite agent path and the rewritten idea is saved through `update-idea-agent-edit`.
- Edit an existing Change idea, exit the editor, and verify `mch` calls `update-idea`, enters `RewriteIdeaState`, then saves the rewritten idea through `update-idea-agent-edit`.
- After a rewritten idea save, verify `ChangeDetailsScreen` reloads the Change from the backend and displays the rewritten idea.
- Run `mch` `/reference` for a Change when local `changes/<slug>` already exists; verify the app checks out that branch and refreshes details.
- Run `mch` `/reference` for a Change when only local `changes/<ref>-*` exists; verify the app checks it out, renames it to `changes/<slug>`, and refreshes details.
- Run `mch` `/reference` for a Change when only remote `changes/<slug>` exists; verify the app checks out the remote branch and refreshes details.
- Run `mch` `/reference` for a Change when only remote `changes/<ref>-*` exists; verify the app checks it out, renames it locally and remotely to `changes/<slug>`, and refreshes details.
- Run `mch` `/reference` for a Change with no matching local or remote branch; verify the app creates and checks out `changes/<slug>`.
- Run `mch` `/reference` outside a Git repository; verify a recoverable error is shown and no branch mutation is attempted.
- Simulate backend failure while saving the rewritten idea; verify `mch` shows a recoverable error and does not display a local-only success state.
- Simulate backend failure during `/reference`; verify no Git branch reconciliation is attempted after the failed API call.
- Simulate Git checkout, rename, push, and delete failures during `/reference`; verify errors are recoverable and Change details remain loaded from backend data.
- Simulate agent failure during idea rewrite; verify no backend save is attempted and the user sees recoverable command output.
- Verify existing Changes with legacy body/spec data still load or migrate according to the implemented compatibility path.

## Review Focus

- Backend DTO and API contract changes for required `idea`, optional `spec`, empty `change_types` arrays, optional PR artifacts, and optional/unassigned ref and slug.
- Persistence changes that add `change.idea`, default `change_types` to an empty array, move identity assignment to `sp_change_ref_update`, preserve existing data, and avoid unauthorized database behavior.
- Reference allocation behavior of `POST /api/v1/change/reference` and `sp_change_ref_update`, especially preserving existing refs while allowing explicit slug refresh and branch reconciliation.
- Compatibility between old body-oriented Change behavior and the new spec artifact.
- `mch` state transitions around `Create Change?`, `CreateIdeaState`, `UpdateIdeaState`, `RewriteIdeaState`, and `/reference`.
- CLI idea prompt rendering, especially markdown preview order, parse-error options, and temporary idea cleanup on cancel.
- Git branch reconciliation correctness for exact slug branches, ref-prefix branches, local branches, remote branches, and failure paths.
- Frontend and CLI handling of null optional fields, empty type arrays, and unassigned ref/slug values without crashes or stale field names.
- Automated tests that prove cancellation, validation, backend failure, agent failure, Git failure, ref preservation, slug refresh, and reload behavior.

## Follow-Ups

- Implement the real `Write Spec with Agent?` Yes path that generates and saves a spec.
- Consider a separate migration or compatibility Change for renaming repository `agent/changes` terminology if product naming continues moving from Change body to spec artifacts.
