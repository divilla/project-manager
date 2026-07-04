# Application Remake Frontend

## Goal

Refactor the frontend from the old task-based UI and API contract to the current change-based product model delivered by the backend.

## Scope

- Rename frontend feature folders, routes, stores, composables, components, tests, API clients, model types, and visible product copy from task terminology to change terminology.
- Update frontend API usage from `/api/v1/task` routes and old payload fields to `/api/v1/change`, `/api/v1/epic`, and `/api/v1/requirement` routes documented by the backend API.
- Replace change text fields from `name` and `description` behavior to `title` and `body`.
- Replace hierarchical task UI behavior with the fixed structure where a change is standalone or linked to one epic by `epic_id`.
- Add frontend support for epics where the change board, create/edit flows, detail view, and project-scoped refresh behavior need epic data.
- Update requirement UI and state handling so requirements attach to changes by `change_id` and refresh visible completeness from backend mutation responses.
- Update frontend tests to use change naming, new routes, new payload fields, and the current frontend feature structure.

## Requirements

- Frontend source, tests, route paths, type names, API wrappers, stores, fixtures, and visible UI text must not use the old task vocabulary.
- Navigation must expose the Changes workflow using the current application shell and current project selector behavior.
- The Changes route must list project-scoped changes grouped by `change_phase` and use backend reference options from `/api/v1/change/reference`.
- Change create and edit flows must use `title`, `body`, `change_phase`, `change_types`, optional `epic_id`, and `closed`.
- Change detail must render sanitized backend `body_html`, linked requirements, completion counters, phase/type data, and epic context when present.
- Requirement create, edit, done toggle, reassignment, and delete flows must use `change_id` and update visible completeness from backend responses without client-side guessing.
- Project list and delete affordances must use `change_count`; deleting a project with changes must remain blocked and clearly represented in the UI.
- Epic list, create, edit, delete, and conflict states must use `/api/v1/epic` endpoints and must not reintroduce nested change hierarchy behavior.
- Current project switching must refresh project-scoped change, epic, and requirement state and return nested change routes to `/changes`.
- Frontend tests must cover the renamed change workflows, requirement mutation behavior, current project behavior, and destructive confirmation states.

## Acceptance Criteria

- The frontend compiles and passes type checking with the change-based model.
- The frontend test suite passes with no old task fixtures, route names, API paths, or visible copy.
- Top navigation, route definitions, and page labels use Changes rather than Tasks.
- `frontend/src/features/changes` owns the change board/detail/create/edit behavior; no active frontend feature folder remains named `tasks`.
- The UI calls `/api/v1/change/*`, `/api/v1/epic/*`, and `/api/v1/requirement/update-change` instead of removed task endpoints.
- Change forms submit `title` and `body`, never `name` or `description` as change text fields.
- Change detail and requirement flows refresh completion counters from backend responses after requirement mutations.
- The frontend has no active source or test references to the old task hierarchy, `parent_id`, `task_id`, `task_count`, `task_phase`, `task_type`, `/api/v1/task`, `name` as a change title field, or `description` as a change body field.

## Non-Goals

- No agent-authored backend API, schema, seed data, or DTO changes; those are handled by `agent/changes/002-app-remake-backend.md`. Backend, schema, seed, or DTO changes already committed by the user with a `by user` commit message are allowed to remain in this Change branch.
- No authentication, authorization, deployment, or multi-user behavior.
- No new planning or LLM workflow behavior beyond keeping existing frontend calls compatible with the current API contract.
- No broad visual redesign beyond making the existing application shell and workflows match the current product vocabulary and data model.
- No restoration of old task routes, compatibility aliases, or hierarchical task behavior.

## Design Notes

- Treat `docs/architecture/frontend-spa.md`, `docs/architecture/backend-api.md`, and `backend/internal/dto` as the implementation contract.
- Keep the first screen as the usable application shell, not a landing page.
- Preserve Quasar, Vue, Pinia, Vue Router, and the existing feature-oriented frontend structure.
- Prefer renaming and adapting existing frontend workflows over replacing them with a new UI architecture.
- Use backend-provided rendered body HTML for display and backend mutation responses for completion updates.
- Keep destructive actions behind persistent confirmation dialogs with `Cancel` and `OK` labels.

## Relevant Specs

- `docs/concepts.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/functionality/current-project-context.md`
- `docs/operations/verification.md`
- `backend/internal/dto`

## Verification

- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`
- `rg "\b[Tt]ask\b|task_" frontend/src frontend/ARCHITECTURE.md --glob '!frontend/src/css/github-markdown-scoped.css' --glob '!frontend/dist/**'`
- `rg "/api/v1/task|task_id|task_count|task_phase|task_type|parent_id|description_html" frontend/src frontend/ARCHITECTURE.md --glob '!frontend/src/css/github-markdown-scoped.css' --glob '!frontend/dist/**'`
- `find frontend/src -path '*task*' -o -path '*tasks*'`

## Review Focus

- Whether every active frontend API call matches the current backend API and DTO field names.
- Whether current project switching refreshes changes, epics, and requirements without stale task state.
- Whether requirement mutations update change and epic completeness from backend responses.
- Whether the old hierarchy model is fully removed from UI state, routing, and tests.
- Whether visible labels and test fixtures use current product vocabulary.

## Follow-Ups

- PR comment `IC_kwDOTA2Xls8AAAABHt4yog`: Already addressed in this branch; `sp_epic_to_history` writes `project_id` to `public.epic_history`, and `sp_epic_requirement_recalculate` uses its `_epic_id` input without a shadowing local declaration.
