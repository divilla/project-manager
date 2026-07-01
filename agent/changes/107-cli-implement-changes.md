# Implement Backend-Backed Change Screens In `mch`

## Goal

Make the `mch` Changes list, detail, create, update, and filter flows operate on backend Change data for the current project instead of dummy navigation data.

## Scope

- Implement `ChangesListScreen` with backend-loaded Changes for the current project.
- Implement `ChangeDetailsScreen` with backend-loaded read-only Change identity, metadata, and requirement content.
- Implement `ChangeCreateScreen` to create a Change from editor-provided requirement markdown through the backend API.
- Implement `ChangeUpdateScreen` to edit an existing Change from editor-provided requirement markdown through focused backend update APIs.
- Implement `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, and `/clear-filters` on `ChangesListState`.
- Preserve existing Changes navigation commands, return targets, command ordering, help routing, and delete confirmation routing.
- Add focused CLI tests for successful load, create, update, filtering, clearing filters, backend failures, validation failures, and cancel behavior.
- If needed change backend files

## Requirements

- `ChangesListState` must load Changes with `POST /api/v1/change/list` using the saved current project ID as a numeric `project_id`.
- `ChangesListState` must render a boxed, scrollable selectable Changes table with columns `#Ref`, `Phase`, `Types`, `Epic`, `Title`, `Don`, `Tot`, `%`, and `Modified`, in that order.
- Change list `Types` values must be at most 30 characters wide, `Epic` values must be at most 20 characters wide, and `Title` values must be at most 80 characters wide; longer values must be truncated at that position without adding a suffix.
- The Change list table must render at its natural column width when the terminal is wide enough, rather than expanding to full screen width; columns must shrink only when the available terminal width is smaller than the natural table width.
- Numeric Change refs in list and detail views must render as six digits with leading zeroes and no `#` prefix, for example `000003`.
- Change list `Modified` values must render as `YYYY-MM-DD HH.MM`; missing or invalid timestamps must render as `not a date`.
- Up and down arrows in `ChangesListState` must move the selected row within bounds and keep it visible inside the table viewport.
- PgUp and PgDown in `ChangesListState` must move the selected row by one visible table page and keep it within bounds.
- Enter or Return in `ChangesListState` must select the highlighted Change.
- Selecting a Change from `ChangesListState` must open `ChangeDetailsState` and load the selected Change with `POST /api/v1/change/get`.
- `ChangeDetailsState` must render `ChangeDetailsScreen - Title: Change Details` with backend-provided `ref` and `slug` as read-only identity data.
- `ChangeDetailsState` must render the Change title, phase, types, epic, closed state, and requirement body from the backend response.
- `ChangeCreateState` must render `ChangeCreateScreen - Title: New Change` and expose `/save` and `/cancel`.
- Opening `ChangeCreateState` must open the external editor for requirement markdown entry.
- The only editable Change create input must be the requirement markdown body.
- Saving a new Change must parse the title, change types, optional epic, and persisted `requirement_body` from the editor body before calling the backend.
- The editor body first non-blank line must be an H1 title, and the extracted title must be non-empty.
- The first non-blank line after the H1 title must be formatted exactly as `Types: <type-slugs>`.
- `<type-slugs>` must contain one or more backend type slugs joined by `|` with no spaces.
- Missing title, missing `Types:`, blank `Types:`, invalid type slug, or malformed type formatting must show a recoverable validation error and must not call the backend.
- The next non-blank line after `Types:` may be formatted as `Epic: <epic-name>`.
- A missing `Epic:` line or blank `Epic:` value must be treated as no epic.
- A non-blank `Epic:` value must resolve to an epic from `POST /api/v1/epic/list` for the current project before save.
- The full editor markdown, including H1, `Types:`, optional `Epic:`, and all following sections, must be preserved as the backend `requirement_body` on every create or update save.
- Change create must send `POST /api/v1/change/create` with `project_id`, `title`, `requirement_body`, `change_types`, and optional `epic_id`.
- Change create must not send `ref`, `slug`, `change_phase`, `pull_request_body`, or `pull_request_url`.
- Successful Change create must open `ChangeDetailsState` for the created Change.
- `ChangeUpdateState` must render `ChangeUpdateScreen - Title: Edit Change` and expose `/save` and `/cancel`.
- `/edit` from `ChangeDetailsState` must open `ChangeUpdateState` and open the external editor with the selected Change represented in the same requirement markdown format.
- The only editable Change update input must be the requirement markdown body.
- Saving an edited Change must parse and validate title, change types, optional epic, and persisted `requirement_body` using the same rules as Change create before calling the backend.
- Change update must persist title changes through `POST /api/v1/change/update-title`.
- Change update must persist requirement body changes through `POST /api/v1/change/update-requirement-body`.
- Change update must persist change type changes through `POST /api/v1/change/update-change-types`.
- Change update must persist epic changes through `POST /api/v1/change/update-epic`, using `null` when the editor body omits `Epic:` or has a blank `Epic:` value.
- Successful Change update must reload the Change with `POST /api/v1/change/get` before refreshing `ChangeDetailsState`.
- `/cancel` and `Esc` from Change create must return to `ChangesListState` without a create request.
- `/cancel` and `Esc` from Change update must return to `ChangeDetailsState` without an update request.
- Failed list, get, create, or update requests must keep the user in a recoverable state and display the backend error.
- `ChangesListState` must expose exactly `/new-change`, `/phase-filter`, `/epic-filter`, `/type-filter`, `/find-filter`, `/clear-filters`, `/help`, and `/return` in that order.
- `/phase-filter` must load phase options from `POST /api/v1/change/reference`, apply the selected phase filter in `ChangesListState`, and keep the user on `ChangesListState`.
- `/type-filter` must load type options from `POST /api/v1/change/reference`, apply the selected type filter in `ChangesListState`, and keep the user on `ChangesListState`.
- `/epic-filter` must load epic options from `POST /api/v1/epic/list` using the current project ID as a numeric JSON value, apply the selected epic filter in `ChangesListState`, and keep the user on `ChangesListState`.
- Phase, type, and epic filter option lists must append `/clear` as the final option to clear only that filter field.
- `/find-filter` must apply a text filter to visible Changes by matching title, `ref`, `slug`, phase, type, epic, or requirement body text already loaded for the list.
- `/clear-filters` must clear phase, type, epic, and find filters without leaving `ChangesListState`.
- Change screens must not prompt for, derive, submit, or edit backend-owned `ref`, `slug`, or project reference counters.
- Tests must use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`.

## Acceptance Criteria

- Opening `/changes` from `MainState` calls `POST /api/v1/change/list` with the current numeric project ID and renders `ChangesListScreen - Title: Changes List`.
- A Change row displays the backend-provided `ref` as a six-digit number, phase, types, epic label, title, done count as `Don`, total count as `Tot`, completed percentage as `%`, and modified timestamp inside a boxed, scrollable table.
- Up/down arrows move the selected Change row without moving past the first or last visible Change, and scrolling keeps the selected row visible.
- PgUp and PgDown move the selected Change row by one visible table page without moving past the first or last visible Change.
- Pressing Enter or Return on a selected Change calls `POST /api/v1/change/get` and renders `ChangeDetailsScreen - Title: Change Details` with the fetched Change data.
- Change details display `ref` and `slug` as read-only values and do not expose editable inputs for either field.
- `/new-change` opens `ChangeCreateScreen - Title: New Change` and opens the external editor for requirement markdown entry.
- Saving a valid new Change extracts the H1 title, `Types:` slugs, optional `Epic:` value, and persisted `requirement_body`, sends `POST /api/v1/change/create` with `project_id`, `title`, `requirement_body`, `change_types`, and optional `epic_id`, then opens details for the created Change.
- Saving a new Change with a missing H1 title, missing `Types:`, blank `Types:`, invalid type slug, malformed type line, or unresolved non-blank `Epic:` does not call the backend and shows a recoverable validation error.
- Create cancel through `/cancel` or `Esc` returns to `ChangesListState` without calling `POST /api/v1/change/create`.
- `/edit` from details opens `ChangeUpdateScreen - Title: Edit Change` and opens the external editor with the current Change represented as requirement markdown.
- Saving an edited Change extracts and validates title, `Types:`, optional `Epic:`, and `requirement_body`, then calls the focused update endpoints for changed title, requirement body, change types, and epic.
- After a successful edit save, `mch` reloads the Change through `POST /api/v1/change/get` and renders the refreshed details.
- Update cancel through `/cancel` or `Esc` returns to `ChangeDetailsState` without calling update endpoints.
- Backend failures during list, get, create, or update remain visible and leave enough user input or selection state to retry or cancel.
- `/phase-filter`, `/type-filter`, and `/epic-filter` open filter option overlays on `ChangesListState`, apply the selected filter, and keep the list title visible.
- Selecting `/clear` from a phase, type, or epic filter clears only that filter field.
- `/find-filter` filters visible Changes by the entered text and shows a no-results state when no Change matches.
- `/clear-filters` clears all active list filters and restores the unfiltered loaded list.
- `cd cli && make lint`, `cd cli && go test ./...`, and `cd cli && go build -o /tmp/mch ./cmd/mch` pass.

## Non-Goals

- No direct database writes from `mch`.
- No Change delete persistence changes.
- No Test Case create, update, or detail behavior changes beyond preserving existing navigation from Change details.
- No pull request body or pull request URL editing in `mch`.
- No frontend SPA Changes changes.
- No changes to current project selection or `.config/config.yaml` persistence.

## Design Notes

- Change create and update use the strict requirement markdown contract from `agent/prompts/build-requirement-with-agent.md`: H1 title is mandatory, `Types:` is mandatory, and `Epic:` is optional.
- Epic can be omitted or blank. Title and types cannot be omitted or blank.
- Metadata extraction must not strip metadata lines from `requirement_body`; the full editor markdown is the saved body.
- Assumption: list filtering can operate on the currently loaded Change list; the backend does not need new filtered list endpoints for this Change.
- `docs/architecture/mch.md` defines the authoritative `mch` state names, screen titles, command ordering, filter behavior, package boundaries, and test strategy.
- `docs/architecture/backend-api.md` defines the Change, Epic, and reference endpoints and the payload fields that clients may send.
- `docs/concepts.md` defines `ref` and `slug` as backend-owned Change identity fields that clients cannot set or edit.
- Backend APIs remain authoritative for Change validation and persistence; `mch` should surface backend errors rather than bypassing them.
- Change create and update must open the external editor because requirement markdown is the only editable input for these flows.

## Relevant Specs

- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/concepts.md`

## Verification

- From the repository root: `cd cli && make lint`
- From the repository root: `cd cli && go test ./...`
- From the repository root: `cd cli && go build -o /tmp/mch ./cmd/mch`

## QA Test Cases

- Open `/changes` with a valid current project and confirm the list renders backend Changes for that project.
- Open `/changes` when the backend list request fails and confirm a recoverable error is visible.
- Select a Change from the list and confirm details render the fetched Change, including read-only `ref` and `slug`.
- Navigate the Change list with up/down arrows and PgUp/PgDown on a short terminal and confirm the selected row scrolls into view.
- Create a Change with a valid H1 title, `Types:` line, requirement body, and matching epic, then confirm the created Change details open.
- Create a Change with a valid H1 title and `Types:` line but omitted `Epic:` line, then confirm the create request omits `epic_id`.
- Create a Change with a valid H1 title and `Types:` line but blank `Epic:` line, then confirm the create request omits `epic_id`.
- Try to create a Change with a missing H1 title and confirm no backend create request is sent.
- Try to create a Change with a missing, blank, malformed, or invalid `Types:` line and confirm no backend create request is sent.
- Try to create a Change with a non-blank unresolved `Epic:` value and confirm no backend create request is sent.
- Cancel Change create with `/cancel` and with `Esc`, and confirm no create request is sent.
- Edit a Change requirement markdown through the external editor, save, and confirm the focused update endpoints are called before details reload.
- Save an edit where only the title changed and confirm the requirement body endpoint is not called.
- Save an edit where only the requirement body changed and confirm the title endpoint is not called.
- Save an edit where only `Types:` changed and confirm only the change types endpoint is called before reload.
- Save an edit where `Epic:` is omitted or blank and confirm the epic endpoint is called with `null` when the previous Change had an epic.
- Try to save an edit with missing H1 title or invalid `Types:` and confirm no update endpoint is called.
- Cancel Change edit with `/cancel` and with `Esc`, and confirm no update request is sent.
- Apply a phase filter and confirm the visible list only shows matching Changes.
- Apply a type filter and confirm the visible list only shows matching Changes.
- Apply an epic filter and confirm the epic request sends numeric `project_id`.
- Select `/clear` from one filter overlay and confirm only that filter field is cleared.
- Apply `/find-filter` and confirm matching Changes remain visible while non-matching Changes are hidden.
- Use `/clear-filters` and confirm all filters are cleared and the loaded list is restored.

## Review Focus

- Verify Change create and update parse the strict editor body contract correctly and never submit backend-owned `ref`, `slug`, or project reference counters.
- Check that list and detail loads use the current project context correctly and send numeric IDs where required.
- Inspect filter state transitions so filter overlays remain inside `ChangesListState` and do not create separate filter screens.
- Confirm create, update, cancel, backend failure, and validation failure paths leave the user in recoverable states.
- Verify tests cover API calls, command ordering, screen titles, filter clearing, no-results behavior, and read-only identity fields.

## Follow-Ups

- Review fixes applied: preserved long Change editor markdown, matched find filters against displayed refs, validated Change markdown before reference lookups, cleared omitted epics, and clamped filtered Change selection.
- Add Change delete persistence through `POST /api/v1/change/delete`.
- Add `mch` editing for pull request body and pull request URL fields.
