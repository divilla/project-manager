# AGENTS.md

This file provides guidance to Agent when working with code in this repository. Uses skills:
- `new change NAME`
- `commit change`
- `implement change`

## Artifacts

### Epics

- Epic is non-hierarchical group of Changes
- Epic represents large business/product capatibility
- Epics are defined in `agent/epics.md`

### Areas

- Areas are subsytems of this project and they 
- Areas are also folders of main repository
- Areas are defined in `agent/areas.md`

### Documentation

- Documentation is all stored in folder `docs`
- Documentation precisely defines the desired external behavior and constraints
- Documentation are single-source-of-truth for developers - they support every decision relevant to project
- Documentation must not be overly detailed and a single doc file has a maximum of 300 lines
- Documentation rules are defined in `docs/docs-rules.md`

### Changes

- change is a basic unit-of-work (to-do item) in this workflow
- changes are stored in files: `agent/changes/001-change-name.md`
- change files must have exact structure identical to `agent/change-example` - only Goal and Requirements are a mandatory sections - order must be obeyed
- change must have and is indentified in agent with a branch named `change/001-change-name` 
- if during implementation a current branch is set to something other then `change/ddd-xxx` alert user
- change lifecycle: backlog -> branch/rejected -> pull-request -> stage/rejected -> master/rejected

## GitHub PR Reviews

When explicitly asked to review a PR, the agent must post the review comment with `gh`, but only for repositories owned by the user's GitHub account `divilla`.

Strong constraints:

- Do not post PR comments unless the user explicitly requested a review.
- Do not post PR comments on repositories outside the `divilla` account/organization.

## Database

- no database objects should be created, altered or dropped (refactored, optimized) under any circumstances, unless there is explicit command to do so in prompt or specification.
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

- `make check` - Run linting, vetting, and race condition tests (default target)
- `make init` - Install required linting tools (golint, staticcheck)
- `make lint` - Run staticcheck and golint
- `make vet` - Run go vet
- `make test` - Run short tests
- `make race` - Run tests with race detector
- `make benchmark` - Run benchmarks

## Backend Code Architecture

- `backend/`: Backend working directory
- `backend/cmd`: All the main and starter files
- `backend/internal/`: All the domain logic with Screaming Architecture
- `backend/internal/project`: Code structure immediately communicates its business purpose
- `backend/pkg`: Package and other wrappers

### Core External Packagas

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
