# Change Lifecycle

## Overview
A change is the delivery unit. It can exist independently or reference one epic. It is never part of a nested change tree.

Each change has a backend-generated `ref_uuid` and may have a backend-owned `ref` and `slug`. The `ref_uuid` is a non-null read-only UUID identity. The `ref` is a project-scoped numeric reference allocated from the owning project. The `slug` is a branch identifier generated from the assigned reference and the Change title at first Flow assignment.

Users and clients must not create, edit, or overwrite `ref`, `ref_uuid`, `slug`, Flow snapshot fields, Run state fields, or the project's reference counter. New Changes may exist without a `ref`, `slug`, or Flow snapshot; those fields are assigned by backend Flow assignment and returned to clients as read-only data.

## Create
Creating a change requires:

- project ID
- title
- idea

The title and idea must be non-blank after validation. `spec`, `pr`, and `pr_url` are optional artifact strings and default to empty text when omitted. `epic_id` defaults to null, `change_types` defaults to an empty array when omitted or explicitly empty, and `change_phase` defaults to `backlog` when clients do not provide a phase.

Only the AI-assisted `IdeaCreate` path initializes the Change title from an artifact H1: it sends the canonical Idea H1 as the explicit create `title`. Later Idea, Spec, and PR titles and metadata are independent from the Change and from one another.

After a successful create, the returned change includes its database ID, generated `ref_uuid`, string artifact fields, and may have unassigned `ref` and `slug` values. Creating a Change does not advance the project reference sequence.

Codex-assisted planning requires canonical Idea bytes with a body containing at least one non-whitespace character before `Create Change?`, then creates the Change and runs the composed Rewrite and Review Steps. Title-only, whitespace-body, and metadata-only artifacts are invalid in this workflow. Agent-produced canonical changes save through `POST /api/v1/change/update-idea` with `agent_edit: true`; user Editor saves use `agent_edit: false`. These Changes use the backend default phase until the user moves them through the normal lifecycle.

## Flow Assignment
`POST /api/v1/change/assign-flow` assigns or returns identity and Flow configuration for an existing Change identified by numeric `id`.

For an unreferenced Change, the endpoint allocates the next project-scoped `ref`, derives `slug` from that reference and the current Change title, copies the current global Flow stages and default stage modes onto the Change, preserves existing artifacts, and returns refreshed Change data. For a referenced Change, the endpoint returns the existing `ref`, `slug`, and copied Flow arrays without regenerating identity.

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

The Change title, type set, and linked Epic change only through their focused update flows; those operations preserve an assigned slug. Every submitted Idea, Spec, or PR requires a body containing at least one non-whitespace character, then validates and canonicalizes its own optional metadata before persistence. Types use `Types: <type-slugs>` with valid slugs joined by `|` and no spaces. Epic uses `Epic: <epic-title> #<epic-id>` with the canonical current-project Epic title and ID. Artifact titles and metadata may differ from both the Change and one another without updating Change fields.

Artifact update payloads for `idea`, `spec`, and `pr` include `agent_edit` to record whether that artifact save was agent-produced. After create, `agent_edit` is changed only by those artifact update flows.

Focused updates return the refreshed change with its existing `ref` and `slug`. Updates must work before and after Flow assignment. Updating the title changes the user-facing Change title but does not regenerate an assigned slug or let clients supply replacement identity values.

## Delete
Deleting a change is destructive and must be confirmed. Test cases linked to the change are archived or removed according to backend history rules before the active change is removed.

## Epic Link
A change may reference one epic. Updating this reference updates aggregate completeness for the old and new epic as needed.
