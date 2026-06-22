# AGENTS.md

This file provides guidance to Agent when working with code in this repository.

## About This Project Backend

Echo is a high performance, minimalist Go web framework. This is the main repository for Echo v4, which is available as a Go module at `github.com/labstack/echo/v4`.

## Backend Development Commands

The project uses a Makefile for common development tasks:

- `make check` - Run linting, vetting, and race condition tests (default target)
- `make init` - Install required linting tools (golint, staticcheck)
- `make lint` - Run staticcheck and golint
- `make vet` - Run go vet
- `make test` - Run short tests
- `make race` - Run tests with race detector
- `make benchmark` - Run benchmarks

Example commands for development:
```bash
# Setup development environment
make init

# Run all checks (lint, vet, race)
make check

# Run specific tests
go test ./middleware/...
go test -race ./...

# Run benchmarks
make benchmark
```

## Database

- no database objects should be created, altered or dropped (refactored, optimized) under any circumstances, unless there is explicit command to do so in prompt or specification.
- Use simple, conventional transactions (`Begin`, deferred `Rollback`, and `Commit`) to keep multi-step mutations atomic.
- Do not introduce project-wide or aggregate locking protocols, advisory locks, isolation-level escalation, or coordinated locking across repository paths unless explicitly requested.
- Prefer the simpler transaction design when stronger concurrency control would add substantial implementation and maintenance complexity. Accept the documented concurrency trade-off until requirements justify that complexity.

## Backend Code Architecture

### Core Packages

* **Echo (`[Echo](https://github.com/labstack/echo)`)**
* **Zerolog (`[Zero Allocation JSON Logger](https://github.com/rs/zerolog)`)**
* **PGX (`[pgx - PostgreSQL Driver and Toolkit](https://github.com/jackc/pgx)`)**
* **Google UUID (`[uuid](https://github.com/google/uuid)`)**
* **Gookit config (`[Config](https://github.com/gookit/config)`)**
* **Testify (`[Testify - Thou Shalt Write Tests](https://github.com/stretchr/testify)`)**

## Backend File Organization

- `backend/`: Backend working directory
- `backend/cmd`: All the main and starter files
- `backend/internal/`: All the domain logic with Screaming Architecture is a software design philosophy coined by Robert C. Martin (Uncle Bob) that dictates a system's folder and code structure should immediately communicate its business purpose, rather than the technology stack it uses
- `backend/pkg`: Package wrappers

## Backend API

- Make all API endpoints POST
- Only keep /health GET

## Code Style

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
