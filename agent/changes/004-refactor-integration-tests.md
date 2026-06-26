# Refactor Integration Tests

## Goal

Replace the current backend API integration tests with a complete, endpoint-grouped suite and document the API-test coverage rules for future backend work.

## Scope

- Delete the current backend integration tests under `backend/api-tests`.
- Rebuild `backend/api-tests` around one folder per backend API group.
- Cover every active backend endpoint with at least one API integration test.
- Update `.agent/AGENTS.md` during implementation with the API-test rules from this Change.
- Keep existing backend production behavior unchanged except where tests expose a real bug that must be fixed to satisfy current documented behavior.

## Requirements

- Integration tests must live under `backend/api-tests`.
- Each backend API group must have its own integration-test folder matching the API group name.
- Current API groups are `change`, `epic`, `health`, `project`, and `requirement`.
- Every endpoint declared by the backend API code must have at least one integration test.
- Integration tests must exercise API behavior through HTTP requests, not by calling service or repository methods directly.
- The new suite must preserve shared API-test helpers where they reduce duplication without hiding endpoint behavior.
- Documentation must consistently describe API-tests as integration tests.
- `.agent/AGENTS.md` must state that new backend endpoints and endpoint groups require immediate API-test coverage.
- `.agent/AGENTS.md` must tell reviewers to inspect endpoint additions for matching API-test coverage.

## Acceptance Criteria

- `backend/api-tests` contains only the new integration-test suite and its shared helpers.
- Each active backend API group has a matching `backend/api-tests/<group>` folder.
- Every route in `backend/internal/*/api.go` is covered by at least one API integration test.
- `.agent/AGENTS.md` includes the updated API-test rules.
- `make test` passes from `backend`.
- `make api-test` passes from `backend`.
- Documentation clearly identifies API-tests as integration tests.

## Non-Goals

- No frontend changes.
- No broad backend API redesign.
- No database schema or seed-data changes unless a current API behavior cannot be tested correctly without a targeted bug fix.
- No new unit-test policy beyond the AGENTS API-test guidance requested by this Change.

## Design Notes

- Use `docs/architecture/backend-api.md` as the endpoint contract.
- Use `docs/operations/verification.md` for backend verification commands.
- Treat `.agent/AGENTS.md` as agent operating guidance, not product documentation.
- Update `.agent/AGENTS.md` during implementation immediately before the API-test refactor, as requested by the rough Change.
- Prefer direct, readable test cases over a heavily abstracted test framework.

## Relevant Specs

- `docs/architecture/backend-api.md`
- `docs/operations/verification.md`
- `.agent/AGENTS.md`

## Verification

- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `find backend/api-tests -maxdepth 2 -type f -name '*_test.go' | sort`
- `rg -n "\\.GET\\(|\\.POST\\(" backend/internal/*/api.go`

## Review Focus

- Whether every backend endpoint has integration-test coverage.
- Whether the new test folder structure maps cleanly to backend API groups.
- Whether shared helpers simplify setup without making endpoint assertions vague.
- Whether documentation consistently describes API-tests as integration tests.
- Whether `.agent/AGENTS.md` clearly requires API tests for new endpoint work.
- Whether implementation avoids production behavior changes that are not required by failing tests.

## Follow-Ups

- None.
