# CLI Agent Integration

## Summary
- Adds the AI-assisted `/new-change` flow from `ChangesListState`, using `/tmp/mch` planning files and Codex CLI handoff to turn a Markdown idea into a backend Change.
- Moves planning ownership into `cli/internal/agent` with workspace, runner, parser, command-output formatting, and session ID extraction support.
- Updates CLI docs, backend API docs, Change prompts, and tests for backend-provided type options, generated `Types:` metadata, and the `/api/v1/change/create` contract.

## Behavior
- `/new-change` now validates the numeric current project before opening an editor, prepares `/tmp/mch`, and supports `/resume`, `/clear`, and `/cancel` when an existing idea file is present.
- The idea and review stages use the existing full-screen editor handoff; empty ideas and unchanged resumed ideas return to the list without Codex or backend create calls.
- Codex rewrite runs from the repository root with `Use $change-idea-tmp.`, writes `/tmp/mch/codex-run.jsonl` and `/tmp/mch/codex-output.txt`, extracts `thread.started.thread_id` with legacy fallbacks, and requires exact output `Done.`.
- Unchanged review starts interactive `codex resume <session_id> 'Use $change-spec-tmp.'`; the generated `/tmp/mch/initial-change.md` is parsed into title, full body, `change_types`, and ordered QA test case scenarios.
- Successful generation creates the Change through `POST /api/v1/change/create`, creates one backend test case per generated QA scenario through `POST /api/v1/test-case/create`, and reloads the created Change detail before rendering success.
- Parse, editor, Codex, backend create, test case create, and reload failures stay recoverable and do not display local-only success state.

## CLI
- Startup loads phase and type options from the backend option endpoints so filters, selectors, manual Change parsing, and generated Change parsing use database-provided options instead of hardcoded slugs.
- Change list and detail phase rendering now use backend-provided phase color metadata with a neutral fallback.
- `AgentRunningScreen` replaces the Change list while Codex runs and shows a spinner, elapsed seconds, and formatted command output.

## Docs
- Documents the `/new-change` planning flow, `/tmp/mch` workspace, Codex command handoff, strict generated Change parsing, backend create payload, generated QA test case persistence, and no local-only success behavior.
- Updates the backend API and lifecycle docs to keep `/api/v1/change/create`, backend-owned identity fields, backend default phase, and option endpoint usage explicit.
- Updates Change generation prompts so generated Change files start with an H1 title followed by `Types: <type-slugs>` using backend type slugs, and replaces the old Change build prompt with a branch-derived init prompt.

## Verification
- Passed: `wc -l docs/architecture/mch.md docs/architecture/cli.md docs/functionality/agent-interaction.md docs/functionality/change-lifecycle.md docs/functionality/current-project-context.md docs/functionality/requirements-and-acceptance.md docs/operations/verification.md`
- Passed: `rg -n "api/v3|internal/planning" docs/architecture/mch.md docs/architecture/cli.md docs/functionality/agent-interaction.md cli` returned no matches.
- Failed: `git diff --check origin/stage...HEAD` reported trailing whitespace in `agent/ideas/111-cli-agent-integration.md` and `agent/prompts/change-file-init-prompt.md`.
- Not run while drafting this PR body: `cd cli && make lint`
- Not run while drafting this PR body: `cd cli && go test ./...`
- Not run while drafting this PR body: `cd cli && go build -o /tmp/mch ./cmd/mch`

## References
- `agent/changes/111-cli-agent-integration.md`
- `agent/prompts/change-file-structure.md`
- `agent/prompts/change-file-init-prompt.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/architecture/system-architecture.md`
- `docs/concepts.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/current-project-context.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/operations/verification.md`
