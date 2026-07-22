# Refine Default Scripts and Prompts

## Summary
- Use one immutable Change UUID across CLI creation, the API, backend persistence, and database records, preserving caller-supplied identities and generating UUIDv7 values when omitted.
- Rework CLI artifact writing around shared `.mch/default` and `.mch/tmp` paths, UUID-scoped idea workspaces, startup-cached Change types, and a common optional `Types:` metadata contract.
- Refine the default Flow prompts, session scripts, branch helpers, promotion guards, workflow documentation, and automated coverage to match the resulting behavior.

## Behavior
- `POST /api/v1/change/create` accepts optional `ref_uuid`; omitted and explicit `null` values generate a UUIDv7, supplied UUIDs are preserved, malformed values fail binding, and later mutations cannot replace the persisted identity.
- `POST /api/v1/change/update-change-types` normalizes submitted arrays, silently filters unsupported slugs against backend options, and replaces the complete type set with the filtered result, including an empty array.
- `/new-change` generates its UUID before opening the editor, creates blank `input.md` and `output.md` files under `.mch/tmp/<ref_uuid>/idea`, confirms before validation, and cleans up only pre-persistence cancellations and initialization failures.
- Successful creation becomes the persistence boundary: later copy, type-update, or rewrite failures recover against the existing Change and never issue a second create request.
- Existing Idea, Spec, and PR edits reload backend state, reuse the UUID-scoped `idea` workspace and saved session, skip unchanged output, and sequence user save, type update, reload, prompt execution, agent save, and final reload.
- `Types:` may be omitted, empty, or populated in Idea, Spec, and PR artifacts. Artifact parsing does not query the type catalog or reject unsupported values, and `Epic:` is treated as ordinary artifact text.
- CLI-launched artifact processes receive fixed `MCH_DEFAULT_DIR`, `MCH_TEMP_DIR`, `MCH_REF_UUID`, and `MCH_STAGE=idea` values without executing configured Flow hooks or Make targets.
- `.mch/config.yaml` no longer owns `temp_dir`; legacy values are ignored and runtime workspaces always resolve beneath the repository `.mch/tmp` directory.

## Data Model and API
- `fn_change_insert` now accepts `_project_id`, `_ref_uuid`, `_title`, and `_idea`, stores the resolved UUID with the Change, and still creates the initial Idea history row in the same transaction.
- Demo seeding supplies a generated UUID to the four-argument function, while the existing unique `change_ref_uuid_index` remains the duplicate-identity boundary.
- Backend service tests and Change API tests cover generated, supplied, malformed, duplicate, filtered-type, response, and immutability behavior.

## Default Flow
- Renames and consolidates workflow prompts around `idea-write`, `spec-write`, `spec-review`, `pr-write`, documentation, implementation, review, fix, and reconciliation responsibilities.
- Publishes startup-loaded Change type slugs to `prompts/change-types.md`; writing prompts consume the catalog while the canonical Spec structure validates optional metadata syntax independently.
- Uses stage-local `session-id`, `agent-output.md`, `events.jsonl`, and `error.log` artifacts, renders `/stg-tmp-dir/` and `/def-dir/` placeholders, and preserves Codex failures through new and restored sessions.
- Adds guarded Change branch initialization, shared slug validation, and master promotion checks for dirty trees, stale refs, divergent history, and remote movement.
- Keeps the default Makefile as an external Flow harness; the CLI does not invoke it at runtime.

## Verification
- Passed: `(cd backend && GOCACHE=/tmp/project-manager-pr-backend-cache make test)`
- Passed: `(cd cli && GOCACHE=/tmp/project-manager-pr-cli-cache make test)`
- Passed: `(cd cli && GOCACHE=/tmp/project-manager-pr-cli-cache make integration-test)`
- Passed: `(cd cli && GOCACHE=/tmp/project-manager-pr-cli-cache make terminal-test)`
- Passed: `bash -n .mch/default/scripts/*.sh`
- Passed: `git diff --check`
- The first sandboxed CLI integration attempt failed because `httptest` could not bind `[::1]:0` (`operation not permitted`); the same command passed when local socket binding was allowed.
- Not run during PR drafting: backend and CLI lint, vet, and race targets; backend API integration tests.

## Risks
- Backend deployment must use the matching four-argument `fn_change_insert` contract before serving create requests from this version.
- Pre-persistence cancellation removes the generated UUID workspace, while post-persistence recovery deliberately retains it; reviewers should keep that boundary intact.
- CLI artifact writes intentionally remain in `MCH_STAGE=idea` for this Change; deriving stages from Flow steps is deferred.

## References
- `specs/120-refine-default-scripts-prompts.md`
- `AGENTS.md`
- `docs/concepts.md`
- `docs/docs-rules.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/pr-integration.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/operations/verification.md`
