# Command Line Interface

Best practice is to create scripts with predictable names, arguments, output, and exit codes, then document them for the AI.

Example:

./scripts/test-backend
./scripts/test-frontend
./scripts/check-pr
./scripts/start-dev

Each script should:

- Exit 0 on success, non-zero on failure.
- Print concise, useful errors.
- Avoid interactive prompts.
- Be safe to rerun.
- Avoid destructive behavior unless explicitly named and documented.
- Accept flags instead of asking questions.

Then add a repo instruction file, for example AGENTS.md:

## Common Commands

- `make test` from `backend/`: runs backend unit tests.
- `make api-test` from `backend/`: starts test API server and runs API tests.
- `pnpm --dir frontend test`: runs frontend unit tests.
- `pnpm --dir frontend typecheck`: runs TypeScript checks.
- `pnpm --dir frontend build`: runs production frontend build.

## PR Verification

Before finalizing a PR review or implementation, run:

```sh
make test
make api-test
pnpm --dir frontend test
pnpm --dir frontend typecheck
pnpm --dir frontend build


You can also make a single orchestration script:

```bash
./scripts/check-pr

Then tell the AI:

> Use ./scripts/check-pr as the authoritative PR verification command.

That works very well because it removes ambiguity. The AI does not need to infer which commands matter.

The deeper point:

> AI is much better at using tools when the tools are small, deterministic, documented, and composable.
