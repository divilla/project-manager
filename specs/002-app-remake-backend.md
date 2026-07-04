# Application Remake Backend

## Goal

Refactor the backend from the old task-based model to the current change-based product model, using `db/init.sql`, `db/seed.sql`, and `backend/internal/dto` as the implementation contract.

## Scope

- Rename backend domain code, routes, DTO usage, services, repositories, tests, and API test packages from task terminology to change terminology.
- Align all backend SQL with the refactored database schema: `project`, `epic`, `change`, `requirement`, reference tables, history tables, views, and stored procedures.
- Preserve the refactored database direction instead of restoring old task tables or hierarchical task behavior.
- Replace `name` and `description` behavior for changes with `title` and `body`.
- Replace hierarchical parent behavior with the fixed change structure: standalone changes or changes linked to one epic by `epic_id`.
- Add backend support for epics and epic history where required by the schema and product docs.
- Update requirement behavior so requirements attach to changes through `change_id` and return recalculated change data.
- Update backend unit tests and API integration tests to use change naming, change endpoints, and the new schema.

## Requirements

- Backend code must compile without references to removed task DTOs or removed task schema objects.
- Backend route groups must expose change endpoints under `/api/v1/change` and requirement reassignment under `/api/v1/requirement/update-change`.
- Project responses must use `change_count` from `vw_project`.
- Change create, list, get, rendered body, update, phase, epic, closed, and delete behavior must use the fields defined by `backend/internal/dto` and the database schema.
- Change history must write `change_history` before history-bearing update or delete behavior.
- Epic mutations must write `epic_history` before history-bearing update or delete behavior.
- Requirement create, list, update, done toggle, reassignment, and delete behavior must use `change_id` and must recalculate the affected change and epic completeness in one transaction.
- Backend code and backend tests must not use the old task hierarchy, `parent_id`, `task_phase`, `task_type`, `task_id`, `name` as a change title field, or `description` as a change body field.
- Backend code must not depend on `vw_change`; `public.change` now contains the generated `completed` field that the old task model previously sourced from a view.
- Add `-port` and `-db` flags to the backend server binary so tests can start it with temporary runtime overrides.
- Existing frontend code is not part of this Change.

## Acceptance Criteria

- `go test ./...` passes from `backend`.
- `make test` passes from `backend`.
- `make api-test` passes from `backend` when the local test database is available.
- The server starts with `go run ./cmd/server` against the refactored database.
- `POST /api/v1/project/list` returns projects with `change_count`.
- Change endpoints operate against `public.change`, `public.change_phase`, `public.change_type`, and `public.sp_change_*` objects without requiring `public.vw_change`.
- Requirement endpoints operate against `public.requirement.change_id` and return mutation responses containing `change`.
- No backend package path, package name, route, request field, response field, test package, or API test fixture uses old task terminology.
- Deleting a project with changes is rejected without deleting active changes or requirements.
- Deleting a change archives or removes linked requirements according to history rules and removes the active change.
- Moving a change between standalone and epic-linked states updates affected completeness counters.

## Non-Goals

- No frontend refactor; that is tracked by `agent/changes/003-app-remake-frontend.md`.
- No restoration of old task tables, task views, task procedures, or hierarchical task behavior.
- No new planning or LLM behavior.
- No authentication, authorization, multi-user behavior, or deployment work.
- No broad redesign beyond making the backend match the current database and DTO contract.

## Design Notes

- Treat `db/init.sql`, `db/seed.sql`, and `backend/internal/dto` as the source of truth for backend implementation details.
- Use the product docs for vocabulary and behavior boundaries; do not keep compatibility aliases for old task API routes.
- Change body markdown rendering should keep the existing backend sanitizer/parser approach while using change naming.
- Keep mutating resource endpoints as POST endpoints.
- Keep history insert and active-row mutation in the same transaction.
- Where the current DTOs still contain leftover old names, normalize them to the change vocabulary before implementing dependent code.

## Relevant Specs

- `docs/concepts.md`
- `docs/architecture/system-architecture.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/history.md`
- `docs/operations/local-development.md`
- `docs/operations/verification.md`
- `backend/internal/dto`
- `db/init.sql`
- `db/seed.sql`

## Verification

- `go test ./...` from `backend`
- `make test` from `backend`
- `make api-test` from `backend`
- `go run ./cmd/server -db postgres://postgres:postgres@localhost:5432/changes_test -port 19080` from `backend`
- `rg "\b[Tt]ask\b|task_" backend`
- `rg "/api/v1/task|task_id|task_count|task_phase|task_type|parent_id|description_html" backend`
- `rg "public\.task|vw_task|sp_task_|fn_task_descendants|task_history" backend db --glob '!db/backup/**'`

## Review Focus

- Whether the backend truly follows the change schema instead of partially adapting old task code.
- Whether history and completeness recalculation are transactionally correct for changes, epics, and requirements.
- Whether removed fields and hierarchy behavior are gone from API contracts and tests.
- Whether API route changes match the documented backend API.
- Whether test coverage proves project, epic, change, requirement, and history behavior against the refactored schema.

## Follow-Ups

- Refactor the frontend to the new change endpoints and payloads in `agent/changes/003-app-remake-frontend.md`.
