# Verification

## Backend
From `backend`:

```sh
make lint
make test
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics. Change API coverage should include title plus idea creation, missing or blank title and idea validation, null optional artifacts, empty `change_types`, unassigned `ref` and `slug`, Flow assignment, repeated Flow assignment that preserves the existing `ref` and copied Flow arrays, removal of the old `/api/v1/change/reference` route, Run claiming, duplicate claim handling, stale claim rejection, completed Run handling, invalid Run IDs and claim IDs, informational Run values, and Change detail response field shapes.

After every backend code change, agents must run `make lint` from `backend` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the backend code change.

`make api-test` runs API integration tests from `backend/api-tests`. These tests exercise backend endpoints over HTTP and should be organized by API endpoint group.

`make api-test` recreates the disposable `changes_test` database from `db/init.sql` and `db/seed.sql`, then starts a temporary backend on port `19080` with `-db postgres://postgres:postgres@localhost:5432/changes_test` and `-port 19080`. If that port is already in use, run `API_TEST_PORT=19081 make api-test` to use a different local port.

API integration tests must interact with the backend only through HTTP requests and responses. They may read the runner-provided HTTP base URL, but must not open database connections, run SQL, or inspect tables directly.

## `mch`
From `cli`:

```sh
make lint
go test ./...
go build -o /tmp/mch ./cmd/mch
```

Verification commands may write build outputs and fixtures under the system temp directory. That use is separate from runtime `mch` planning behavior, where temporary planning files must use the loaded `.mch/config.yaml` `temp_dir` value and must not fall back to a built-in temp path.

After every `mch` code change, agents must run `make lint` from `cli` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the `mch` code change.

Changes to the `mch` AI-assisted `/new-change` and idea edit flows should also be verified with tests that fake editor, backend, Codex execution, and Git execution. Coverage should include temporary planning files, `/resume`, `/clear`, and `/cancel`, unparsable idea titles with `error parsing title:`, `/edit`, and `/cancel`, raw idea previews before parse errors and confirmation prompts, `Create Change?` No as a no-op, create-before-rewrite behavior, update-before-rewrite behavior, rewrite failures, title plus idea create payload fields, `update-idea-agent-edit` saves, backend save failures, absence of unsupported Flow assignment and branch reconciliation commands, and refreshed detail navigation.

Changes to `mch` configuration loading should verify repository-root resolution from nested directories, `.mch/config.yaml` parsing, `.mch/default/flow.yaml` and `.mch/default/help.yaml` parsing, configured `temp_dir` use, shared numeric `project_id` persistence, `/config` command routing, read-only `/config` rendering from in-memory structs, exact hook command display without execution, custom Flow help option slugs, empty or missing Flow help option groups, duplicate Flow Step slug rejection, missing or unsupported Flow Step mode rejection, and verbose startup errors for missing or malformed `.mch` files without fallback to old config paths.

## Frontend
From the repository root:

```sh
pnpm --dir frontend test
pnpm --dir frontend typecheck
pnpm --dir frontend build
```

Frontend checks should cover feature logic, visible component behavior, routing, and project-scoped refresh behavior. Change UI coverage should include title plus idea create flows, optional `spec`, optional PR fields, empty type arrays, unassigned reference rendering, and detail reload behavior.

## Documentation
Documentation checks should list files, enforce the 300-line limit, and run the vocabulary checks from the active Change spec.

For CLI config Changes, documentation checks should confirm active docs describe repository-root `.mch/config.yaml`, `.mch/default` Flow loading, `/config`, `temp_dir`, and no active `cli/.config/config.yaml` contract.

For Change workflow branch or spec-path changes, documentation checks should confirm active behavior docs use `change/<change-name>` for repository workflow branches and `specs/<change-name>.md` for Change specs. Remaining `changes/` matches must be application routes, package paths, or historical research/spec context.

Personal research under `docs/research` is not product documentation and is excluded from rewrite verification.
