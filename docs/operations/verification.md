# Verification

## Backend
From `backend`:

```sh
make test
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics.

`make api-test` runs API integration tests from `backend/api-tests`. These tests exercise backend endpoints over HTTP and should be organized by API endpoint group.

`make api-test` recreates the `changes_test` database from `db/init.sql` and `db/seed.sql`, then starts a temporary backend. It verifies fresh-schema behavior only.

Use `AIPM_API_BASE_URL=http://localhost:8080 make api-test-existing` from `backend` to run the same API tests against an already-running local backend without recreating schema or restarting the server. This mode mutates data through the API and must not target production. Use it when repository SQL, database views, or endpoint response contracts changed, or when an existing backend reports a database-contract error.

## Frontend
From the repository root:

```sh
pnpm --dir frontend test
pnpm --dir frontend typecheck
pnpm --dir frontend build
```

Frontend checks should cover feature logic, visible component behavior, routing, and project-scoped refresh behavior.

## Documentation
Documentation checks should list files, enforce the 300-line limit, and run the vocabulary checks from the active Change file.

Personal research under `docs/research` is not product documentation and is excluded from rewrite verification.
