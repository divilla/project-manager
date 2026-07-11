# CLI Improve Changes and Agentic Workflow

## Goal

Make the `mch` Changes list and Change detail workflows clearer and more consistent by updating screen chrome, selectors, filter display, boolean icons, test case workflows, open-state toggling, phase colors, shortcuts, and command availability.

## Scope

- Update `mch` Change list, Change detail, Change phase selector, Change type selector, confirmation, prompt, shortcut, and related command behavior.
- Update the documented `mch` TUI contract to match the new Change workflow behavior.
- Update backend behavior only where needed to return Change detail test cases in deterministic ID order.
- Update CLI client behavior only where needed to create, update, delete, or toggle test cases and toggle Change `open` through existing backend APIs and refresh the visible screen.
- Add or update focused CLI, backend, and API coverage for the changed contracts.

## Requirements

- `mch` must rename the user-facing type selector screen from `SelectTypesDropDownScreen` to `SelectChangeTypesDropDownScreen`.
- `mch` must rename the user-facing phase selector screen from `SelectPhasesDropDownScreen` or `SelectPhaseDropDownScreen` to `SelectChangePhasesDropDownScreen`.
- The Change type selector must render a hint above the selector content: `press <space> to change`.
- In the Change type selector, selected types must render with a leading `-` and unselected types must render with a leading `+`.
- In the Change type selector, pressing `<space>` on the highlighted type must toggle only the pending in-selector selection state and must keep the selector open.
- In the Change type selector, pressing `<return>` must route back to `ChangeDetailsState`. If at least one type was toggled, `<return>` must persist the pending type set through `POST /api/v1/change/update-change-types`, then reload the Change through `POST /api/v1/change/get` before rendering details. If no types were toggled, `<return>` must return to details without calling a mutation endpoint.
- The Change type selector must preserve the existing selector rules for backend option order and sorted submitted type slug lists.
- The Change phase selector must render normal options with a leading `-`, such as `-backlog` and `-progress`.
- The application shell must render a one-line header with `Make a change v<version>` left-aligned and the active screen context right-aligned.
- The application shell must render screen shortcut hints at the left of the footer before status, current project, and color cells.
- Footer shortcut hints must use `</> command`, not `</> command menu`.
- `MainState` must render `MainScreen` as the right header context and include `</> command` and `<esc> cancel` in its footer shortcut hint.
- `ChangeDetailsScreen` must render the Change body row label as `Body` instead of `Requirement`.
- `ChangeDetailsScreen` must render the PR body row label as `PR` instead of `Pull Request`.
- `ChangeDetailsScreen` must render the `Agent Edit` boolean as green U+2714 when true and red U+2718 when false.
- `ChangeDetailsScreen` must render the `Open` boolean as U+2705 when true and U+274C when false.
- `ChangeDetailsScreen` must render the `Complete` value with Lip Gloss color `10`.
- Pressing `<space>` on the selected `Open` row in `ChangeDetailsScreen` must toggle the Change open state through `POST /api/v1/change/update-open`, then reload the Change through `POST /api/v1/change/get` and refresh `ChangeDetailsScreen`.
- `ChangeDetailsScreen` must render linked test cases between the `Body` row and the `PR` row.
- Test case rows in `ChangeDetailsScreen` must not use or render the label `Test Cases`.
- Each test case row must render the test case `done` value in the first column as a right-aligned U+2705 icon when true or U+274C icon when false.
- Each test case row must render `<test_case.scenario> (#<test_case.id>)` in the value column.
- `ChangeDetailsScreen` must render a divider between the `Body` row and linked test case rows when test cases are present.
- Test cases returned for Change detail must be ordered by `test_case.id`.
- Pressing `<space>` on a selected test case row in `ChangeDetailsScreen` must toggle `test_case.done` through `POST /api/v1/test-case/update-done`, immediately fetch the current test cases for the Change, and refresh the entire Change detail screen.
- `/new-testcase` must be the Change detail command for adding a test case; `/new-test-case` must not be shown in the command menu.
- Selecting `/new-testcase` or pressing `<ctrl+n>` on `ChangeDetailsScreen` must open a scenario prompt. Pressing `<return>` with a scenario must save through `POST /api/v1/test-case/create`, route back to `ChangeDetailsScreen`, and refresh the Change data.
- Pressing `<return>` on a selected test case row in `ChangeDetailsScreen` must open a scenario edit prompt. Pressing `<return>` must save through `POST /api/v1/test-case/update`, route back to `ChangeDetailsScreen`, and refresh the Change data.
- Pressing `<del>` on a selected test case row in `ChangeDetailsScreen` must show an `Are you sure?` `/yes` `/no` confirmation. Selecting `/yes` must delete through `POST /api/v1/test-case/delete`, route back to `ChangeDetailsScreen`, and refresh the Change data. Selecting `/no`, `<esc>`, or `<ctrl+c>` must cancel without a delete request.
- `ChangeDetailsScreen` footer hint must include `<ctrl+n> new testcase`, `<return> edit`, `<space> toggle`, `<del> delete`, and `</> command`.
- `/find-filter` on `ChangesListState` must reopen with the previously applied find-filter text ready to edit. After `/clear-filters`, reopening `/find-filter` must start with a blank prompt.
- Pressing `<ctrl+f>` on `ChangesListState` must behave the same as selecting `/find-filter`.
- `/phase-filter` and `/type-filter` on `ChangesListState` must reopen with the currently applied filter selected instead of defaulting to the first option.
- `ChangesListScreen` must render the app title/header on the first line with `ChangesListScreen` right-aligned, then render active filter settings on the second line, then render the table immediately with no blank line between filters and table.
- `ChangesListScreen` filter settings must render in this format with exactly three spaces between fields: `/filter-phase <value>   /filter-type <value>   /filter-epic <value>   /filter-find <value>`.
- `ChangesListScreen` filter settings must be right-aligned to the rendered table width. Filter labels must use the existing muted grey style, and only filter values must render pure white.
- `ChangesListScreen` footer hint must include `<ctrl+n> new change`, `<return> view`, and `</> command`.
- `/new-change` must be removed from `MainState` and must be available only from `ChangesListState`.
- Pressing `<ctrl+n>` on `ChangesListScreen` must behave the same as selecting `/new-change`.
- The phase color for `production` must be Lip Gloss color `12` on both `ChangesListScreen` and `ChangeDetailsScreen`.
- The phase color for `progress` must be Lip Gloss color `10` on both `ChangesListScreen` and `ChangeDetailsScreen`.
- All confirmation dialogs in `mch` must render `Are you sure?` instead of `Confirm`, and `<ctrl+c>` must behave the same as selecting `/no`.
- Prompt dialogs must show footer shortcut hints; edit scenario prompts must include `<return> save`, `<ctrl+c> delete prompt`, and `<esc> cancel`, while delete confirmations must include `<return> select` and `<esc> or <ctrl+c> cancel`.
- Backend APIs remain authoritative for Change and test case persistence; `mch` must not write application database tables directly.
- Backend and CLI failures during selector save, open toggle, test case create, test case update, test case delete, or test case done toggle must leave the user on a recoverable screen with a visible error and no hidden local success state.
- Documentation under `docs/` must be updated during implementation so `docs/architecture/mch.md` and related behavior docs match this Change.

## Acceptance Criteria

- `SelectChangeTypesDropDownScreen` and `SelectChangePhasesDropDownScreen` are the rendered user-facing screen titles for Change type and phase selectors.
- Opening the Change type selector shows `press <space> to change`, renders `+` and `-` prefixes correctly, allows multiple `<space>` toggles without leaving the selector, and saves exactly once on `<return>` only when the pending selection changed.
- Pressing `<return>` in the Change type selector with no pending toggles returns to `ChangeDetailsScreen` without calling `POST /api/v1/change/update-change-types`.
- Change phase selector options render with a leading `-`.
- The app header renders `Make a change v<version>` and the active screen context on one line; shortcut help renders in the footer on all screens.
- `ChangeDetailsScreen` row labels use `Body` and `PR`, never `Requirement` or `Pull Request`.
- `Agent Edit`, `Complete`, `Open`, and test case `done` values render with the required Unicode icons and colors.
- Pressing `<space>` on `Open` persists the inverted `open` value through `POST /api/v1/change/update-open`, reloads details through `POST /api/v1/change/get`, and shows the refreshed icon.
- Change detail test cases appear after a divider below `Body` and before `PR`, ordered by numeric ID, formatted as scenario plus `(#id)`, and without a `Test Cases` label.
- Pressing `<space>` on a test case row persists the inverted `done` value through `POST /api/v1/test-case/update-done`, refreshes current test case data, and rerenders the Change detail screen.
- `/new-testcase` and `<ctrl+n>` on `ChangeDetailsScreen` create a test case through the backend, return to details, and refresh data.
- Pressing `<return>` on a test case row updates the scenario through the backend, returns to details, and refreshes data.
- Pressing `<del>` on a test case row confirms with `Are you sure?`; `/yes` deletes through the backend, while `/no`, `<esc>`, and `<ctrl+c>` cancel without mutation.
- `/find-filter` preserves the previous filter text for editing until `/clear-filters` clears it.
- `<ctrl+f>` on `ChangesListScreen` opens the same find-filter prompt as `/find-filter`.
- `/phase-filter` and `/type-filter` reopen with the applied filter highlighted.
- `ChangesListScreen` shows the documented `/filter-*` settings line on the second line, right-aligned to the table width, with no blank line between filters and table, muted labels, pure-white values, and exactly three spaces between filters.
- `MainState` no longer lists or executes `/new-change`; `ChangesListState` still lists and executes `/new-change`.
- `<ctrl+n>` on `ChangesListScreen` opens Change create in the same way as `/new-change`.
- All confirmation dialogs use `Are you sure?`, and `<ctrl+c>` cancels confirmation in the same way as `/no`.
- `progress` and `production` phase values use the required colors consistently in both Change list and Change detail rendering.
- Updated docs describe the new command availability, labels, selector names, key behavior, filter display, boolean icons, test case rows, footer shortcut hints, confirmation behavior, and phase colors.
- Relevant CLI, backend service, and API integration tests pass.

## Non-Goals

- No new Change full edit, Change delete, epic selector, title edit, body edit, PR edit, or PR URL edit behavior beyond label, shortcut, and confirmation changes listed here.
- No database schema, migration, seed, fixture, or reference taxonomy changes.
- No direct database reads or writes from `mch`.
- No frontend UI changes.
- No non-interactive `mch` command tree or automation commands.
- No changes to `AGENTS.md`.

## Design Notes

- `docs/architecture/mch.md` is the primary source of truth for `mch` state names, commands, screen titles, selector behavior, rendering, refresh behavior, and CLI test strategy.
- `docs/architecture/backend-api.md` defines `POST /api/v1/change/update-open`, `POST /api/v1/change/update-change-types`, `POST /api/v1/test-case/update-done`, and the Change detail test case response contract.
- `docs/concepts.md` and `docs/functionality/requirements-and-acceptance.md` define test cases as the binary completion unit for Changes.
- This Change intentionally updates existing documentation that currently says `/new-change` exists on `MainState`, progress is bright cyan, production is color `10`, and detail rows use `Requirement` and `Pull Request`.
- The old draft used `Requirement` as a visible row label, but current product vocabulary uses Change `body`; the required visible label is `Body` while the underlying API field remains `body`.
- The old draft used `Pull Request` as a visible row label, but the required visible label is `PR` while the underlying API field remains `pr`.
- The `Agent Edit` icon pair is U+2714 and U+2718 with green and red styling. The `Open` and test case done icon pair is U+2705 and U+274C.
- The type selector pending state must be local to the open selector. Canceling the selector or returning with no toggles must not call a mutation endpoint.
- The test case refresh after toggling `done` may use a dedicated test case list endpoint or the recalculated data returned by the mutation endpoint, but the visible detail screen must reflect backend data and not a local-only guess.
- Test case create, scenario update, delete, and done-toggle flows must all route back to `ChangeDetailsScreen` and display backend-refreshed data rather than optimistic local-only state.
- The final Changes list filter display intentionally uses `/filter-phase`, `/filter-type`, `/filter-epic`, and `/filter-find` labels without colons; command names remain `/phase-filter`, `/type-filter`, `/epic-filter`, and `/find-filter`.
- `ChangesListScreen` keeps the app header on the first line and renders filters on the second line; there is intentionally no blank line between filters and the table.
- Shortcut hints are footer chrome rather than in-body help text. Long footer text may wrap with terminal width, but the ordering is shortcut hint, status when present, current project, then color cells.
- Use existing backend endpoints and DTO names; do not introduce old names such as `requirement_body`, `pull_request_body`, or `closed` in new contracts.

## Relevant Specs

- `agent/changes/110-cli-improve-changes-and-agentic-workflow.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/concepts.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/operations/verification.md`

## Verification

- From `backend`: `make lint`
- From `backend`: `make test`
- From `backend`: `make api-test`
- From `cli`: `make lint`
- From `cli`: `go test ./...`
- From `cli`: `go build -o /tmp/mch ./cmd/mch`
- Inspect updated docs to confirm every changed `mch` behavior in this Change is documented and every edited doc remains at or below 300 lines.

## QA Test Cases

- Open `mch`, confirm `MainState` does not show `/new-change`, then open `/changes` and confirm `ChangesListState` still shows `/new-change`.
- Confirm `MainState`, `ChangesListState`, and `ChangeDetailsState` show `Make a change v<version>` on the first line and render shortcut hints in the footer.
- Open a Change detail, open the Change type selector, confirm the title is `SelectChangeTypesDropDownScreen`, the hint says `press <space> to change`, selected types show `-`, unselected types show `+`, `<space>` toggles without leaving, and `<return>` saves the pending set and returns to refreshed details.
- Open the Change type selector and press `<return>` without pressing `<space>`; confirm the selector returns to details without a type update request.
- Open the Change type selector, press `<space>` to toggle, then cancel; confirm no type update request is made and details remain unchanged.
- Open the Change phase selector and confirm the title is `SelectChangePhasesDropDownScreen`, options render with leading `-`, and the currently selected phase behavior remains correct.
- Open a Change detail with body, PR body, `agent_edit`, `open`, and test cases; confirm row labels and row ordering are `Body`, test case rows, `PR`, `PR URL`, `Agent Edit`, `Complete`, `Open`.
- Confirm `Requirement`, `Pull Request`, and `Test Cases` do not appear on `ChangeDetailsScreen`.
- Confirm true and false `Agent Edit` values render as green U+2714 and red U+2718.
- Confirm true and false `Open` values render as U+2705 and U+274C.
- Confirm `Complete` renders with Lip Gloss color `10`.
- Confirm a divider appears between `Body` and the first linked test case row when test cases exist.
- Select the `Open` row and press `<space>`; confirm `POST /api/v1/change/update-open` receives the inverted boolean, the Change is reloaded, and the refreshed screen shows the new icon.
- Simulate an `update-open` backend failure; confirm the user stays on a recoverable Change detail screen with a visible error and no local-only icon change.
- Open a Change detail with multiple test cases returned out of order by ID at the repository layer; confirm the rendered order is ascending by `test_case.id`.
- Select an incomplete test case row and press `<space>`; confirm `POST /api/v1/test-case/update-done` receives the inverted boolean, current test case data is refreshed, and the screen rerenders with updated completion state.
- Simulate a test case done-toggle backend failure; confirm the user sees a recoverable error and the row is not shown as successfully toggled from local-only state.
- On `ChangeDetailsScreen`, select `/new-testcase` and enter a scenario; confirm `POST /api/v1/test-case/create` saves it, the screen returns to refreshed details, and the new scenario appears.
- On `ChangeDetailsScreen`, press `<ctrl+n>`; confirm it opens the same test case scenario prompt as `/new-testcase`.
- Select a test case row and press `<return>`; edit the scenario, save, and confirm `POST /api/v1/test-case/update` persists it and refreshed details show the updated scenario.
- Select a test case row and press `<del>`; confirm `Are you sure?` appears, `/no`, `<esc>`, and `<ctrl+c>` cancel without mutation, and `/yes` deletes through `POST /api/v1/test-case/delete` and refreshes details.
- Apply `/find-filter`, reopen `/find-filter`, and confirm the previous search text is available to edit.
- Press `<ctrl+f>` on `ChangesListScreen` and confirm it opens the same prompt and behavior as `/find-filter`.
- Run `/clear-filters`, reopen `/find-filter`, and confirm the prompt is blank.
- Apply `/phase-filter`, reopen it, and confirm the applied phase is highlighted instead of the first option.
- Apply `/type-filter`, reopen it, and confirm the applied type is highlighted instead of the first option.
- On `ChangesListScreen`, confirm the first line renders the app title and `ChangesListScreen`, the second line renders `/filter-phase <value>   /filter-type <value>   /filter-epic <value>   /filter-find <value>`, and the table begins on the next line with no blank line.
- Confirm the Changes list filter row is right-aligned to the table width, labels are muted grey, values are pure white, and fields are separated by exactly three spaces.
- Confirm blank or unset filters render consistently in the filter settings line without breaking the list layout.
- On `ChangesListScreen`, press `<ctrl+n>` and confirm it opens the same Change create flow as `/new-change`.
- Open any delete confirmation and confirm the label is `Are you sure?` and `<ctrl+c>` behaves the same as selecting `/no`.
- Confirm `progress` uses color `10` and `production` uses color `12` in both the Change list and Change detail views, including selected-row rendering.
- Run the CLI with a backend unavailable during selector option load, type save, open toggle, test case create, test case update, test case delete, and test case done toggle; confirm each path reports a recoverable error and preserves usable navigation.

## Review Focus

- Selector state handling for pending Change type toggles, especially no-op return, cancel, backend failure, and single-save behavior.
- Backend ordering and DTO compatibility for Change detail test cases.
- Correct use of `update-open`, `update-change-types`, `update-done`, and refresh calls without direct database writes.
- Correct use of test case create, update, delete, done-toggle, and refresh calls without direct database writes.
- `ChangeDetailsScreen` row ordering, dividers, labels, icon rendering, color styling, and `<space>`, `<return>`, and `<del>` key routing.
- Filter state preservation, highlighting, shortcut routing, and display formatting on `ChangesListScreen`.
- Documentation consistency for changed `mch` command availability, selector names, phase colors, body/PR labels, boolean icons, test case rows, footer hints, confirmation behavior, and filter chrome.

## Follow-Ups

- None.
