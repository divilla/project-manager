# Agent Stages and Tasks Backend

## Summary
- Replaces the backend Change reference endpoint with Flow assignment through `POST /api/v1/change/assign-flow`.
- Adds backend Run claim and update operations for `POST /api/v1/change/start-run`, `POST /api/v1/change/update-run`, and `POST /api/v1/change/reset-claim`.
- Extends Change detail responses, database init/seed contract, service behavior, API tests, and docs for Flow snapshots and Run state.

## Behavior
- Flow assignment allocates missing project-scoped `ref`, refreshes `slug`, copies the current default Flow stages and stage modes onto the Change, and returns refreshed Change data.
- Repeated Flow assignment preserves the existing `ref` and copied Flow arrays, while the old `POST /api/v1/change/reference` route is no longer registered.
- Change detail responses include `flow_stages`, `flow_stage_modes`, and `run_*` fields; Change list responses continue to omit detail-only Flow and Run fields.
- Run start claims an unclaimed Change Run and returns `claim_id`; duplicate start requests return the no-claim response without replacing the active claim.
- Run update trims request values, treats stage/step/status/error as informational Worker progress fields, updates only for the active claim, returns `change_id` on success, and returns the no-update response for stale or mismatched claims.
- Completed Run updates clear `run_claim_id`, preserve completed Run state in Change detail, and set `run_is_completed`.
- Claim reset delegates stale-claim clearing to the database function and returns the new `claim_id`.

## API
- Added `ChangeUpdateRunRequest`, `ChangeRunClaimResponse`, and `ChangeRunUpdateResponse` DTOs.
- Change API route set now includes `/assign-flow`, `/start-run`, `/update-run`, and `/reset-claim`.
- Backend callers use `public.sp_change_assign_flow` instead of `public.sp_change_ref_update`.

## Data Model
- `db/init.sql` adds `public.config`, Flow snapshot columns, Run state columns, detail/list views, and the `sp_change_assign_flow`, `fn_change_start_run`, `fn_change_update_run`, and `fn_change_reset_claim` database routines.
- `db/seed.sql` seeds the default Flow stages, default stage modes, task statuses, task steps, help arrays, and placeholder Flow script/prompt arrays.
- Change history docs and database contract keep Flow snapshot and Run state fields out of `public.change_history`.

## Tests
- Service unit tests cover invalid input, request normalization, repository delegation, and error propagation for Flow assignment and Run operations.
- Change API integration tests cover Flow assignment, repeated assignment, removed `/reference`, detail/list response shapes, Run claim lifecycle, duplicate claims, stale claims, completion, reset, and invalid Run values.

## Docs
- Backend API, concepts, lifecycle, history, verification, CLI, frontend, and `mch` docs now describe Flow assignment, Run state fields, backend verification expectations, and out-of-scope frontend/CLI controls.
- Change workflow artifacts add the active Change spec and PR draft, replace the old agent phases idea with the backend Flow idea, add follow-up Flow/worker ideas, and renumber the existing idea backlog files.

## Verification
- Passed: `git diff --check origin/stage...HEAD`.
- Not run: `(cd backend && make lint)`.
- Not run: `(cd backend && make test)`.
- Not run: `(cd backend && make api-test)`.

## Risks
- The branch includes `db/init.sql` and `db/seed.sql` changes. They were inspected as branch content only; no database mutation or verification command was run while drafting this PR body.
- Older historical specs and idea files still mention `/reference`; the active Change spec and touched product docs supersede those references for this PR.

## References
- `specs/114-agent-stages-and-tasks-backend.md`
- `docs/concepts.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/history.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/mch.md`
- `docs/operations/verification.md`
