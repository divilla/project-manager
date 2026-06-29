# Verification

## Backend
From `backend`:

```sh
make lint
make test
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics.

After every backend code change, agents must run `make lint` from `backend` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the backend code change.

`make api-test` runs API integration tests from `backend/api-tests`. These tests exercise backend endpoints over HTTP and should be organized by API endpoint group.

`make api-test` recreates the disposable `changes_test` database from `db/init.sql` and `db/seed.sql`, then starts a temporary backend on port `18080` with `-db postgres://postgres:postgres@localhost:5432/changes_test` and `-port 18080`.

API integration tests must interact with the backend only through HTTP requests and responses. They must not choose targets from environment variables, open database connections, run SQL, or inspect tables directly.

## `mch`
From `make-a-change`:

```sh
make lint
go test ./...
go build -o ./bin/mch ./cmd/mch
```

After every `mch` code change, agents must run `make lint` from `make-a-change` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the `mch` code change.

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
