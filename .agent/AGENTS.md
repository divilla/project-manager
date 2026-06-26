# AGENTS.md

This file provides guidance to Agent when working with code in this repository. Use the
`project-manager-change-workflow` skill for Change workflow prompts:

- `make-change new NAME`
- `make-change commit`
- `make-change implement`
- `make-change pr`

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
- Change files are stored as `agent/changes/001-change-name.md`.
- Change files must use the standard structure from the Change workflow:
  Goal, Scope, Requirements, Acceptance Criteria, Non-Goals, Design Notes,
  Relevant Specs, Verification, Review Focus, and Follow-Ups.
- Change branches use `changes/001-change-name`.
- If implementation or PR work starts on a branch other than `changes/<change-name>`, stop and alert the user.
- Change lifecycle: backlog -> branch/rejected -> pull-request -> stage/rejected -> master/rejected.
- The Change file is the PR contract. Do not implement before the user says `make-change implement`.
- Keep implementation scoped to the active Change. Record useful out-of-scope work as Follow-Ups instead of expanding the PR.

## GitHub PR Reviews

When explicitly asked to review a PR, the agent must post the review comment with `gh`, but only for repositories owned by the user's GitHub account `divilla`.

Strong constraints:

- Do not post PR comments unless the user explicitly requested a review.
- Do not post PR comments on repositories outside the `divilla` account/organization.

## Database

- AI agents must never edit, create, delete, move, or otherwise alter files under the repository-root `db` folder or any of its subfolders unless the user explicitly instructs that exact database-file change.
- AI agents must never run `create`, `alter`, `drop`, `truncate`, `grant`, `revoke`, migration, restore, or any other PostgreSQL command that changes database structure unless the user explicitly instructs that exact database-structure change.
- If the database contract appears wrong, blocks implementation, or causes a test hang or failure, report the database blocker to the user instead of changing SQL files or mutating live database structure.
- Use simple, conventional transactions (`Begin`, deferred `Rollback`, and `Commit`) to keep multi-step mutations atomic.
- Do not introduce project-wide or aggregate locking protocols, advisory locks, isolation-level escalation, or coordinated locking across repository paths unless explicitly requested.
- Prefer the simpler transaction design when stronger concurrency control would add substantial implementation and maintenance complexity. Accept the documented concurrency trade-off until requirements justify that complexity.

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
- Tests include service unit tests, integration tests, and benchmarks
- Use `backend/api-tests` folder for all integration tests
- Add subfolder to `backend/api-tests` for each api group specified in code like `project`, `task`, etc...
- Race condition testing is required (`make race`)
- Test files follow `*_test.go` naming convention
- Build all test types for all the code built by AI or fix existing tests
