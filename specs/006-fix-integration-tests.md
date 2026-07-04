# Fix Integration Tests

## Goal

Make backend API integration tests exercise the backend only through HTTP requests and responses, with the test database lifecycle owned by the test runner instead of the tests.

## Scope

- Refactor `backend/api-tests` so tests do not open direct database connections.
- Remove direct SQL/history assertions from API integration tests or replace them with behavior visible through API responses.
- Keep `make api-test` responsible for rebuilding `changes_test` from repository-root `db` scripts before every run.
- Make `make api-test` start the backend server with explicit `-port` and `-db` flags.
- Make it obvious in `backend/Makefile` that API tests run only against `changes_test`.
- Remove API-test environment-variable configuration that can point tests at mismatched API and database targets.
- Update verification documentation to describe the HTTP-only integration-test boundary.

## Requirements

- API integration tests must interact with the system under test only through backend HTTP endpoints.
- API integration tests must not import `pgx`, create connection pools, run SQL, or inspect database tables directly.
- `make api-test` must rebuild `changes_test` from existing files under `db/` before starting tests.
- `make api-test` must start the backend executable with `-port` and `-db`.
- `make api-test` must never target the `changes` database.
- Test helpers must not read environment variables to choose API URLs or database URLs.
- Any behavior that cannot be verified through existing API responses must either be removed from API-test assertions or exposed through an intentional API contract in a separate Change.

## Acceptance Criteria

- No files under `backend/api-tests` contain `os.Getenv`.
- No files under `backend/api-tests` import `pgx` or `pgxpool`.
- No files under `backend/api-tests` execute SQL queries.
- `backend/api-tests/shared` contains no direct database helper.
- `(cd backend && make api-test)` rebuilds `changes_test` from `../db/init.sql` and `../db/seed.sql`.
- `(cd backend && make api-test)` starts `aipm-server` with `-db postgres://postgres:postgres@localhost:5432/changes_test` and `-port 18080`.
- API tests connect only to the fixed local API-test backend URL started by `make api-test`.
- The API-test suite still covers every endpoint group required by `.agent/AGENTS.md`.

## Non-Goals

- No new database schema, seed, backup, or restore changes.
- No new history-read API unless a separate Change explicitly defines that product behavior.
- No changes to production or local development database contents outside the disposable `changes_test` verification flow.
- No frontend behavior changes.

## Design Notes

- Treat `backend/api-tests` as black-box API tests: request in, response out.
- Keep direct database verification out of API integration tests. If lower-level database behavior needs coverage, use another explicitly named test layer in a separate Change.
- `changes_test` is disposable test state. Rebuilding it inside `make api-test` is expected and required.
- The backend server already supports `-port` and `-db`; test runners should use those flags instead of relying on environment overrides.

## Relevant Specs

- `.agent/AGENTS.md`
- `docs/operations/verification.md`
- `docs/operations/local-development.md`
- `docs/architecture/backend-api.md`

## Verification

- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `rg -n "os\\.Getenv|pgx|pgxpool|QueryRow|Query\\(" backend/api-tests`
- `rg -n "changes_test|changes\\b|AIPM_|DATABASE_URL|PORT" backend/Makefile backend/api-tests docs/operations`

## Review Focus

- Whether API tests are now truly HTTP-only.
- Whether `make api-test` is visibly and safely pinned to `changes_test`.
- Whether direct database history assertions were removed without weakening endpoint contract coverage.
- Whether any remaining configuration path can accidentally run API tests against the `changes` database.

## Follow-Ups

- Consider a separate Change for a deliberate history-read API if history behavior must be verified through public contracts.
- Fixed PR comment `IC_kwDOTA2Xls8AAAABHvm6Dg`: pinned `make api-test` variables against command-line overrides before any database rebuild command runs.
