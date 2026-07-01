# AGENTS.md

This file provides guidance to Agent when working with code in this repository. Use the
`project-manager-change-workflow` skill for Change workflow prompts:

- `make-change new NAME`
- `make-change commit`
- `make-change implement`
- `make-change pr`

AGENTS.md file must never be altered unless there is an explicit prompt to override rule and make change in AGENTS.md.

## Artifacts

### Epics

- Epic is a non-hierarchical group of Changes
- Epic represents a large business or product capability
- Epics are defined in `agent/epics.md`

### Areas

- Areas are subsystems of this project
- Areas are also folders of the main repository
- Areas are defined in `agent/areas.md`

### Documentation

- Documentation is stored in the `docs` folder
- Documentation precisely defines the desired external behavior and constraints
- Documentation is the single source of truth for developers and supports every decision relevant to the project
- Documentation must not be overly detailed and a single doc file has a maximum of 300 lines
- Documentation rules are defined in `docs/docs-rules.md`

### Changes

- A Change is the basic unit of work in this workflow.
- Change files are stored as `agent/changes/<change-name>.md`.
- Change files must use the standard structure from the Change workflow:
  Goal, Scope, Requirements, Acceptance Criteria, Non-Goals, Design Notes,
  Relevant Specs, Verification, QA Test Cases, Review Focus, and Follow-Ups.
- Change branches use `changes/<change-name>`.
- If implementation or PR work starts on a branch other than `changes/<change-name>`, stop and alert the user.
- Change lifecycle: backlog -> branch/rejected -> pull-request -> stage/rejected -> master/rejected.
- The Change file is the PR contract. Do not implement before the user says `make-change implement`.
- Keep implementation scoped to the active Change. Record useful out-of-scope work as Follow-Ups instead of expanding the PR.

## GitHub PR Reviews

When explicitly asked to review a PR, the agent must post the review comment with `gh`, but only for repositories owned by the user's GitHub account `divilla`.

## Review guidelines

When reviewing a PR, build fresh context from the repository instead of conversation memory:

- Read the active Change file and linked docs.
- Inspect the full diff against the PR base branch.
- Identify changed public contracts, data model changes, migrations, tests, docs, and workflows.
- Run or inspect the listed verification commands when feasible.
- Treat `agent/changes/<change-name>.md` as the PR contract; verify every Requirement and Acceptance Criteria item against the diff and tests.

Prioritize findings only. Focus on correctness bugs, behavioral regressions, data loss or migration risk, security or privacy issues, broken API/UI contracts, missing tests for changed behavior, and brittle tests that can pass while behavior is broken.

For each finding include severity (`P0`, `P1`, `P2`, or `P3`), file and line reference, concrete impact, and a specific fix direction. Do not list style nits, preferences, praise, or summaries unless there are no findings. If no blocking issues exist, say exactly `No blocking issues found.`

Strong constraints:

- Do not post PR comments unless the user explicitly requested a review.
- Do not post PR comments on repositories outside the `divilla` account/organization.

## Database Hard Boundary

AI agents may read database-related source files only for context.

AI agents must never perform any action that writes to, mutates, resets, migrates, seeds, restores, truncates, recreates, or changes data or structure in any database unless the user explicitly instructs that exact database operation.

This ban does not apply to documented repository verification commands that operate only on disposable test databases, such as `(cd backend && make api-test)`.

Agents must never run PostgreSQL structure-changing commands, including `create`, `alter`, `drop`, `truncate`, `grant`, `revoke`, migration, restore, or any SQL file/Make target/test target that may execute those operations, unless the user explicitly
names the exact intended database operation.

Agents must never run read queries against a live or local database unless the user explicitly asks for that exact inspection.

- If the database contract appears wrong, blocks implementation, or causes a test hang or failure, report the database blocker to the user instead of changing SQL files or mutating live database structure.
- AI agents may read files under `db/**`, but must never write to them unless the user gives an explicit instruction naming the exact database file and the exact intended change.
- This ban includes creating, editing, deleting, moving, renaming, formatting, reverting, restoring, conflict-resolving, chmodding, staging generated edits, or applying patches to any file under `db/**`.
- If `db/**` is modified, agents may inspect and review those file changes, but must not edit, stage, revert, execute, seed, migrate, or mutate database files or database state unless explicitly instructed.
- Use simple, conventional transactions (`Begin`, deferred `Rollback`, and `Commit`) to keep multi-step mutations atomic.
- Do not introduce project-wide or aggregate locking protocols, advisory locks, isolation-level escalation, or coordinated locking across repository paths unless explicitly requested.
- Prefer the simpler transaction design when stronger concurrency control would add substantial implementation and maintenance complexity. Accept the documented concurrency trade-off until requirements justify that complexity.
- Do not create foreign keys - this is hard limit

## About Backend

Backend is a classic http API server operating on port 8080 by default.

Example endpoint:
```bash
    curl localhost:8080/api/v1/health
```

## Backend Make Commands

The project uses a Makefile for common development tasks:

- `(cd backend && make check)` - Run linting, vetting, and race condition tests (default target)
- `(cd backend && make init)` - Install required linting tools (golint, staticcheck)
- `(cd backend && make lint)` - Run staticcheck and golint
- `(cd backend && make vet)` - Run go vet
- `(cd backend && make test)` - Run short tests
- `(cd backend && make api-test)` - Run API integration tests
- `(cd backend && make race)` - Run tests with race detector
- `(cd backend && make benchmark)` - Run benchmarks
- `pnpm --dir frontend test` - Run frontend unit tests
- `pnpm --dir frontend typecheck` - Run frontend type checking
- `pnpm --dir frontend build` - Build the frontend

## Backend Code Architecture

- `backend/`: Backend working directory
- `backend/cmd`: All the main and starter files
- `backend/internal/`: All the domain logic with Screaming Architecture
- `backend/internal/project`: Code structure immediately communicates its business purpose
- `backend/pkg`: Package and other wrappers

### Core External Packages

* [Echo](https://github.com/labstack/echo)
* [Zero Allocation JSON Logger](https://github.com/rs/zerolog)
* [pgx - PostgreSQL Driver and Toolkit](https://github.com/jackc/pgx)
* [Config](https://github.com/gookit/config)
* [Testify - Thou Shalt Write Tests](https://github.com/stretchr/testify)

Always use core external packages for all the relevant code built. Warn when core external packages are not used where they might have been used.

## Backend API

- Make all API endpoints POST
- Only keep /health GET

## Backend Code Style

- Go code uses tabs for indentation (per .editorconfig)
- Follows standard Go conventions and formatting
- Uses gofmt, golint, and staticcheck for code quality

## Testing

- Standard Go testing with `testing` package
- Use `github.com/stretchr/testify` wherever possible
- Unit tests are for service-layer behavior only (that includes service.go and *_service.go files)
- Do not write unit tests for API handlers, repositories, config helpers, or other non-service layers.
- API and cross-layer behavior must be covered by integration tests instead of unit tests.
- Tests include service unit tests, API-tests (integration tests), and benchmarks.
- Race condition testing is required (`make race`)
- Test files follow `*_test.go` naming convention
- Build all test types for all the code built by AI or fix existing tests

## API-tests (integration tests)

- Use `backend/api-tests` for all API integration tests.
- Add a subfolder to `backend/api-tests` for each backend API group specified in code, such as `project` or `change`.
- New or changed backend endpoints and endpoint groups require API integration-test coverage in the same Change that introduces them.
- Reviewers must inspect backend endpoint additions for matching API integration tests under `backend/api-tests`.
- Every endpoint must be covered by at least one API-test - all possible request and response fields must be included in tests.

## Test Integrity

Tests are evidence, not obstacles.

If tests fail, agents must treat the failure as a real signal until proven otherwise. Agents must not hide, bypass, weaken, delete, skip, rebaseline, or adapt test setup just to make tests pass.

A failing test is allowed and expected when the implementation, environment, contract, or assumptions are wrong. The agent must preserve that evidence and report it honestly.

Agents must not make verification pass by changing the thing being verified unless the user explicitly requested that exact change and it is within scope of the active task.

Agents must not:
- change schema, seed data, fixtures, mocks, snapshots, golden files, expected values, or test harness behavior just to convert a failure into a pass
- rerun tests after an unauthorized setup change and present the result as valid
- remove or revert the evidence of a failure without reporting it
- claim tests passed if they passed only after unrelated, unauthorized, or hidden changes
- treat a failing integration test as a reason to mutate external state automatically

When tests fail, agents must report:
- the exact command that failed
- the failing test or error message
- the current relevant diff
- whether any local uncommitted changes may have influenced the result
- whether the failure indicates a product bug, test bug, environment issue, or unclear database/contract dependency

If the agent caused the failure or contaminated the verification environment, it must say so plainly and invalidate any affected test result.

Passing tests are not valid evidence if they only passed after the agent changed setup, schema, fixtures, or expectations outside the authorized scope.
