# Change Table `ref_uuid` and Redesigned History

## DB

Implemented:

- Add `ref_uuid` to the Change table to support future UUID-scoped temporary directories.
- Generate `ref_uuid` on insert and enforce uniqueness.
- Remove legacy `sp_change_to_history`.
- Update `fn_change_insert` to support the redesigned history model.
- Add `sp_change_idea_update`, `sp_change_spec_update`, and `sp_change_pr_update`.
- Create Changes through `fn_change_insert`, which requires `idea` and writes the initial idea history.

ToDo:

- Rebuild `seed-demo.sql` for the updated Change table and history model.

## Backend

Implemented:

- Update `backend/internal/dto/change.go` as the reference for backend naming and field shapes.

ToDo:

- Remove all backend usage of legacy `sp_change_to_history`.
- Remove transactions that only existed to support legacy `sp_change_to_history` calls.
- Use `sp_change_idea_update` from `api/v1/change/update-idea`.
- Use `sp_change_spec_update` from `api/v1/change/update-spec`.
- Use `sp_change_pr_update` from `api/v1/change/update-pr`.
- Pass `agent_edit = true` to the new history procedures when an agent made the edit to `idea`, `spec`, or `pr`; otherwise pass `false`.
- Read `agent_edit` from the request payload on `update-idea`, `update-spec`, and `update-pr`.
- Keep `agent_edit` meaningful for `change_history`. The final active-row `agent_edit` should reflect whether an agent generated the last `idea`, `spec`, or `pr` update.
- Create keeps the default `agent_edit` value for the active row.
- Remove `api/v1/change/update-idea-agent-edit` and `api/v1/change/update-agent-edit`; after create, only the three artifact update endpoints should change `agent_edit`.
- Remove `dto.ChangeUpdateIdeaAgentEditRequest` with the endpoint.
- Treat `idea`, `spec`, `pr`, and `pr_url` as strings where empty response text represents no value for not-yet-submitted optional artifacts.
- Refactor `dto.Change.Spec`, `SpecHTML`, `PR`, `PRHtml`, and `PRUrl` to non-pointer strings as shown in `dto/change.go`.
- Refactor `ChangeUpdateIdeaRequest`, `ChangeUpdateSpecRequest`, `ChangeUpdatePRRequest`, and `ChangeUpdatePRUrlRequest` to stop accepting `*string` fields.
- Stop accepting `null` or empty request values for focused `idea`, `spec`, `pr`, and `pr_url` updates.
- Return empty strings instead of `null` for `idea`, `spec`, `spec_html`, `pr`, `pr_html`, and `pr_url` JSON fields.
- Clean up unnecessary code left behind by the legacy history path.
- Update API tests and docs to remove both legacy agent-edit endpoints.

## Frontend and CLI

- After the backend changes, update frontend and CLI code for the removed agent-edit endpoints, non-null artifact fields, empty-string artifact responses, and `agent_edit` on artifact update payloads.
- Adjust backend Go/API tests, frontend tests/typecheck/build, and CLI tests for the API changes, and make sure the relevant suites pass.
