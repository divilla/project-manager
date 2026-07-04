# Ideas And Change Refactor

## Summary
- Refactors Changes from the former body-required create contract to an idea-first artifact model with required `title` and `idea`, optional `spec`, optional PR artifacts, optional epic, empty `change_types`, and nullable `ref`/`slug`.
- Adds backend, CLI, frontend, docs, database initialization, demo seed, and automated test updates for creation before reference assignment and explicit reference assignment later.
- Preserves the agent execution scaffolding needed for the future spec-writing path while routing the current `/new-change` flow through idea entry, create confirmation, agent rewrite, persisted reload, and details.

## Behavior
- New Changes can be created with only `project_id`, `title`, and `idea`; they default to `backlog`, empty `change_types`, null optional artifacts, and unassigned `ref`/`slug`.
- `POST /api/v1/change/reference` assigns the missing project-scoped `ref` and `slug` for an existing Change, returns refreshed Change data, and preserves an existing `ref` on repeated calls while allowing slug refresh from the current title.
- Focused updates now use `idea`, `spec`, `pr_body`, `pr_url`, title, phase, types, epic, open, and agent-edit endpoints that work before and after reference assignment.

## API
- Replaces rendered body naming with `POST /api/v1/change/rendered-artifacts`, returning rendered `spec_html` and `pr_html`.
- Adds `POST /api/v1/change/update-idea`, `POST /api/v1/change/update-idea-agent-edit`, `POST /api/v1/change/update-spec`, and `POST /api/v1/change/reference`.
- Updates backend and client DTOs so `ref` and `slug` are nullable, `idea` is required, `spec`/PR artifacts are nullable, and `change_types` can be empty.

## Data Model
- Updates `db/init.sql` so `change` stores required `idea`, optional `spec`, nullable `ref`/`slug`, and empty-array defaults for `change_types`.
- Changes `fn_change_insert` to insert only `project_id`, `title`, and `idea`, and moves reference allocation into `sp_change_ref_update`.
- Updates demo seed data to create referenced demo Changes through the new path and to include visible empty-type and null-spec cases.

## CLI
- Adds `CreateIdeaState`, `UpdateIdeaState`, and `RewriteIdeaState` behavior around `/new-change`, `/edit`, and agent rewrite saves.
- Shows raw idea markdown before parse prompts/errors, supports `/edit` and `/cancel`, removes `/tmp/mch/initial-idea.md` on cancel or successful rewrite save, and reloads Change details after persistence.
- Adds `/reference` as the first Change detail command, calling the backend reference endpoint before reconciling local/remote `changes/<slug>` branches.
- Renames detail spec editing to `/edit-spec` while leaving future spec-writing agent code available.

## Frontend
- Updates Change create/edit/detail screens and API types for title-plus-idea creation, optional spec and PR artifacts, empty type arrays, and unassigned `ref`/`slug` rendering.
- Updates rendered artifact API usage and tests so detail and board views do not crash or display placeholder identity for unreferenced Changes.

## Docs
- Updates product and architecture docs for the idea/spec artifact model, nullable identity fields, empty `change_types`, backend reference endpoint, CLI idea rewrite flow, branch reconciliation, and verification expectations.

## Verification
- Not run during this PR draft pass.
- Change file lists intended verification: `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`, `(cd backend && GOCACHE=/tmp/project-manager-go-build make lint)`, `(cd backend && GOCACHE=/tmp/project-manager-go-build make api-test)`, `(cd backend && GOCACHE=/tmp/project-manager-go-build make race)`, `(cd cli && make lint)`, `(cd cli && go test ./...)`, `(cd cli && go build -o /tmp/mch ./cmd/mch)`, `pnpm --dir frontend test`, `pnpm --dir frontend typecheck`, and `pnpm --dir frontend build`.

## Risks
- The branch changes `db/init.sql` and `db/seed-demo.sql`; database-mutating verification was not run while drafting this PR body.
- The branch is broad across backend persistence, API contracts, CLI state transitions, Git branch reconciliation, frontend screens, and docs, so reviewers should focus on cross-layer contract consistency and the reference allocation edge cases.

## References
- `agent/changes/112-ideas-and-change-refactor.md`
- `agent/ideas/112-ideas-and-change-refactor.md`
- `docs/concepts.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/cli.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/mch.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/history.md`
- `docs/operations/verification.md`
- `docs/docs-rules.md`
- `frontend/ARCHITECTURE.md`
