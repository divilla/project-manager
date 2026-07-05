# Agent Stages and Tasks Backend

Types: feature|docs|test

## Goal

Backend Change APIs expose Flow assignment and run-state operations so automation runners can claim, update, reset, and inspect a Change run using the Flow configuration copied onto that Change.

## Scope

- Update backend Change API, service, repository, DTO, and API-test coverage for Flow assignment and run-state operations.
- Replace the public Change reference endpoint contract with `POST /api/v1/change/assign-flow`.
- Add backend endpoints for `POST /api/v1/change/start-run`, `POST /api/v1/change/update-run`, and `POST /api/v1/change/reset-claim`.
- Return Flow snapshot and run-state fields on Change detail responses according to the existing `internal/dto` JSON field names.
- Align documentation that describes Change identity, Flow assignment, run state, backend Change endpoints, and verification expectations.
- Update repository agent instructions for backend layer boundaries and Go import formatting.
- Use the already-updated database contract in `db/init.sql` and `db/seed.sql` as the persistence basis for this Change.

## Requirements

- `POST /api/v1/change/assign-flow` must replace `POST /api/v1/change/reference` as the supported backend endpoint for assigning Change identity and Flow configuration.
- Assigning a Flow to an unreferenced Change must allocate the next project-scoped `ref`, refresh `slug` from the current title, copy the current global `flow_stages` and `flow_stage_modes_default` from `public.config` onto the Change, and return the refreshed Change.
- Assigning a Flow to a Change that already has `ref` must not allocate a new reference number. It may refresh `slug` from the current title and must preserve the Change's copied `flow_stages` and `flow_stage_modes`.
- The previous database procedure name `public.sp_change_ref_update` must no longer be used by backend code; assignment must go through `public.sp_change_assign_flow`.
- Change detail responses must include `flow_stages`, `flow_stage_modes`, `run_claim_id`, `run_flow_stage`, `run_task_step`, `run_task_status`, `run_error`, `run_is_completed`, `run_started_at`, and `run_updated_at`.
- Change list responses must remain list-appropriate and must not grow detail-only Flow or run-state fields unless documentation is explicitly updated to say otherwise.
- `POST /api/v1/change/start-run` must accept a Change `id`, claim the run when `run_claim_id` is empty, and return `claim_id`.
- Starting a run that is already claimed must not replace the active claim and must return the documented no-claim result instead of silently claiming the run.
- `POST /api/v1/change/update-run` must accept `id`, `run_claim_id`, `run_flow_stage`, `run_task_step`, `run_task_status`, `run_error`, and `run_is_completed`, then update only when the Change ID and active claim ID match.
- `POST /api/v1/change/update-run` must return `change_id` for a successful update and the documented no-update result when `_id`, `_claim_id`, or both do not match an active claim.
- Updating a run must treat stage, step, status, and error values as informational runner progress fields; the active `run_claim_id` is the concurrency guard for run updates.
- A successful run update must set `run_updated_at`; the first claim must set `run_started_at`.
- When `run_is_completed` is true, the update must mark the run completed and clear the active claim so the completed run is no longer considered claimed.
- `POST /api/v1/change/reset-claim` must accept a Change `id`, clear a stale active claim so the run can be claimed again, and return `claim_id` for the new claim state defined by `public.fn_change_reset_claim`.
- Direct backend writes to `flow_*` and `run_*` columns must be avoided outside the defined procedure and functions.
- `run_error` must be empty when no current error is recorded. A non-empty `run_error` must be preserved as the latest run or active-task error.
- Empty run stage, task step, task status, and error values mean the task has not started yet or no current value is recorded.
- Run state values must be updated after task execution finishes, not before a task has produced an outcome.
- `public.change_history` must not store the Flow snapshot or run-state columns.
- The global Flow configuration must be represented through `public.config` with ordered arrays for Flow stages, stage help, default modes, stage mode help, task statuses, task status help, task steps, and task step help.
- The default Flow stages must be ordered as `idea`, `spec`, `ready`, `docs`, `code`, `polish`, `pr`, `review`, `fix`, `sync`, `merge`, `stage`, `master`.
- Stage modes must be `skip`, `prompt`, and `exec`.
- Task statuses must be `queued`, `running`, `paused`, `stopped`, `waiting`, `completed`, and `failed`.
- Task steps must be `none`, `entry`, `prompt`, `agent`, `exit`, and `done`.
- Documentation must use the Flow naming convention: Flow defines Steps; a Run executes a Flow; a Task performs one Step within a Run; a Worker executes the Task.
- Documentation must state that frontend and CLI support for these Flow controls is out of scope for this backend Change.
- `AGENTS.md` must define API handlers as Echo/HTTP-only boundary code and repositories as database-driver-only boundary code, with validation, normalization, business rules, orchestration, and endpoint-specific behavior owned by the service layer unless those concerns directly require transport or database-driver objects.
- `AGENTS.md` must name `goimports` instead of `gofmt` for Go code quality formatting.

## Acceptance Criteria

- `POST /api/v1/change/assign-flow` assigns `ref`, `slug`, `flow_stages`, and `flow_stage_modes` for an unreferenced Change and returns a refreshed Change response.
- Repeating `POST /api/v1/change/assign-flow` for the same Change does not advance the project reference counter and does not overwrite the Change's copied Flow arrays.
- `POST /api/v1/change/reference` is removed from the backend public route set and backend callers/tests no longer depend on it.
- `POST /api/v1/change/start-run` returns a non-empty `claim_id` for an unclaimed Change run.
- Repeating `POST /api/v1/change/start-run` while the Change is already claimed returns the documented no-claim response and leaves the existing claim intact.
- `POST /api/v1/change/update-run` with a matching `id` and `run_claim_id` returns `change_id` and persists the submitted run stage, task step, task status, error, completion flag, and update timestamp.
- `POST /api/v1/change/update-run` with a missing, stale, or mismatched claim does not mutate run state and returns the documented no-update response.
- A completed run update clears `run_claim_id`, sets `run_is_completed` true, and leaves the completed run visible through Change detail.
- `POST /api/v1/change/reset-claim` clears stale claim state according to the database function and returns `claim_id`.
- Change detail API responses include the Flow snapshot and run-state JSON fields using the `internal/dto` field names.
- API integration tests cover the new and renamed endpoints, happy paths, duplicate claim handling, stale claim handling, completed-run handling, invalid IDs and claim IDs, informational run values, and response field shapes.
- Service-layer unit tests cover request validation, run-state normalization, repository call behavior, and error propagation for the new run operations.
- Docs under `docs/` describe the new endpoint names, Flow configuration, run-state fields, out-of-scope frontend/CLI work, and backend verification commands.
- Repository instructions document strict API handler and repository layer boundaries and require `goimports` for Go formatting.

## Non-Goals

- Frontend screens or controls for Flow configuration, per-Change stage modes, runner state, or claim reset.
- CLI support for assigning Flow, adjusting per-Change stage modes, starting runs, updating runs, or resetting claims.
- User-editable reordering, adding, or deleting of Flow stages.
- Creating foreign keys or adding project-wide locking, advisory locks, isolation-level escalation, or coordinated locking protocols.
- Persisting Flow snapshot or run-state columns to `public.change_history`.

## Design Notes

- Existing documentation currently describes the older `POST /api/v1/change/reference` contract; this Change intentionally supersedes that with `POST /api/v1/change/assign-flow`.
- `internal/dto` is the naming reference for JSON fields. If field types conflict with the database contract, implementation should resolve them toward the product contract: `run_is_completed` is a boolean run-completion flag.
- `public.config` stores global Flow configuration. Its arrays are positional; help text and default modes must match `flow_stages` by array index.
- Existing Changes keep their copied Flow configuration. Later global Flow configuration updates apply only to Changes assigned after those updates.
- Keep run mutation logic simple and transactional. Use conventional `Begin`, deferred `Rollback`, and `Commit` around multi-step mutations.
- Keep backend layer-boundary decisions consistent with `AGENTS.md`: handlers isolate Echo and HTTP objects, repositories isolate database-driver objects, and service code owns application behavior.
- Database files have already been changed for this Change. Agents remain bound by repository database safety rules and must not edit `db/**` unless the user explicitly authorizes the exact database-file change.
- The intended plain-language model is: Flow = reusable automation definition, Step = one named stage inside the flow, Run = one execution attempt of a flow, Task = one unit of work inside a run for a specific step, Worker = executor/tool/process that performs a task.

## Relevant Specs

- `specs/114-agent-stages-and-tasks-backend.md`
- `docs/concepts.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/agent-interaction.md`
- `docs/architecture/backend-api.md`
- `docs/operations/verification.md`

## Verification

- `(cd backend && make lint)`
- `(cd backend && make test)`
- `(cd backend && make api-test)`

## QA Test Cases

- Create a Change without a reference, call `POST /api/v1/change/assign-flow`, and verify the response includes `ref`, `slug`, `flow_stages`, and `flow_stage_modes`.
- Call `POST /api/v1/change/assign-flow` twice for the same Change and verify the second call does not allocate a new `ref` or replace the copied Flow arrays.
- Call the removed `POST /api/v1/change/reference` route and verify it is no longer available.
- Start a run for an unclaimed Change and verify `claim_id`, `run_claim_id`, and `run_started_at` are set.
- Start a run for an already claimed Change and verify the active claim is not replaced.
- Update a run with the active claim ID and submitted stage, task step, and task status; verify the response `change_id` and refreshed detail run fields.
- Update a run with a stale or mismatched claim ID and verify no run fields are changed.
- Complete a run and verify `run_is_completed` is true, `run_claim_id` is cleared, and detail still shows the last stage, step, status, error, and update timestamp.
- Reset a stale claim and verify the returned `claim_id` can be used to continue or reclaim the run as documented.
- Submit invalid IDs and blank or malformed claim IDs and verify validation errors are returned without mutation.
- Submit unknown or runner-local Flow stages, task steps, and task statuses with an active claim and verify they are persisted as informational run state.
- Verify an empty `run_error` represents success/no current error and a non-empty `run_error` is returned as the latest error.
- Verify Change history rows do not include Flow snapshot or run-state columns.

## Review Focus

- Public API route rename from `/reference` to `/assign-flow` and any remaining backend callers or tests that still use the old route.
- Correct claim lifecycle behavior for start, update, completion, stale-claim mismatch, and reset paths.
- Transaction boundaries around assignment and run updates, especially preserving existing refs and copied Flow arrays on repeated assignment.
- Claim lifecycle behavior without rejecting informational stage, task step, task status, or error values from older or runner-local flows.
- DTO/database type alignment for `run_is_completed`, timestamp fields, and nullable `run_claim_id`.
- API integration tests under `backend/api-tests/change` proving the observable endpoint contracts.
- Documentation consistency for Flow, Step, Run, Task, and Worker naming.

## Follow-Ups

- Add frontend Flow and run-state controls.
- Add CLI runner support for per-Change stage modes, claiming, progress updates, and reset behavior.
