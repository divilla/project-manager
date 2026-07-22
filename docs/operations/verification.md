# Verification

## Backend
From `backend`:

```sh
make lint
make vet
make test
make race
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics. Change API coverage should include title plus idea creation; preserved caller-supplied `ref_uuid`; generated UUIDv7 identities for omitted and null values; blank, malformed, and duplicate identity rejection; immutable identity across later mutations; missing or blank title and idea validation; null and empty artifact update rejection; empty-string artifact response values; `agent_edit` on idea/spec/PR update payloads; empty and unsupported `change_types`; unassigned `ref` and `slug`; Flow assignment; repeated Flow assignment that preserves the existing `ref` and copied Flow arrays; removal of old Change routes; Run claiming; duplicate claim handling; stale claim rejection; completed Run handling; invalid Run IDs and claim IDs; informational Run values; and Change detail response field shapes.

After every backend code change, agents must run `make lint` from `backend` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the backend code change.

`make api-test` runs API integration tests from `backend/api-tests`. These tests exercise backend endpoints over HTTP and should be organized by API endpoint group.

`make api-test` recreates the disposable `changes_test` database from `db/init.sql` and `db/seed.sql`, then starts a temporary backend on port `19080` with `-db postgres://postgres:postgres@localhost:5432/changes_test` and `-port 19080`. If that port is already in use, run `API_TEST_PORT=19081 make api-test` to use a different local port.

API integration tests must interact with the backend only through HTTP requests and responses. They may read the runner-provided HTTP base URL, but must not open database connections, run SQL, or inspect tables directly.

## `mch`
From `cli`:

```sh
make lint
make vet
go test ./...
make race
go build -o /tmp/mch ./cmd/mch
```

Verification commands may write build outputs and fixtures under the system temp directory. That use is separate from runtime `mch` planning behavior, where workspaces use the fixed repository-relative `.mch/tmp` root and never a configured or system temporary path.

After every `mch` code change, agents must run `make lint` from `cli` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the `mch` code change.

Changes to `/new-change` should be verified through the complete Bubble Tea program with controlled key input and rendered output, an HTTP test backend, a subprocess-backed fake editor, and an injected agent runner. Coverage should enforce UUIDv7 generation before workspace creation, blank `idea/input.md` and `idea/output.md` creation before editing `output.md`, byte comparison after editor exit, equal-file cancellation with workspace removal, `Create Change?` confirmation before artifact validation, No-choice cleanup without a request, missing-title and blank-body validation after Yes with retained files and no request, optional and unsupported `Types:` values without client-side rejection, `Epic:` as ordinary artifact text, Yes-choice generated `ref_uuid` submission, partial-initialization cleanup, backend failure retry state, and protection of existing Change workspaces. Program integration tests must not call model workflow helpers or construct completion messages directly.

Existing Idea, Spec, and PR editor coverage should reload the Change, use `.mch/tmp/<ref_uuid>/idea`, refresh only `input.md` and `output.md`, preserve the saved session and logs, and skip updates for unchanged output. Changed output must exercise the matching user save, type update when present, Change reload, post-save `output.md`-to-`input.md` promotion, `idea-write`, `spec-write`, or `pr-write` prompt, agent save, and final reload. Every operation must assert `MCH_STAGE=idea`, including Spec and PR writing. Program integration coverage should drive the complete CLI with controlled input/output and fake editor, Codex, and backend processes.

Changes to `mch` configuration loading should verify repository-root resolution from nested directories, `.mch/config.yaml` parsing without `temp_dir`, ignored legacy `temp_dir` values, `.mch/default/flow.yaml` and `.mch/default/help.yaml` parsing, fixed `.mch/tmp` workspace use, shared numeric `project_id` persistence, `/config` command routing, read-only config rendering, exact hook display without execution, custom Flow help option slugs, empty or missing help option groups, Flow Step validation, and verbose startup errors without fallback to old config paths. Artifact-write process checks should verify `MCH_DEFAULT_DIR=.mch/default`, `MCH_TEMP_DIR=.mch/tmp`, the created or loaded Change UUID as `MCH_REF_UUID`, `MCH_STAGE=idea`, inherited-value replacement, and command-failure propagation without Flow-profile stage lookup.

Startup option-catalog checks should verify one Change type request, backend-order caching, defensive copies from `app.ChangeTypes()`, and exact atomic replacement of `.mch/default/prompts/change-types.md`. Failure coverage should preserve the previous cache and file when loading or prompt writing fails and report prompt-write failures through the option-catalog error path. Prompt checks should verify writing prompts use the generated catalog, the Spec structure validates optional `Types:` syntax without the catalog, and no prompt Markdown contains a backend call instruction.

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

For Change artifact contract changes, documentation checks should confirm active docs describe `ref_uuid` as an optional create input that is read-only after creation, including the generated `/new-change` value, string artifact fields with empty strings as the no-value state, and `agent_edit` on `update-idea`, `update-spec`, and `update-pr` payloads.

For CLI config Changes, documentation checks should confirm active docs describe repository-root `.mch/config.yaml`, fixed repository-relative `.mch/tmp`, `.mch/default` Flow loading, `/config` without `temp_dir`, ignored legacy `temp_dir` values, and no active `cli/.config/config.yaml` contract.

For Change workflow branch or spec-path changes, documentation checks should confirm active behavior docs use `change/<change-slug>` for repository workflow branches, `agent/ideas/<change-slug>.md` for ideas, `specs/<change-slug>.md` for Change specs, and `.mch/default/prompts/spec-file-structure.md` for the spec structure template. Remaining `changes/` matches must be application routes, package paths, or historical research/spec context.

Personal research under `docs/research` is not product documentation and is excluded from rewrite verification.
