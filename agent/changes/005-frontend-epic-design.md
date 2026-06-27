# Frontend Epic Design

## Goal

Move epic management out of the Changes page into dedicated frontend pages and expose enough backend epic data for those pages to manage edit and delete behavior.

## Scope

- Add a top-navigation Epics entry that opens a dedicated epic list route.
- Replace inline epic create, rename, and delete controls on the Changes page with dedicated epic list, create, and edit routes.
- Add frontend routes for listing, creating, and editing epics.
- Update backend epic list and get responses to include the number of linked changes.
- Update database epic read views so backend list and get handlers return the linked change count.
- Keep the existing epic create, update, delete, and completeness behavior.

## Requirements

- The top navigation must include `Epics`, linking to `#/epics`.
- `#/epics` must render a Quasar `QMarkupTable` with no pager, page title, or description.
- `#/epics` must include a `New Epic` button that opens `#/epics/create`.
- Each epic table row must show edit and delete icon actions on the right side.
- The edit action must use a pencil icon and route to `#/epics/edit/[id]`.
- The delete action must use a trash icon, open the global `Are You Sure?` confirmation dialog, and delete the epic only when the user confirms.
- Epic delete must be disabled when the epic has linked changes.
- Epic list rows must use the data returned by the epic list endpoint; there is no separate epic detail page.
- `#/epics/create` must show title `Create Epic`, submit button `Create`, and redirect to `#/epics` after a successful create.
- `#/epics/edit/[id]` must show title `Edit Epic`, submit button `Save`, and redirect to `#/epics` after a successful save.
- Backend epic list and get responses must include `change_count` using the same field spelling as project responses.
- The database must provide `vw_epic.change_count`, and epic list and get queries must read from `vw_epic`.
- The local development and test databases must be recreated from `db/init.sql` so `vw_epic` exists consistently.

## Acceptance Criteria

- The Epics top-navigation item opens `#/epics`.
- The Changes page no longer contains the inline epic management table or inline epic create/edit/delete controls.
- `#/epics` lists epics in a Quasar `QMarkupTable` without pagination.
- Epic rows route to the edit page through the pencil icon.
- Epic rows disable the trash icon when `change_count > 0`.
- The trash icon opens the shared confirmation dialog and confirmed deletes remove the epic from the list.
- Create and edit forms use the specified route paths, titles, submit labels, and post-submit redirect.
- `POST /api/v1/epic/list` and `POST /api/v1/epic/get` return `change_count`.
- Backend epic repository reads list and get data from `public.vw_epic`.
- Frontend tests cover route navigation, disabled delete behavior, confirmed delete behavior, create redirect, and edit redirect.
- Backend tests cover epic `change_count` in list and get responses.

## Non-Goals

- No epic detail page.
- No pager, search, or filter controls on the epic list page.
- No redesign of project selection or current-project routing beyond what the Epics pages require.
- No changes to epic completeness calculations.
- No change to the global confirmation dialog labels or behavior.

## Design Notes

- Use `change_count`, not `changes_count`, to match the existing project summary field.
- Use `docs/architecture/frontend-spa.md` for frontend navigation and route expectations.
- Use `docs/architecture/backend-api.md` for epic endpoint response expectations.
- Keep route pages thin and place reusable epic behavior under frontend feature modules.
- Prefer direct component and route tests over browser-level tests unless routing cannot be verified at the unit/component layer.

## Relevant Specs

- `docs/architecture/frontend-spa.md`
- `docs/architecture/backend-api.md`
- `docs/product-overview.md`
- `docs/operations/verification.md`

## Verification

- `(cd backend && make test)`
- `(cd backend && make api-test)`
- `pnpm --dir frontend test`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend build`

## Review Focus

- Whether epic management is fully removed from the Changes page and available from the dedicated Epics routes.
- Whether delete disabling is based on backend `change_count`, not inferred frontend state.
- Whether `vw_epic` and backend DTO/query changes use `change_count` consistently.
- Whether create and edit routes handle direct entry and post-submit redirects correctly.
- Whether frontend tests cover the route and confirmation workflows without relying on brittle snapshots.

## Follow-Ups

- PR comment `IC_kwDOTA2Xls8AAAABHvX1iw` P1 was implemented: `db/backup/restore-test.sh` restores into `changes_test` again.
- PR comment `IC_kwDOTA2Xls8AAAABHvX1iw` P2 will be handled in the next PR through `agent/changes/006-fix-integration-tests.md`.
