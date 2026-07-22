# Change Lifecycle

## Overview
A change is the delivery unit. It can exist independently or reference one epic. It is never part of a nested change tree.

Each change has a resolved `ref_uuid` and may have a backend-owned `ref` and `slug`. A create request may supply `ref_uuid`; when it is omitted or null, the backend generates one. The resolved identity is non-null and read-only after creation. The `ref` is a project-scoped numeric reference allocated from the owning project. The `slug` is a branch identifier generated from the assigned reference and current title.

Clients may set `ref_uuid` only as an optional Change create field. Later requests cannot replace or clear it. Users and clients must not create, edit, or overwrite `ref`, `slug`, Flow snapshot fields, Run state fields, or the project's reference counter. New Changes may exist without a `ref`, `slug`, or Flow snapshot; those fields are assigned or refreshed only by backend Flow assignment and returned to clients as read-only data.

## Create
Creating a change requires:

- project ID
- title
- idea

The title and idea must be non-blank after validation. `ref_uuid` is optional: the backend preserves a supplied UUID or generates a UUIDv7 when the field is omitted or null. Blank or malformed non-null values fail request binding without creating a Change or initial history. `spec`, `pr`, and `pr_url` are optional artifact strings and default to empty text when omitted. `epic_id` defaults to null, `change_types` defaults to an empty array when omitted or explicitly empty, and `change_phase` defaults to `backlog` when clients do not provide a phase.

After a successful create, the returned change includes its database ID, resolved `ref_uuid`, string artifact fields, and may have unassigned `ref` and `slug` values. Creating a Change does not advance the project reference sequence.

Codex-assisted planning tools create Changes after the user confirms `Create Change?`, then may run an agent rewrite and save the rewritten idea through `POST /api/v1/change/update-idea` with `agent_edit` set to true. These Changes use the backend default phase until the user moves them through the normal lifecycle.

## Flow Assignment
`POST /api/v1/change/assign-flow` assigns or returns identity and Flow configuration for an existing Change identified by numeric `id`.

For an unreferenced Change, the endpoint allocates the next project-scoped `ref`, refreshes `slug` from the current title, copies the current global Flow stages and default stage modes onto the Change, preserves existing artifacts, and returns refreshed Change data. For a referenced Change, the endpoint must not allocate a new `ref` or overwrite the copied Flow arrays; it may refresh `slug` from the current title.

Flow configuration copied onto a Change is stable for that Change. Later global Flow configuration changes apply only to Changes assigned after those global changes.

## List
Project-scoped lists show active changes for one project. Lists include list-appropriate fields only: identity, phase and type data, linked epic identity and `epic_name` when present, title, `agent_edit`, open state, completion counters, and modified time.

List items include `ref` and `slug` when assigned and must represent unassigned identity without deriving it locally. Clients must use the backend response order.

Flow snapshot and Run state are detail-only fields. List responses must not include `flow_stages`, `flow_stage_modes`, or `run_*` fields unless the list contract is explicitly changed.

## Detail
The detail view shows:

- project-scoped reference and slug
- title, idea, and spec
- `pr` and `pr_url` when present
- phase and type information
- linked epic and `epic_name` when present
- test case list ordered by test case ID
- completion counters
- `agent_edit`
- open state
- Flow snapshot fields `flow_stages` and `flow_stage_modes`
- Run state fields `run_claim_id`, `run_flow_stage`, `run_task_step`, `run_task_status`, `run_error`, `run_is_completed`, `run_started_at`, and `run_updated_at`
- version, created time, and modified time

Markdown `spec` and `pr` rendering is sanitized by the backend before display.

## Run State
`POST /api/v1/change/start-run` claims an unclaimed Change Run and returns `claim_id`. If a Run is already claimed, the endpoint returns the documented no-claim result and leaves the active claim unchanged.

`POST /api/v1/change/update-run` updates Run state only when the submitted Change `id` and `run_claim_id` match the active claim. Successful updates return `change_id`, set `run_updated_at`, and persist the submitted Flow Step, Task Step, Task status, error, and completion flag. Missing, stale, or mismatched claims return the documented no-update result without mutation.

Run updates use the active `run_claim_id` as the concurrency guard. Flow Step, Task Step, Task status, and error values are informational progress fields from the Worker and may reflect an older copied Flow or runner-local state. Empty stage, task step, task status, and error values mean the Task has not started yet or no current value is recorded. A non-empty `run_error` is the latest Run or active Task error.

When `run_is_completed` is true, the update marks the Run completed, clears the active claim, and leaves the completed Run state visible in Change detail. Run state values are updated after a Task execution finishes, not before a Task has produced an outcome.

`POST /api/v1/change/reset-claim` clears stale active claim state so the Run can be claimed again and returns `claim_id` according to the backend claim-reset contract.

Frontend and CLI controls for Flow assignment, per-Change stage modes, Run claiming, Run updates, and claim reset are out of scope until dedicated frontend or CLI Changes add them.

## Update
Editing a change can update title, `idea`, `spec`, `pr`, `pr_url`, type classification, epic reference, phase, and open state. Artifact responses use strings; empty text is the no-value state for optional artifacts that have not been submitted and for rendered artifact HTML. Focused artifact and PR URL update payloads must be non-empty. Open-state updates use an explicit boolean value and return refreshed backend state.

Artifact update payloads for `idea`, `spec`, and `pr` include `agent_edit` to record whether that artifact save was agent-produced. After create, `agent_edit` is changed only by those artifact update flows.

The CLI edits existing Idea, Spec, and PR artifacts in `.mch/tmp/<ref_uuid>/idea`. It reloads the Change before editing, refreshes `output.md` and then `input.md` from the backend artifact, and preserves the stage session and logs. Changed user output is first saved with `agent_edit=false`, then copied from `output.md` to `input.md`; only after both steps succeed does the matching `idea-write`, `spec-write`, or `pr-write` prompt run with `MCH_STAGE=idea` and the shared session. Agent output is saved with `agent_edit=true`. Unchanged output sends no artifact update. Both accepted saves apply present type metadata and reload the Change before continuing.

Idea, Spec, and PR artifacts share one optional `Types:` metadata contract. Omitting `Types:` leaves the Change's existing types untouched and sends no type update. When `Types:` is present, it always sends a request to `POST /api/v1/change/update-change-types`: an empty value sends an empty string array, while a non-empty value treats pipes and whitespace as separators, then removes every `[^A-Za-z\-_]` character from each value and discards empty values. Clients do not load the type options catalog first. The backend intersects every submitted array with its currently available type values, silently removes unsupported values, and replaces the Change's complete type set with the filtered result, including an empty array. Unsupported artifact type values are not validation errors. Existing-Change artifact saves reload the Change through `POST /api/v1/change/get`; initial creation uses the complete Change returned by its post-create type update.

Focused updates return the refreshed change with its existing `ref` and `slug`. Updates must work before and after Flow assignment. Updating the title does not let clients supply replacement identity values.

## Delete
Deleting a change is destructive and must be confirmed. Test cases linked to the change are archived or removed according to backend history rules before the active change is removed.

## Epic Link
A change may reference one epic. Updating this reference updates aggregate completeness for the old and new epic as needed.
