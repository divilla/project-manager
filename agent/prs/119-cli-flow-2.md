# CLI Flow 2: Complete Idea Stage

## Summary
- Completes the CLI Idea Stage by composing the reusable Flow runtime into Idea creation, editing, Rewrite, Review, Chat, Preview, cancellation, and error paths.
- Replaces resumable pre-create drafts and the legacy post-save workflow with isolated UUID-scoped creation attempts, a built-in typed Idea definition, generic `DocEdit`, and ordered Step Task traversal.
- Adds universal artifact validation and canonicalization, provenance-aware focused saves, repository-root workspace enforcement, updated Codex scripts, comprehensive CLI tests, and the supporting architecture and verification documentation.

## Behavior
- Every `/new-change` invocation atomically creates a fresh `.mch/tmp/<attempt-uuid>/new-idea.md`, correlates asynchronous replies by UUID, retries directory collisions, and cleans up the attempt without recovering an older draft.
- Idea creation validates a one-line H1, required non-whitespace body, and optional API-backed `Types:` and `Epic:` metadata before confirmation; accepted content is canonicalized before persistence, and only this create path initializes the Change title from the Idea H1.
- Change-details Idea editing now uses generic `DocEdit`: it loads an isolated byte-identical Editor baseline, canonicalizes before comparison, skips byte-identical saves, persists changed user output with `agent_edit: false`, and enters Preview before Rewrite.
- Rewrite and Review execute as ordered `ExecTask -> ChatTask` Steps. Matching final output skips Chat and reaches Preview; empty or unexpected output advances to the same Step's Chat fallback, while Preview `/chat` reactivates that Chat Task without restarting the Step.
- Exec and Chat canonicalize and save changed artifact output with `agent_edit: true`; artifact titles and metadata remain independent from Change title, types, Epic, slug, and other artifacts.
- Preview commands are mode-aware: Editor Preview exposes `/continue`, `/edit`, and `/cancel`, while Rewrite and Review Preview also expose `/chat`. Review `/continue` temporarily exits the Flow to `MainState` without starting Spec work.
- Caller-aware Editor returns preserve the exact invoking Details, Preview, or Chat Screen and its comparison files when canonical output is byte-identical. Validation, workspace, option-loading, persistence, editor, execution, and session failures enter the configured Error path without false success navigation.

## Runtime and Workspace
- Makes every Step own a non-empty ordered Task collection whose exact shape is validated by Mode, tracks active Step separately from Task position, and renames the reusable resumable agent-instruction vocabulary and boundary to Chat.
- Uses `const TmpDir = ".mch/tmp"` as the single executable temp-root value, removes `temp_dir` from repository configuration and `/config`, and requires `.mch` plus `.mch/tmp` to exist as directories at startup.
- Updates execution and resume scripts to derive Change- and artifact-scoped workspaces from `MCH_TEMP_DIR`, `MCH_REF_UUID`, and `MCH_ARTIFACT`, while removing superseded session and legacy agent-flow scripts and code.

## Tests
- Covers built-in definition freshness and validation, exact Mode Task shapes, ordered Exec/Chat traversal, mode-derived Preview `/chat`, typed Step and Screen destinations, caller-aware Editor return, and pending-operation correlation.
- Covers fresh Idea creation attempts, collision retry and cleanup, document validation and metadata canonicalization, conditional option loading, creation confirmation, user-versus-agent save provenance, focused artifact persistence, and Change-field independence.
- Covers workspace isolation, startup prerequisites, script environment contracts, missing temp-root handling, Flow errors, cancellation routes, and removal of the legacy draft, rewrite, and Spec-generation paths.

## Verification
- Passed: `cd cli && make lint`.
- Passed: `cd cli && go test ./...`.
- Passed: `cd cli && make race`.
- Passed: `cd cli && go build -o /tmp/mch ./cmd/mch`.
- Passed: `wc -l docs/docs-rules.md docs/architecture/cli.md docs/architecture/mch.md docs/architecture/backend-api.md docs/functionality/agent-interaction.md docs/functionality/change-lifecycle.md docs/operations/verification.md` (46, 57, 262, 121, 61, 83, and 76 lines respectively).
- Passed: `rg -n "Idea Stage|IdeaRewrite|IdeaReview|Step|Task|ExecTask|ChatTask|ordered|Chat|/chat|agent_edit|MCH_TEMP_DIR|TmpDir|attempt-uuid|new-idea.md|Types:|Epic:|canonical|/continue|/edit|/cancel|MainState|ChangeDetailsState" docs/architecture/cli.md docs/architecture/mch.md docs/architecture/backend-api.md docs/functionality/agent-interaction.md docs/functionality/change-lifecycle.md docs/operations/verification.md`.
- Passed: `! rg -n "temp_dir" .mch/config.yaml cli/internal/app/config.go docs/architecture/cli.md docs/architecture/mch.md`.
- Passed: `! rg -n '<[^>]*_' specs/119-cli-flow-2.md docs/docs-rules.md`.
- Passed: `! rg -n -i -P '(?<!non-)interactive' cli/internal/flow`.
- Passed: `! rg -n "Interactive|/interactive|interactive-session" docs/architecture/cli.md docs/architecture/mch.md docs/operations/verification.md`.
- Passed: `rg -n -F 'const TmpDir = ".mch/tmp"' cli --glob '*.go'`.
- Passed: `! rg -n "initial-idea\\.md|Resume idea\\?" cli`.
- Passed: `! rg -n "RewriteIdeaState|openAgentSpecInit|handleAgentRewrite|agentSpecInit" cli`.
- Passed: `git diff --check origin/stage...HEAD`.

## Risks
- The Idea definition remains built into Go rather than loaded from `.mch/default/flow.yaml`; product Flow configuration is a follow-up.
- Review `/continue` intentionally returns to `MainState` until the Spec Write and review stages are implemented.
- Runtime workspaces isolate different Changes, but the documented invariant still permits only one active Flow process per Change and introduces no cross-process locking.

## References
- `specs/119-cli-flow-2.md`
- `specs/118-cli-flow-1.md`
- `.mch/default/prompts/spec-file-structure.md`
- `docs/docs-rules.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/change-lifecycle.md`
- `docs/operations/verification.md`
