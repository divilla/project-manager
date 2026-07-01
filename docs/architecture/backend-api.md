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

- `POST /api/v1/change/reference`
- `POST /api/v1/change/list`
- `POST /api/v1/change/get`
- `POST /api/v1/change/rendered-bodies`
- `POST /api/v1/change/create`
- `POST /api/v1/change/update-epic`
- `POST /api/v1/change/update-phase`
- `POST /api/v1/change/update-closed`
- `POST /api/v1/change/update-change-types`
- `POST /api/v1/change/update-title`
- `POST /api/v1/change/update-requirement-body`
- `POST /api/v1/change/update-pull-request-body`
- `POST /api/v1/change/update-pull-request-url`
- `POST /api/v1/change/delete`

Create payloads use `project_id`, `title`, `requirement_body`, `change_types`, and optional `epic_id`. Clients must not send `ref`, `slug`, `change_phase`, `pull_request_body`, or `pull_request_url` in create payloads.

Change responses include `id`, project-scoped `ref`, stable `slug`, aggregate fields such as `done_tc`, `total_tc`, and `completed`, timestamps, and rendered requirement HTML for display. Change list, get, create, and focused update responses all return `ref` and `slug` when returning a change object.

Change list requests require a numeric `project_id` field. Clients must not send `project_id` as a JSON string. List responses are ordered by `modified` descending.

Change get requests identify the Change by numeric `id`. Clients that navigate from a Change list to detail should reload the selected Change through `POST /api/v1/change/get` instead of treating list row data as the detail source of truth.

Focused update endpoints identify the Change by numeric `id`, mutate only the named field, and return the refreshed Change. They must preserve the existing `ref` and `slug`. Clients that update Change title, `requirement_body`, `change_types`, or `epic_id` should use the matching focused endpoint rather than submitting a broad edit payload.

## Test Cases
Test cases are managed with POST endpoints:

- `POST /api/v1/test-case/list`
- `POST /api/v1/test-case/create`
- `POST /api/v1/test-case/update`
- `POST /api/v1/test-case/update-done`
- `POST /api/v1/test-case/update-change`
- `POST /api/v1/test-case/delete`

Test case payloads use `scenario` for the verifiable condition. Mutation responses include the recalculated change and current test case list when useful.

## Planning
Planning endpoints are backend-mediated LLM workflows:

- `POST /api/v1/planning/decompose`
- `POST /api/v1/planning/chat`
- `POST /api/v1/planning/commit`

Generated changes and test cases must be validated against database-provided reference options before being saved.

## Error Handling
The API maps domain errors to JSON responses with a `message` field. Validation errors use client status codes. Unexpected failures are logged server-side and returned as sanitized server errors.
