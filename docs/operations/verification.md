# Verification

## Backend
From `backend`:

```sh
make lint
make test
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics. Change API coverage should include title plus idea creation, generated `ref_uuid`, missing or blank title and idea validation, null and empty artifact update rejection, empty-string artifact response values, `agent_edit` on idea/spec/PR update payloads, empty `change_types`, unassigned `ref` and `slug`, Flow assignment, repeated Flow assignment that preserves the existing `ref` and copied Flow arrays, removal of old Change routes, Run claiming, duplicate claim handling, stale claim rejection, completed Run handling, invalid Run IDs and claim IDs, informational Run values, and Change detail response field shapes.

After every backend code change, agents must run `make lint` from `backend` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the backend code change.

`make api-test` runs API integration tests from `backend/api-tests`. These tests exercise backend endpoints over HTTP and should be organized by API endpoint group.

`make api-test` recreates the disposable `changes_test` database from `db/init.sql` and `db/seed.sql`, then starts a temporary backend on port `19080` with `-db postgres://postgres:postgres@localhost:5432/changes_test` and `-port 19080`. If that port is already in use, run `API_TEST_PORT=19081 make api-test` to use a different local port.

API integration tests must interact with the backend only through HTTP requests and responses. They may read the runner-provided HTTP base URL, but must not open database connections, run SQL, or inspect tables directly.

## `mch`
From `cli`:

```sh
make lint
go test ./...
make race
go build -o /tmp/mch ./cmd/mch
```

Verification commands may write build outputs and fixtures under the system temp directory. That use is separate from runtime `mch` planning behavior, where temporary planning files must use the loaded `.mch/config.yaml` `temp_dir` value and must not fall back to a built-in temp path.

After every `mch` code change, agents must run `make lint` from `cli` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the `mch` code change.

Changes to the `mch` AI-assisted `/new-change` and idea edit flows should also be verified with tests that fake editor, backend, Codex execution, and Git execution. Coverage should include temporary planning files, `/resume`, `/clear`, and `/cancel`, unparsable idea titles with `error parsing title:`, `/edit`, and `/cancel`, raw idea previews before parse errors and confirmation prompts, `Create Change?` No as a no-op, create-before-rewrite behavior, update-before-rewrite behavior, rewrite failures, title plus idea create payload fields, `update-idea` saves with `agent_edit` false for user edits and true for agent rewrites, backend save failures, absence of unsupported Flow assignment and branch reconciliation commands, and refreshed detail navigation.

Changes to `mch` configuration loading should verify repository-root resolution from nested directories, `.mch/config.yaml` parsing, `.mch/default/flow.yaml` and `.mch/default/help.yaml` parsing, configured `temp_dir` use, shared numeric `project_id` persistence, `/config` command routing, read-only `/config` rendering from in-memory structs, exact hook command display without execution, custom Flow help option slugs, empty or missing Flow help option groups, duplicate Flow Step slug rejection, missing or unsupported Flow Step mode rejection, and verbose startup errors for missing or malformed `.mch` files without fallback to old config paths.

Reusable Flow runtime tests should construct fresh conformance definitions through the composition boundary and cover validation of identifiers, artifacts, Screen and task kinds, allowed task sequences, required and forbidden fields, consistent Step artifacts, commands, and typed Step versus terminal Screen destinations. Invalid definitions must produce field- or reference-specific errors before execution.

Screen and lifecycle tests should use a configured `temp_dir` and fake stores and processes. They should cover exact load bytes in `input.md` and `output.md`, immutable Step baselines, fresh loads for destination Steps, Exec-to-Interactive and Interactive-to-Editor reuse, unchanged no-save completion, changed save-before-Preview ordering, and no persistence after stop, cancel, failure, or save failure. Cover standalone Editor behavior; Exec exact final-line evaluation and cancellation; Interactive entry, resume, edit, and cancel; Preview and Diff rendering for Idea, Spec, and PR including Git statuses; and Error `/return` behavior for every failure class.

Persistence adapter tests should load Changes through the existing CLI API boundary and verify Idea, Spec, and PR saves select only their matching focused endpoint with the active Change ID; Editor-initiated saves must send `agent_edit: false`, and unsupported artifacts must be rejected. Navigation tests should prove typed `step` destinations perform a fresh load and typed `screen` destinations navigate directly without one. External editor, preview, diff, Exec, Interactive, and API tests must use fakes so model tests do not start Codex or contact a real backend.

Compatibility tests should confirm the reusable runtime remains uncomposed: `/new-change`, Change-detail editing, `.mch/default/flow.yaml` loading, and `/config` retain their existing behavior and do not enter a conformance Flow.

## Frontend
From the repository root:

```sh
pnpm --dir frontend test
pnpm --dir frontend typecheck
pnpm --dir frontend build
```

Frontend checks should cover feature logic, visible component behavior, routing, and project-scoped refresh behavior. Change UI coverage should include title plus idea create flows, string `spec`, string PR fields, empty artifact rendering, `ref_uuid` handling, empty type arrays, unassigned reference rendering, and detail reload behavior.

## Documentation
Documentation checks should list files, enforce the 300-line limit, and run the vocabulary checks from the active Change spec.

For Change artifact contract changes, documentation checks should confirm active docs describe `ref_uuid` as read-only response data, string artifact fields with empty strings as the no-value state, and `agent_edit` on `update-idea`, `update-spec`, and `update-pr` payloads.

For CLI config Changes, documentation checks should confirm active docs describe repository-root `.mch/config.yaml`, `.mch/default` Flow loading, `/config`, `temp_dir`, and no active `cli/.config/config.yaml` contract.

For Change workflow branch or spec-path changes, documentation checks should confirm active behavior docs use `change/<change-slug>` for repository workflow branches, `agent/ideas/<change-slug>.md` for ideas, `specs/<change-slug>.md` for Change specs, and `.mch/default/prompts/spec-file-structure.md` for the spec structure template. Remaining `changes/` matches must be application routes, package paths, or historical research/spec context.

Personal research under `docs/research` is not product documentation and is excluded from rewrite verification.
