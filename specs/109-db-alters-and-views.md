# Database Alters And Change Views

## Goal

Align the application, documentation, tests, and demo data with the new `db/init.sql` Change persistence contract so Change records use `body`, `pr_body`, `pr_url`, `agent_edit`, `open`, and the new list/detail views consistently across the product.

## Scope

- Treat the current `db/init.sql` schema as the source of truth for Change persistence.
- Align backend DTOs, repositories, services, API handlers, API tests, and service tests with the new Change column names and response fields.
- Align frontend API types, stores, forms, detail views, rendered body behavior, and tests with the new Change field names.
- Align `cli` or `mch` code that reads or writes Change API payloads with the new Change contract.
- Update product documentation so documented Change fields, history behavior, backend API payloads, frontend behavior, and verification match the new schema.
- Update demo seed data so disposable local databases can initialize with the new Change columns and usable test case data.
- Replace the combined Change reference endpoint with separate database-backed options endpoints for change phases and change types.

## Requirements

- Active code and docs must use `body`, `pr_body`, and `pr_url` for Change requirement content, pull request body, and pull request URL.
- Active code must not read, write, serialize, or document `requirement_body`, `pull_request_body`, or `pull_request_url` as current Change fields.
- Change responses that expose linked epic display data must use backend-provided `epic_name` from the new Change read contract.
- Change list behavior must use the new list read shape and include only list-appropriate fields.
- Change detail behavior must use the new detail read shape and include body fields, timestamps, identity fields, phase/type data, linked epic data, completion counters, `agent_edit`, and `open`.
- `agent_edit` must indicate whether the current Change version was produced by an agent-assisted edit and must be preserved through Change read and history behavior.
- Change history behavior must match the new `change_history` schema and must preserve the prior active row before history-bearing updates or deletes.
- Backend focused Change update endpoints must update only their named fields, preserve `ref` and `slug`, record history where required, and return the refreshed Change using the new field names.
- Change creation must continue assigning backend-owned `ref` and `slug`; clients must not submit or edit those values.
- Change creation must use the database default open state; clients must not submit `open` on create.
- Change creation must use the database default phase unless the documented API contract is explicitly updated and tested to accept a client phase.
- Reference options for `change_phase` and `change_type` must come from separate `/api/v1/options/change-phases-list` and `/api/v1/options/change-types-list` endpoints.
- The options endpoints must return `ChangePhase` and `ChangeType` lists separately; active code must not use or expose a combined `ChangeReferences` DTO.
- Change phase and change type options must be ordered by `priority` and `slug`.
- API request and response JSON fields must match the new Change contract consistently across backend, frontend, CLI, tests, and docs.
- Demo seed data must initialize successfully against the new schema and include representative test cases so completeness can be exercised.

## Acceptance Criteria

- A fresh disposable database initialized from `db/init.sql` creates Change rows with `body`, `pr_body`, `pr_url`, `agent_edit`, and `open`.
- No active repository query, DTO, API test, frontend type, or CLI payload references the old Change body field names.
- `POST /api/v1/change/create` returns a Change with backend-assigned `ref` and `slug`, default phase, new body field names, completion counters, and timestamps.
- `POST /api/v1/change/list` returns project-scoped Change list items with `epic_name` when linked to an epic.
- `POST /api/v1/change/get` returns Change detail data with `body`, `pr_body`, `pr_url`, `agent_edit`, `open`, linked test cases, and current completion counters.
- Rendered body behavior renders markdown from `body` and `pr_body` and returns sanitized HTML using the new response field names.
- Focused update endpoints for title, body, PR body, PR URL, phase, types, epic, open state, and agent edit behave consistently with the new field names.
- `POST /api/v1/options/change-phases-list` returns the database-backed `ChangePhase` list ordered by `priority, slug`.
- `POST /api/v1/options/change-types-list` returns the database-backed `ChangeType` list ordered by `priority, slug`.
- No active backend, frontend, CLI, or test code uses a `ChangeReferences` DTO or `/api/v1/change/reference`.
- Change history rows are written before history-bearing updates and deletes using the new history schema.
- Frontend create, detail, edit, board, and rendered body flows display and submit the new Change fields without client-side field translation to old names.
- CLI or `mch` workflows that consume Change payloads compile and operate with the new field names.
- Documentation names the new Change fields and no longer describes the old fields as the active contract.
- `db/seed-demo.sql` loads against the new schema in the disposable API-test database and creates demo Changes with associated test cases.
- Backend, frontend, and CLI verification commands pass with the new contract.

## Non-Goals

- No restoration of `requirement_body`, `pull_request_body`, or `pull_request_url` as active Change fields.
- No foreign keys.
- No production migration plan beyond the repository init and seed contract.
- No authentication, authorization, hosting, or multi-user behavior changes.
- No new PR publishing integration beyond storing and displaying `pr_body` and `pr_url`.
- No change to the core rule that `ref` and `slug` are backend-owned and read-only for clients.
- No unrelated redesign of frontend layout or CLI interaction.

## Design Notes

- The current `db/init.sql` schema is the source of truth for this Change, even where existing docs still describe the previous field names.
- Documentation conflicts must be resolved by updating docs to the new schema contract.
- `change_phase` and `change_type` tables expose `slug` and `priority`; ordering should use `priority, slug`, not `priority, name`.
- Existing product rules still apply unless they directly conflict with the new schema: `ref` and `slug` remain backend-owned, Change lists remain project-scoped, and reference options remain database-provided through the options API.
- History insert and active-row mutation must remain in one transaction.
- API integration tests must exercise behavior through HTTP and must not inspect database tables directly.
- Demo seed data is only for disposable local databases.

## Relevant Specs

- `agent/changes/109-db-alters-and-views.md`
- `docs/concepts.md`
- `docs/product-overview.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/history.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/agent-interaction.md`
- `docs/operations/local-development.md`
- `docs/operations/verification.md`

## Verification

- From `backend`: `make lint`
- From `backend`: `make test`
- From `backend`: `make api-test`
- From `backend`: `GOCACHE=/tmp/project-manager-go-build go build -o /tmp/project-manager-server ./cmd/server`
- From `cli`: `make lint`
- From `cli`: `GOCACHE=/tmp/project-manager-go-build go test ./...`
- From `cli`: `GOCACHE=/tmp/project-manager-go-build go build -o /tmp/mch ./cmd/mch`
- From the repository root: `pnpm --dir frontend test`
- From the repository root: `pnpm --dir frontend typecheck`
- From the repository root: `pnpm --dir frontend build`
- From the repository root: `rg "requirement_body|pull_request_body|pull_request_url" backend frontend cli docs db --glob '!frontend/dist/**'`

## QA Test Cases

- Create a Change with a title, type, optional epic, and body; verify the response and UI show `ref`, `slug`, `body`, default phase, completion counters, and linked epic name when present.
- Open a Change detail route from a fresh browser navigation; verify the detail view loads from the backend and shows `body`, `pr_body`, `pr_url`, `agent_edit`, `open`, test cases, and completion values.
- Update Change body text; verify only the body changes, history is preserved, rendered HTML refreshes, and `ref` and `slug` remain unchanged.
- Update PR body and PR URL; verify both persist, render or display correctly, and survive page reload.
- Toggle `agent_edit`; verify the value persists and is returned by list and detail responses where the contract requires it.
- Toggle `open`; verify the value persists through `POST /api/v1/change/update-open` and is returned by list and detail responses.
- Submit `agent_edit` and `open` update payloads with the named boolean field omitted or replaced by an old field name; verify the backend returns a client-status error and preserves the existing boolean value.
- Move a Change between phases and update change types; verify database-backed reference validation rejects unknown slugs and accepts valid slugs.
- Load phase and type options separately; verify each endpoint returns only its own option type and preserves backend ordering.
- Link and unlink an epic; verify `epic_name` appears or clears in list and detail responses and epic completeness remains correct.
- Create, update, toggle, move, and delete test cases for a Change; verify Change and epic completeness update from backend responses.
- Submit invalid Change payloads with missing title, missing project, invalid type, invalid phase, invalid epic, or malformed IDs; verify client-status errors with safe JSON messages.
- Submit no-op updates with unchanged values; verify the backend returns the current Change without corrupting history or identity fields.
- Simulate backend unavailability in the frontend; verify the UI shows a clear non-blocking error state instead of stale or empty Change data.
- Load the disposable demo database; verify seeded Changes and test cases appear in the frontend and completion counters are meaningful.

## Review Focus

- Check every backend SQL query for alignment with `body`, `pr_body`, `pr_url`, `agent_edit`, `open`, `vw_change_list`, `vw_change_details`, and the new `change_history` shape.
- Verify DTO and JSON field names are consistent across backend, frontend, CLI, API tests, and documentation.
- Verify phase/type option loading uses `/api/v1/options/change-phases-list` and `/api/v1/options/change-types-list` without a combined `ChangeReferences` DTO.
- Inspect Change history paths for transaction safety and correct prior-row capture.
- Confirm create and focused update endpoints preserve backend-owned `ref` and `slug`.
- Verify old field names are not still accepted or emitted accidentally unless explicitly documented as compatibility behavior.
- Inspect seed data changes for compatibility with the new schema and for useful demo coverage without introducing non-disposable database assumptions.

## Follow-Ups

- Consider adding richer history browsing or revert UI for `agent_edit` and Change body changes.
- Consider adding migration scripts for existing non-disposable databases if the project needs upgrade support beyond fresh init and demo seed workflows.
