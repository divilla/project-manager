# Global Change And Test Case Refactor

## Goal

Rename the product's change completion unit from requirement to test case across the database, backend DTOs, APIs, documentation, frontend, and `cli` clients while expanding change records with requirement and pull request body fields.

## Scope

- Rename database tables, history tables, procedures, views, DTOs, response fields, and product documentation from requirement terminology to test case terminology.
- Rename `requirement.definition` to `test_case.scenario`.
- Rename `change.body` to `change.requirement_body`.
- Add `change.pull_request_body` and `change.pull_request_url`.
- Add `test_case_history.change_id`.
- Remove `vw_change` once `change` contains the same returned fields directly.
- Add focused change update endpoints for change types, title, requirement body, and pull request body.
- Update backend, frontend, `cli`, API tests, and database seed/init files so existing workflows compile and pass with the new vocabulary.

## Requirements

- The active database schema must use `test_case` and `test_case_history`; active code must not create or query `requirement` or `requirement_history`.
- `test_case.scenario` must replace `requirement.definition` in request payloads, response payloads, DTOs, services, repositories, frontend state, and tests.
- Test case mutation responses must return `test_case` and `test_cases` fields, not `requirement` or `requirements`.
- Change detail responses must expose test cases as `test_cases`.
- Change records must expose `requirement_body`, rendered requirement HTML, `pull_request_body`, and `pull_request_url`.
- `change.body` must no longer be used by active code; existing body behavior must move to `change.requirement_body`.
- `test_case_history` must store the associated `change_id` for each captured test case version.
- The `change` table must contain every field formerly supplied only by `vw_change`, and active code must use the table or repository query directly instead of relying on `vw_change`.
- Backend must expose mutation endpoints for `POST /api/v1/change/update-change-types`, `POST /api/v1/change/update-title`, `POST /api/v1/change/update-requirement-body`, and `POST /api/v1/change/update-pull-request-body`.
- The new change mutation endpoints must validate IDs and payload fields, write history where appropriate, return the refreshed change, and keep completeness values consistent.
- Legacy `/api/v1/requirement/*` routes must be replaced by `/api/v1/test-case/*` routes unless a documented compatibility shim is intentionally kept.
- Project, epic, change, test case, frontend, backend, and `cli` tests must use the new DTO and JSON field names.
- All binaries in the repository must compile after the rename.

## Acceptance Criteria

- Fresh database initialization creates `test_case`, `test_case_history`, the renamed test case recalculation procedures, and the new change columns.
- Fresh database initialization does not create active `requirement`, `requirement_history`, or `vw_change` objects.
- Creating, updating, moving, toggling done, listing, and deleting a test case use `scenario`, return `test_case` or `test_cases`, and update change and epic completeness.
- Test case history rows include `change_id` when a test case is updated or deleted.
- Epic aggregate counters use `done_tc` and `total_tc` everywhere instead of `done_req` and `total_req`.
- Creating and updating changes use `requirement_body` instead of `body`.
- Change responses include `requirement_body`, rendered requirement HTML, `pull_request_body`, and `pull_request_url`.
- `POST /api/v1/change/update-change-types`, `/update-title`, `/update-requirement-body`, and `/update-pull-request-body` update only their intended fields and return the refreshed change.
- Frontend and `cli` code no longer depend on requirement DTO names or `body` for change requirement content.
- API tests and service tests assert the new test case vocabulary and change mutation endpoint payloads.
- Backend, frontend, and `cli` test suites pass, and backend, frontend, and `cli` binaries build.
- `db/seed-demo.sql` works after the test case refactor and creates usable demo reference data, changes, test cases, and completeness counters.
- The TUI source directory is named `cli/`, and documentation explains that the current `cli` module is the interactive `mch` TUI.

## Non-Goals

- No user authentication, authorization, or account behavior changes.
- No change phase or change type taxonomy changes beyond endpoint support for updating existing fields.
- No project or epic schema changes except aggregate calculations affected by test case completeness.
- No new pull request publishing workflow; this Change stores pull request body and URL fields only.
- No data migration for an existing production database beyond the repository's init/seed contract.

## Design Notes

- `docs/concepts.md` defines test cases as the renamed binary Definition of Done unit for changes.
- `docs/functionality/requirements-and-acceptance.md` remains the documentation home for completion rules, but its content must use test case terminology.
- `docs/architecture/backend-api.md` defines the new test case endpoints and change payload fields.
- `docs/functionality/history.md` defines test case history expectations.
- Use `test_case` for database objects and JSON field names, and `TestCase` for Go type names.
- Keep compatibility shims only if they are explicitly documented and covered by tests; otherwise remove legacy requirement routes and fields.

## Relevant Specs

- `docs/concepts.md`
- `docs/product-overview.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/mch.md`
- `docs/architecture/system-architecture.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/history.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/current-project-context.md`
- `docs/docs-rules.md`

## Verification

- From `backend`: `make lint`
- From `backend`: `make test`
- From `backend`: `make api-test`
- From `backend`: `GOCACHE=/tmp/project-manager-go-build go test ./api-tests/... -run '^$'`
- From `backend`: `GOCACHE=/tmp/project-manager-go-build go build -o /tmp/project-manager-server ./cmd/server`
- From `cli`: `make lint`
- From `cli`: `GOCACHE=/tmp/project-manager-go-build go test ./...`
- From `cli`: `GOCACHE=/tmp/project-manager-go-build go build -o /tmp/mch ./cmd/mch`
- From the repository root: `pnpm --dir frontend test`
- From the repository root: `pnpm --dir frontend build`

## Review Focus

- Verify no active backend or frontend code still reads or writes `requirement`, `requirement_history`, `definition`, or `change.body`.
- Check database procedure renames and completeness recalculation paths for both change and epic counters.
- Inspect API response compatibility and JSON names for test case mutation responses.
- Confirm the four new change update endpoints validate input, write history, and refresh returned change data.
- Check frontend and `cli` compile-time DTO usage after the rename.

## Follow-Ups

- Decide whether docs file paths and package names that still include `requirement` should be renamed in a later mechanical cleanup.
- PR comment `IC_kwDOTA2Xls8AAAABIGsGEQ` was reviewed and intentionally not applied: the requested foreign key restoration conflicts with the root `AGENTS.md` database hard boundary, which explicitly says not to create foreign keys.
