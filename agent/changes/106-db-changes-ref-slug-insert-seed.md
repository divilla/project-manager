# DB Change References, Slugs, Insert Function, and Demo Seed Data

## Goal

Add stable per-project Change references and slugs, route all Change creation through a database insert function, and refresh demo seed data so backend, frontend, and `cli` clients use the new Change identity fields consistently.

## Scope

- Add `ref` and `slug` as backend-owned Change identity fields.
- Add `last_ref` as a project-scoped sequence counter used to allocate the next Change reference.
- Add a PostgreSQL `fn_change_insert` function that is the only supported database path for inserting Change rows.
- Update backend DTOs, repositories, services, API responses, frontend state, and `cli` clients to expose and consume the new Change identity fields.
- Remove obsolete `ProjectListRequest` usage from active code.
- Update `db/seed-demo.sql` with generated demo Change rows based on scraped closed Echo pull request data.
- Add a repository-local database backup artifact for the updated demo database state.

## Requirements

- Each Change must have a `ref` value that is unique within its project and allocated from the owning project's `last_ref`.
- Each Change must have a stable `slug` derived from its project-scoped reference and title.
- Clients must not be able to supply or mutate `change.ref`, `change.slug`, or `project.last_ref` through backend APIs, frontend forms, or `cli` flows.
- Active Change inserts must call `fn_change_insert` and must not insert directly into the `change` table.
- `fn_change_insert` must return the newly created Change ID.
- Change create behavior must preserve existing validation for project ID, title, requirement body, pull request body, pull request URL, phase, types, and optional epic ID.
- Change list, get, create, and focused update responses must include the new `ref` and `slug` fields where Change DTOs are returned.
- Frontend and `cli` Change views must render or carry the new identity fields without requiring users to edit them.
- `ProjectListRequest` must be removed from active backend, frontend, and `cli` usage when the existing project list contract does not require it.
- Demo seed data must create 100 Change rows for project `demo1` from 200 scraped closed Echo pull requests.
- Demo seed data must set project `demo1` `last_ref` to `200` after importing generated Change rows.
- Demo seed data must insert generated Changes by calling `fn_change_insert`, then update `pull_request_body` from scraped PR body content.
- The repository backup artifact must represent the updated schema and demo data state without becoming an automated restore step.

## Acceptance Criteria

- Fresh database initialization creates `change.ref`, `change.slug`, `project.last_ref`, and `fn_change_insert`.
- Fresh database initialization prevents user-controlled inserts or updates of `change.ref`, `change.slug`, and `project.last_ref` through application code paths.
- Creating a Change through `POST /api/v1/change/create` returns a Change with populated `id`, `ref`, and `slug`.
- Creating multiple Changes in the same project increments `ref` from that project's `last_ref` and updates `last_ref` atomically with the insert.
- Creating Changes in different projects allocates references independently per project.
- Change list, get, create, and focused update responses include `ref` and `slug`.
- Frontend Change list/detail/create flows continue to work and do not expose editable inputs for `ref`, `slug`, or `last_ref`.
- `cli` Change list/detail/create flows continue to work and do not expose editable inputs for `ref`, `slug`, or `last_ref`.
- Active code no longer references `ProjectListRequest`.
- Running the demo seed against a disposable initialized database creates project `demo1` with `last_ref = 200`.
- The demo seed creates 100 generated Change rows for project `demo1` using source data scraped from 200 closed Echo pull requests.
- At least one seeded Change demonstrates `fn_change_insert` arguments populated from an Echo pull request and has `pull_request_body` updated after insertion.
- The database backup artifact is present in the repository and matches the updated schema and demo seed intent.

## Non-Goals

- No user-facing editor for `ref`, `slug`, or `last_ref`.
- No change to Change phase, type, epic, test case, history, or completion semantics except what is required to preserve existing behavior after the insert path changes.
- No new foreign keys.
- No production database migration or restore workflow.
- No live database scrape, seed, restore, or mutation as part of writing this specification.
- No new pull request synchronization feature beyond using scraped public PR data to generate demo seed content.
- No change to `/health`; all existing backend mutation APIs remain POST endpoints.

## Design Notes

- `docs/functionality/change-lifecycle.md` defines the existing Change create contract that must continue to be enforced after `fn_change_insert` becomes the insert path.
- `docs/architecture/backend-api.md` defines Change API endpoints and response expectations for backend clients.
- `docs/concepts.md` defines a Change as the primary delivery unit; `ref` and `slug` should support that identity without changing lifecycle semantics.
- The initial idea assumes `slug` is generated by backend/database-owned logic from the allocated project-scoped reference and title.
- The initial idea assumes the implementation can choose the backup artifact path and filename, but the artifact must be reviewable and must not require agents to restore or mutate a live database.
- Echo pull request data is external public demo content; generated seed data should be deterministic once captured in the repository.
- Compare current database and DTO behavior against `db/init.sql` and `backend/internal/dto` during implementation, but keep implementation scoped to the active Change.

## Relevant Specs

- `docs/concepts.md`
- `docs/product-overview.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/mch.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/history.md`
- `docs/operations/verification.md`
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
- From the repository root: `pnpm --dir frontend typecheck`
- From the repository root: `pnpm --dir frontend build`

## QA Test Cases

- Create a Change through the backend API and verify the response includes non-empty `ref` and `slug`.
- Create two Changes in one project and verify the second Change receives the next project-scoped `ref`.
- Create Changes in two different projects and verify their `ref` sequences are independent.
- Attempt to create or update a Change with user-supplied `ref` or `slug` and verify the supplied value is ignored or rejected according to the implemented API contract.
- Attempt to update a project with user-supplied `last_ref` and verify the supplied value is ignored or rejected according to the implemented API contract.
- List and get Changes through the backend API and verify `ref` and `slug` are present on returned Change objects.
- Use the frontend Change create/list/detail workflows and verify users can see or carry the new identity fields without editing them.
- Use the `cli` Change create/list/detail workflows and verify users can see or carry the new identity fields without editing them.
- Run the demo seed against a disposable initialized database and verify project `demo1` has `last_ref = 200`.
- Inspect seeded demo Changes and verify 100 rows were generated from captured Echo pull request data.
- Inspect a seeded Change and verify `pull_request_body` was updated after `fn_change_insert` created the Change.
- Confirm the repository backup artifact is present and corresponds to the updated schema and demo seed state.

## Review Focus

- Verify all active Change inserts go through `fn_change_insert` and preserve existing validation and history behavior.
- Check transaction handling around project `last_ref` allocation and Change insert atomicity.
- Inspect API, frontend, and `cli` DTO usage so `ref`, `slug`, and `last_ref` remain backend-owned.
- Confirm `ProjectListRequest` removal does not break existing project list clients.
- Review generated seed content for determinism, realistic PR-derived fields, and safe handling of scraped Markdown.
- Confirm the backup artifact is reviewable and not wired into an unsafe automatic restore path.

## Follow-Ups

- Decide whether Change URLs or navigation should use `slug` instead of numeric IDs in a later user-facing routing Change.
- Decide whether project-scoped Change references need search, filtering, or sorting controls after the identity fields are available.
