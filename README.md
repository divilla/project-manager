# project-manager

Project Manager is a local-first software planning application for developers. It replaces
estimate-driven planning with verified completeness: the product asks what complete means, records
that as test cases, and tracks progress from evidence.

## Quick Start

With PostgreSQL configured and the Go and frontend dependencies installed, start the backend and
frontend from the repository root:

```sh
make run
```

The root [Makefile](Makefile) owns local setup and run entry points. Database targets are explicit,
destructive operations; inspect the target and database URL before invoking them.

## Repository Layout

- `backend/` — Go API service; executable checks are in [backend/Makefile](backend/Makefile).
- `frontend/` — Vue and Quasar application; commands are in
  [frontend/package.json](frontend/package.json).
- `cli/` — Bubble Tea terminal application; executable checks are in
  [cli/Makefile](cli/Makefile).
- `db/` — PostgreSQL source and explicitly operated backup helpers.
- `.mch/default/` — default Change Flow, prompts, configuration, and workflow commands.
- `agent/` and `specs/` — historical and active Change artifacts.
- `docs/decisions/` — durable architectural decisions.
- `docs/operations/` — operational procedures that cannot safely live in executable commands.
- `docs/research/` — non-authoritative research history.

Use the linked Makefiles and package scripts for setup, verification, and build commands; they are
the executable command reference.
