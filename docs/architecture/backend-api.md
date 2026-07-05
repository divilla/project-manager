# Backend API

## Base
The backend exposes JSON endpoints under `/api/v1`. Request bodies use `application/json`.

Health checks are GET diagnostics:

- `GET /api/v1/health`
- `GET /api/health`

## Projects
Projects are managed with POST endpoints:

- `POST /api/v1/project/list`
- `POST /api/v1/project/get`
- `POST /api/v1/project/create`
- `POST /api/v1/project/update`
- `POST /api/v1/project/delete`

Project list responses include `change_count` so the UI can explain safe deletion. Clients that render interactive project list or detail screens should call the relevant project endpoint each time the user arrives at the screen instead of treating previously rendered rows as a cache.

Project list requests do not require request fields.

Project create requests send a JSON object with a `name` string. Project update requests send a JSON object with numeric `id` and `name` fields. Clients may validate project names by trimming whitespace to reject empty values, but the submitted `name` value should be sent exactly as entered when validation passes, including explicit newline characters. Project get requests identify the project by numeric `id`. Project mutation flows that need complete display data should reload the project with `POST /api/v1/project/get` after create or update.

Project reference counters are backend-owned. Project create and update requests must not accept client-supplied `last_ref`, and clients must not render `last_ref` as an editable field.

## Epics
Epics are managed with POST endpoints:

- `POST /api/v1/epic/list`
- `POST /api/v1/epic/get`
- `POST /api/v1/epic/create`
- `POST /api/v1/epic/update`
- `POST /api/v1/epic/delete`

Epic responses include aggregate completeness fields derived from linked changes. List and get responses also include `change_count` so the UI can disable deletion when an epic has linked changes.

`POST /api/v1/epic/list` requires a numeric `project_id` JSON field, for example `{"project_id": 7}`. Clients must not send `project_id` as a JSON string.

## Changes
Changes are managed with POST endpoints:

- `POST /api/v1/change/list`
- `POST /api/v1/change/get`
- `POST /api/v1/change/rendered-artifacts`
- `POST /api/v1/change/create`
- `POST /api/v1/change/assign-flow`
- `POST /api/v1/change/start-run`
- `POST /api/v1/change/update-run`
- `POST /api/v1/change/reset-claim`
- `POST /api/v1/change/update-epic`
- `POST /api/v1/change/update-phase`
- `POST /api/v1/change/update-open`
- `POST /api/v1/change/update-change-types`
- `POST /api/v1/change/update-title`
- `POST /api/v1/change/update-idea`
- `POST /api/v1/change/update-idea-agent-edit`
- `POST /api/v1/change/update-spec`
- `POST /api/v1/change/update-pr-body`
- `POST /api/v1/change/update-pr-url`
- `POST /api/v1/change/update-agent-edit`
- `POST /api/v1/change/delete`

Create payloads require only numeric `project_id`, non-blank `title`, and non-blank `idea`. Clients must not send `ref`, `slug`, `change_phase`, `change_types`, `epic_id`, `spec`, `pr_body`, `pr_url`, `agent_edit`, or `open` in the minimal create payload. Optional artifact fields default to null, `epic_id` defaults to null, `change_types` defaults to an empty array, and `change_phase` defaults to `backlog`.

Change responses include `id`, optional project-scoped `ref`, optional `slug`, aggregate fields such as `done_tc`, `total_tc`, and `completed`, timestamps where appropriate, `agent_edit`, and `open`. Change list, get, create, Flow assignment, and focused update responses return `ref` and `slug` when returning a change object, using null or the documented unassigned representation until Flow assignment.

Change list requests require a numeric `project_id` field. Clients must not send `project_id` as a JSON string. List responses include only list item fields: identity, project ID, phase and type data, linked epic identity and `epic_name` when present, title, `agent_edit`, open state, completion counters, and modified time. They return `change_types` as an array, including an empty array when no types are selected. They do not include detail-only fields such as `idea`, `spec`, rendered HTML, `pr_body`, `pr_url`, Flow snapshot fields, Run state fields, version, or created time. Clients must render the response order supplied by the backend.

Change get requests identify the Change by numeric `id`. Detail responses include `idea`, optional `spec`, optional `pr_body`, optional `pr_url`, `agent_edit`, `open`, linked epic data, `change_types`, test cases, completion counters, Flow snapshot fields, Run state fields, and timestamps. Detail-only Flow and Run JSON fields are `flow_stages`, `flow_stage_modes`, `run_claim_id`, `run_flow_stage`, `run_task_step`, `run_task_status`, `run_error`, `run_is_completed`, `run_started_at`, and `run_updated_at`. Linked test cases in Change detail responses are ordered by numeric test case ID. Clients that navigate from a Change list to detail should reload the selected Change through `POST /api/v1/change/get` instead of treating list row data as the detail source of truth.

Rendered artifact requests render markdown from `spec` and `pr_body` and return an `artifacts` array with sanitized `spec_html` and `pr_html` fields.

Flow assignment requests identify the Change by numeric `id`. `POST /api/v1/change/assign-flow` assigns missing `ref`, refreshes `slug`, copies the current global Flow stages and default stage modes onto the Change, returns refreshed Change data, and preserves existing artifacts. If the Change already has a `ref`, the endpoint must not advance the project reference counter or overwrite the copied Flow arrays; it may refresh `slug` from the current title.

The backend no longer supports `POST /api/v1/change/reference` as a public route. Clients must use `POST /api/v1/change/assign-flow` for backend identity and Flow assignment.

Run claim requests identify the Change by numeric `id`. `POST /api/v1/change/start-run` claims the Run only when `run_claim_id` is empty and returns `claim_id`; when a Run is already claimed, it returns the documented no-claim result without replacing the active claim. The first successful claim sets `run_started_at`.

Run update requests send `id`, `run_claim_id`, `run_flow_stage`, `run_task_step`, `run_task_status`, `run_error`, and `run_is_completed`. `POST /api/v1/change/update-run` mutates Run state only when the submitted Change ID and claim ID match the active claim. Successful updates return `change_id` and set `run_updated_at`; stale or mismatched claims return the documented no-update result without mutation. When `run_is_completed` is true, the update clears `run_claim_id` and leaves the completed Run visible in Change detail.

Run update validation is limited to the Change ID and active `run_claim_id` ownership. Run stage, task step, task status, and error values are informational progress fields from the Worker and are persisted as submitted after service normalization, even when they are unknown to the current global Flow configuration. Empty Run stage, task step, task status, and error values mean the Task has not started or no current value is recorded. Non-empty `run_error` is preserved as the latest Run or active Task error.

Claim reset requests identify the Change by numeric `id`. `POST /api/v1/change/reset-claim` clears stale active claim state so the Run can be claimed again and returns `claim_id` according to the backend reset contract.

Focused update endpoints identify the Change by numeric `id`, mutate only the named field, and return the refreshed Change. They must work before and after Flow assignment and must preserve the existing `ref` unless Flow assignment is called. Boolean update payloads must explicitly include the named boolean field, such as `agent_edit` or `open`; omitted fields or unsupported field names are invalid. Agent rewrite saves use `POST /api/v1/change/update-idea-agent-edit` with `id` and non-blank `idea`, atomically replacing the idea and setting `agent_edit` to true. PR URL updates accept an empty value or absolute `http` and `https` URLs only. Clients that update Change title, `idea`, `spec`, `pr_body`, `pr_url`, `agent_edit`, `open`, `change_types`, or `epic_id` should use the matching focused endpoint rather than submitting a broad edit payload.

## Options
Change options are managed with separate POST endpoints:

- `POST /api/v1/options/change-phases-list`
- `POST /api/v1/options/change-types-list`

Change phase responses return `ChangePhase` items. Change type responses return `ChangeType` items. Clients must not depend on a combined Change references response.

Change phase and type option slugs come from the database. Clients must not hardcode allowed phase or type values; they must retrieve them from `POST /api/v1/options/change-phases-list` and `POST /api/v1/options/change-types-list`. Change phase and type options are ordered by `priority` and then `slug`. Phase options may include display color metadata; clients should use it when present and fall back to neutral grey when absent. The canonical phase color metadata uses Lip Gloss color numbers: backlog `15`, staging `12`, progress `10`, rejected `9`, production `13`, and review `11`.

## Test Cases
Test cases are managed with POST endpoints:

- `POST /api/v1/test-case/list`
- `POST /api/v1/test-case/create`
- `POST /api/v1/test-case/update`
- `POST /api/v1/test-case/update-done`
- `POST /api/v1/test-case/update-change`
- `POST /api/v1/test-case/delete`

Test case payloads use `scenario` for the verifiable condition. `POST /api/v1/test-case/update-done` identifies the test case and explicitly sends the desired `done` boolean. Mutation responses include the recalculated change and current test case list when useful, so clients can refresh visible completeness and done state from backend data instead of local guesses.

## Planning
Planning endpoints are backend-mediated LLM workflows:

- `POST /api/v1/planning/decompose`
- `POST /api/v1/planning/chat`
- `POST /api/v1/planning/commit`

Generated changes and test cases must be validated against database-provided option values before being saved.

## Error Handling
The API maps domain errors to JSON responses with a `message` field. Validation errors use client status codes. Unexpected failures are logged server-side and returned as sanitized server errors.
