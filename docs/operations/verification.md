# Verification

## Backend
From `backend`:

```sh
make lint
make test
make api-test
```

Backend checks should cover service logic, repository behavior where feasible, API contracts, history behavior, and health diagnostics. Change API coverage should include title plus idea creation, generated `ref_uuid`, missing or blank title and idea validation, null and empty artifact update rejection, empty-string artifact response values, `agent_edit` on idea/spec/PR update payloads, empty `change_types`, unassigned `ref` and `slug`, first Flow assignment deriving identity from the current Change title, repeated Flow assignment preserving the existing `ref`, `slug`, and copied Flow arrays, title updates preserving assigned identity, removal of old Change routes, Run claiming, duplicate claim handling, stale claim rejection, completed Run handling, invalid Run IDs and claim IDs, informational Run values, and Change detail response field shapes.

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

Verification commands may write build outputs and fixtures under the system temp directory. Runtime `mch` behavior is separate: the Go `TmpDir` constant owns `.mch/tmp`, and no repository configuration value may replace it.

After every `mch` code change, agents must run `make lint` from `cli` and fix all findings before handoff. `make lint` may rewrite imports or formatting; review and include those intentional changes with the `mch` code change.

Idea Stage tests must fake Editor, backend, Exec, Chat, Preview, Diff, scripts, and Git. `/new-change` coverage includes UUID v4 generation, atomic attempt-directory creation without a pre-check, collision retry, isolated zero-byte `new-idea.md`, no draft resume or clear route, zero-byte return without API calls, a cancelable animated Processing state while validation is in flight, UUID-correlated validation and create replies, stale command filesystem isolation, validation before confirmation, canonical bytes in `Create Change?`, No as a no-op, create-before-Rewrite, list origin, two-step file-then-directory cleanup with non-empty-directory preservation, and all editor, option-load, canonical-write, cleanup, and create failures.

Artifact validation coverage includes CRLF-to-LF normalization; one-line H1 and exact blank-line separators; a body containing at least one non-whitespace character; unique optional fields and every Types/Epic order; rejection of title-only, whitespace-body, and metadata-only documents before option loading; metadata-looking body text; conditional option API calls; malformed, duplicate, empty, and unknown values; deterministic backend-ordered `Types: <type-slugs>`; refreshed `Epic: <epic-title> #<epic-id>`; and failure without rewrite or persistence. Only `IdeaCreate` initializes the Change title. Later Idea, Spec, and PR titles and metadata remain independent from the Change and one another.

Configuration tests cover Git-root resolution from nested directories, `.mch/config.yaml` without a temp-root field, shared numeric `project_id`, `.mch/default` parsing, and read-only `/config` rendering without a temp-root setting. Startup must reject missing `.mch`, missing `.mch/tmp`, or either path as a non-directory with a path-specific error, without creating directories or starting the TUI.

Definition tests construct the built-in Idea definition twice to prove independent mutability and shared validation. Every Step owns an ordered `1..n` Task collection. Current valid shapes are exactly Editor Mode with `EditorTask`, Chat Mode with `ChatTask`, and Exec Mode with `ExecTask -> ChatTask`; empty, standalone Exec, reversed, duplicate, extra, Script, and superseded task shapes fail before execution. Tests distinguish `IdeaRewriteExec` and `IdeaReviewExec` Step names from `ExecTask`, select typed Step destinations by activating Task position 1, and track active Step separately from Task position.

Task traversal tests run Exec first, advance empty or unexpected output to Chat without leaving the Step, skip only Chat on matching output, and resolve the next ordered position after either route. Current Rewrite and Review reach Preview only because no later Task exists. `Done.` is the Rewrite match and `No questions or suggestions.` is the Review match. Missing or unreadable `agent-output.md` enters Error, while a readable empty file routes to Chat. Chat entry fresh-loads persisted bytes, preserves its Step session and workspace, displays available output without auto-resume, and keeps interactive transcript output terminal-only.

Generic `DocEdit` tests cover byte-identical initial files under `<artifact>/editor`, immutable Editor input, isolation from the caller's artifact-root comparison, canonicalization before comparison, changed user save and artifact-root publish before Preview, empty-output cancellation, invalid non-empty `/fix` and `/cancel`, save or publish failure, and canonical byte-identical return to the exact recorded terminal, Preview, or Chat caller with its Preview/Diff files intact. Preview and Chat `/edit` must select the same-artifact Editor Step from a fresh persisted baseline and abandon unpersisted proposals.

Preview tests verify `from-step`, rendering, Diff statuses, and exact commands. Editor Preview has `/continue`, `/edit`, `/cancel`; Exec- and Chat-mode Preview has `/continue`, `/chat`, `/edit`, `/cancel`, synthesizing `/chat` when omitted statically. `/chat` activates `from-step`'s same-artifact Chat position without selecting a Step. Invalid continuation, edit, cancel, caller, `from-step`, or redirected Chat destinations fail definition validation. Idea Edit continues to Rewrite, Rewrite continues to Review, and Review temporarily continues to `MainState` without Spec work.

Persistence adapter tests load through the CLI API boundary and verify Idea, Spec, and PR select only their focused endpoint. Editor changes send `agent_edit: false`; Exec and Chat changes send `agent_edit: true`; canonical byte-identical output never saves. Every save completes before Preview or expected-output routing. Artifact saves never call Change title, types, or Epic updates and preserve ref and slug. Unsupported artifacts and provenance-sensitive save failures are rejected without false success.

Workspace boundary tests verify `TmpDir` is passed as `MCH_TEMP_DIR`, scripts require it with `MCH_REF_UUID` and `MCH_ARTIFACT`, Editor input/output remain under `<git-root>/<temp-dir>/<ref-uuid>/<artifact>/editor`, and shared Preview/Exec/Chat, session, agent-output, event, and error files remain under the artifact root. Missing variables, malformed UUIDs, unsupported artifacts, workspace scopes, and fallback paths must fail. Two Change UUIDs remain isolated; tests preserve the one-active-process-per-Change invariant without adding locks.

Navigation and legacy-removal tests cover list and detail entry, Rewrite and Review `/stop`, Chat `/cancel`, every Preview and validation `/cancel`, Error `/return`, fresh loads between Steps, and correct origin versus `ChangeDetailsState` returns. They prove no draft-resume, superseded resumable-agent terminology, specialized post-save rewrite, or Spec-generation entry point remains reachable.

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

For CLI config Changes, documentation checks should confirm active docs describe repository-root `.mch/config.yaml`, required `.mch` and `.mch/tmp` directories, `.mch/default` Flow metadata loading, hardcoded `TmpDir`, and `/config` without a temp-root setting.

For Change workflow branch or spec-path changes, documentation checks should confirm active behavior docs use `change/<change-slug>` for repository workflow branches, `agent/ideas/<change-slug>.md` for ideas, `specs/<change-slug>.md` for Change specs, and `.mch/default/prompts/spec-file-structure.md` for the spec structure template. Remaining `changes/` matches must be application routes, package paths, or historical research/spec context.

Personal research under `docs/research` is not product documentation and is excluded from rewrite verification.
