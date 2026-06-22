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

## Backend Code Architecture

### Core Packages

**Echo (`[Echo](https://github.com/labstack/echo)`)**
**Zerolog (`[Zero Allocation JSON Logger](https://github.com/rs/zerolog)`)**
**PGX (`[pgx - PostgreSQL Driver and Toolkit](https://github.com/jackc/pgx)`)**
**Google UUID (`[uuid](https://github.com/google/uuid)`)**
**Gookit config (`[Config](https://github.com/gookit/config)`)**

## Backend File Organization

- `backend/`: Backend working directory
- `backend/cmd`: All the main and starter files
- `backend/internal/`: All the domain logic with Screaming Architecture is a software design philosophy coined by Robert C. Martin (Uncle Bob) that dictates a system's folder and code structure should immediately communicate its business purpose, rather than the technology stack it uses
- `backend/pkg`: Package wrappers

## Code Style

- Go code uses tabs for indentation (per .editorconfig)
- Follows standard Go conventions and formatting
- Uses gofmt, golint, and staticcheck for code quality

## Testing

- Standard Go testing with `testing` package
- Use `github.com/stretchr/testify` wherever possible
- Tests include unit tests, integration tests, and benchmarks
- Race condition testing is required (`make race`)
- Test files follow `*_test.go` naming convention
