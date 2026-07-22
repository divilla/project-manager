# AGENTS.md

This file provides guidance to Agent when working with code in this repository. Use the
`change-*` set of skills for Change workflow prompts:

- `$change-spec`
- `$change-verify`
- `$change-docs`
- `$change-code`
- `$change-review`
- `$change-fix`
- `$change-pr`

AGENTS.md file must never be altered unless there is an explicit prompt to override rule and make change in AGENTS.md.

## Communication style

- Use relaxed, conversational language.
- Avoid corporate or overly formal phrasing.
- Keep explanations concise and natural.
- Technical accuracy still takes priority.

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
- The PR is the de facto Change and the single source of truth for the Change
- Before a PR is published, the complete branch diff is the prospective PR and provisional source of truth
- Code in the PR defines current behavior and technical contracts
- Documentation must follow the PR code and accurately describe its observable behavior and constraints
- When documentation conflicts with code, update the documentation; do not change code solely to match stale documentation
- Documentation must not be overly detailed and a single doc file has a maximum of 300 lines
- Documentation rules are defined in `docs/docs-rules.md`

### Changes

- A Change is the basic unit of work in this workflow.
- Change specs are stored as `specs/<change-name>.md`.
- Change specs must use the standard structure from the Change workflow:
  Goal, Scope, Requirements, Non-Goals, Design Notes,
  Verification, QA Test Cases, Review Focus, and Follow-Ups.
- Change branches use `change/<change-name>`.
- If implementation or PR work starts on a branch other than `change/<change-name>`, stop and alert the user.
- Change lifecycle: backlog -> branch/rejected -> pull-request -> stage/rejected -> master/rejected.
- The PR is the de facto Change and its full diff is the Change contract.
- Before publication, treat the complete branch diff as the prospective PR.
- Before implementation, the Change spec is a provisional implementation guide derived from the Idea, existing code, and any branch changes already applied.
- Once the PR exists, it becomes authoritative and the Change spec becomes a structured representation of it.
- Final Spec reconciliation must include all material PR behavior regardless of whether a developer, agent, or other process applied it before or after the original Spec was written.
- The reconciled Change spec must follow accepted PR behavior and must never override the PR.
- Keep implementation scoped to the active PR. Record useful work not present in the PR as Follow-Ups instead of describing it as part of the Change.

## GitHub PR Reviews

When explicitly asked to review a PR, the agent must post the review comment with `gh`, but only for repositories owned by the user's GitHub account `divilla`.

## Review guidelines

When reviewing a PR, build fresh context from the repository instead of conversation memory:

- Inspect the full PR diff against its base branch first.
- Read the active Change spec as a structured representation of that PR.
- Identify changed public contracts, data model changes, migrations, tests, docs, and workflows.
- Run or inspect the listed verification commands when feasible.
- Verify that every material PR change is represented by the Spec and that every Spec Requirement is supported by the PR diff and tests.
- Treat a Spec/PR mismatch as a Spec reconciliation issue unless the PR behavior itself is incorrect.

Prioritize findings only. Focus on correctness bugs, behavioral regressions, data loss or migration risk, security or privacy issues, broken API/UI contracts, missing tests for changed behavior, and brittle tests that can pass while behavior is broken.

For each finding include severity (`P0`, `P1`, `P2`, or `P3`), file and line reference, concrete impact, and a specific fix direction. Do not list style nits, preferences, praise, or summaries unless there are no findings. If no blocking issues exist, say exactly `No blocking issues found.`

Strong constraints:

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
- Uses goimports, golint, and staticcheck for code quality

## Layer Boundaries

API handlers must contain the minimum code required to work with Echo and HTTP concerns only. Handler code may bind request bodies, read Echo context, access `http.Request` or `http.Response` data, choose HTTP status codes, and encode responses. Echo context, `http.Request`, `http.Response`, and other transport-specific objects must never leak into the service layer. All validation, normalization, authorization decisions, business rules, orchestration, and application behavior must move to the service layer unless the code exists only because it directly uses Echo or HTTP objects.

Repositories must contain the minimum code required to work with database access only. Repository code may execute SQL, call database functions or procedures, manage driver-specific scanning, map database rows and driver types into plain Go values or DTOs, and hide database errors behind domain errors where appropriate. Database drivers, driver-specific types, SQL rows, transactions, and query builders must never be accessible from the service layer. Any business rule, comparison, validation, orchestration, or endpoint-specific behavior must move to the service layer unless the code exists only because it directly uses database drivers.

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

## CLI Testing

The backend service-only unit-test rule does not prohibit CLI model tests. CLI Changes must use
all applicable test tiers below in the same Change.

### Model Tests

- Keep model tests beside the CLI code under `cli/internal/**` using `*_test.go`.
- Drive Bubble Tea models with messages such as `tea.KeyMsg`, execute returned commands, and assert
  state, rendered views, client calls, and side effects.
- Add or update model tests whenever a Change modifies CLI state, update logic, commands,
  navigation, rendering, or error handling.
- Model tests are fast behavioral tests, but they do not satisfy Spec QA Test Cases by themselves.

### Program Integration Tests

- Store CLI program integration tests under `cli/integration/**`.
- Exercise the complete CLI program through controlled input and output, using temporary
  configuration and files plus fake clients or `httptest.Server` for backend behavior.
- Every CLI scenario under the Spec's `QA Test Cases` must map to at least one named program
  integration test. Unit tests, model tests, and manual checks cannot replace this coverage.
- Assert observable output, exit behavior, requests, persisted files, and other side effects
  required by the QA scenario.
- Do not contact live services or mutate a live/local database from CLI integration tests.

### PTY Terminal Tests

- Store pseudo-terminal tests under `cli/integration/terminal/**`.
- Build and launch the real `mch` executable inside a PTY, set a fixed terminal size and terminal
  environment, send actual key bytes or escape sequences, and assert rendered terminal output,
  process exit, and externally visible side effects.
- Add PTY coverage whenever behavior depends on TTY detection, raw mode, key-sequence decoding,
  terminal resizing, ANSI colors, cursor movement, full-screen redraws, scrolling, or interactive
  subprocess/editor transitions.
- PTY tests are additional coverage and do not replace the program integration test required for
  each CLI QA Test Case.
- Use bounded timeouts and read-until-marker synchronization instead of fixed sleeps. Always clean
  up child processes and PTYs, including on failure.
- Stub editors and other external commands with controlled test executables resolved through a
  test-specific `PATH`; never open real interactive tools during automated tests.

### CLI Test Harness and Commands

- Test infrastructure required to cover an in-scope QA Test Case is part of that Change unless the
  Spec explicitly excludes it. If such an exclusion makes required QA coverage impossible, stop
  and report the Spec conflict instead of omitting or downgrading the test.
- `cli/Makefile` must expose `integration-test` for program integration tests and `terminal-test`
  for PTY tests once those tiers exist. `make check` must run every applicable CLI test tier.
- Run focused model, program integration, and PTY tests while implementing, then run every
  applicable CLI test command until all tests pass.

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
