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

Project list responses include `change_count` so the UI can explain safe deletion.

## Epics
Epics are managed with POST endpoints:

- `POST /api/v1/epic/list`
- `POST /api/v1/epic/get`
- `POST /api/v1/epic/create`
- `POST /api/v1/epic/update`
- `POST /api/v1/epic/delete`

Epic responses include aggregate completeness fields derived from linked changes. List and get responses also include `change_count` so the UI can disable deletion when an epic has linked changes.

## Changes
Changes are managed with POST endpoints:

- `POST /api/v1/change/reference`
- `POST /api/v1/change/list`
- `POST /api/v1/change/get`
- `POST /api/v1/change/rendered-bodies`
- `POST /api/v1/change/create`
- `POST /api/v1/change/update`
- `POST /api/v1/change/update-epic`
- `POST /api/v1/change/update-phase`
- `POST /api/v1/change/update-closed`
- `POST /api/v1/change/delete`

Create and update payloads use `title` and `body`.

## Requirements
Requirements are managed with POST endpoints:

- `POST /api/v1/requirement/list`
- `POST /api/v1/requirement/create`
- `POST /api/v1/requirement/update`
- `POST /api/v1/requirement/update-done`
- `POST /api/v1/requirement/update-change`
- `POST /api/v1/requirement/delete`

Requirement mutation responses include the recalculated change and current requirement list when useful.

## Planning
Planning endpoints are backend-mediated LLM workflows:

- `POST /api/v1/planning/decompose`
- `POST /api/v1/planning/chat`
- `POST /api/v1/planning/commit`

Generated changes and requirements must be validated against database-provided reference options before being saved.

## Error Handling
The API maps domain errors to JSON responses with a `message` field. Validation errors use client status codes. Unexpected failures are logged server-side and returned as sanitized server errors.
