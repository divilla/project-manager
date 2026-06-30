# State-Based Navigation For `mch`

Types: feature|test

## Problem Statement

`mch` needs a complete state-based terminal navigation shell so users can move through Main, Changes, Requirements, Epics, Projects, selectors, confirmations, help, and quit behavior. Every supplied state and transition must exist, be testable, and render a deterministic dummy screen title. Save and delete actions are navigation-only in this requirement; they do not persist domain data.

The implementation must also create the local `cli/.config/config.yaml` file with a valid YAML key-value pair so the `cli` app has an initial config file in place.

## Primary Workflows

1. A user starts `mch` in `MainState`, rendered as `MainScreen - Title: Main`.
2. From Main, the user can enter `/changes`, `/epics`, `/projects`, `/select-project`, `/help`, or `/quit`.
3. Slash commands open the shared command dropdown; users can type to filter, use up/down arrows, and confirm the highlighted selection.
4. List item selection uses the same dropdown interaction model as slash commands.
5. `Esc` performs the state-appropriate safe action: `/quit` from `MainState`, `/return` from returnable states, and `/cancel` from cancelable states.
6. Selector dropdowns load real backend options when opened.
7. Confirmation dropdowns require `/yes` or `/cancel`.
8. `/quit` is accepted only from `MainState`.

## Acceptance Criteria

1. Implementation creates `cli/.config/config.yaml`.
2. `config.yaml` contains:
   ```yaml
   backend_url: http://localhost:8080
   ```
3. Starting `mch` initializes `MainState` and renders `MainScreen - Title: Main`.
4. `/changes` from `MainState` transitions to `ChangesListState`, rendering `ChangesListScreen - Title: Changes List`.
5. Selecting a Change transitions to `ChangeDetailsState`, rendering `ChangeDetailsScreen - Title: Change Details`.
6. Selecting a Requirement transitions to `RequirementDetailsState`, rendering `RequirementDetailsScreen - Title: Requirement Details`.
7. Requirement edit, create, delete, save, cancel, and return transitions follow the supplied navigation map.
8. Change edit, create, delete, save, cancel, phase, epic, types, and return transitions follow the supplied navigation map.
9. `ChangeDetailsState /phase` opens `SelectPhaseDropDown`, loads phases from `POST /api/v1/change/reference`, and returns to `ChangeDetailsState`.
10. `ChangeDetailsState /epic` opens `SelectEpicDropDown`, loads epics from `POST /api/v1/epic/list` with the current `project_id`, and returns to `ChangeDetailsState`.
11. `ChangeDetailsState /types` opens `SelectTypesDropDown`, loads type slugs from the `types` group returned by `POST /api/v1/change/reference`, and returns to `ChangeDetailsState`.
12. Changes filters support `/phase-filter`, `/type-filter`, `/find-filter`, and `/clear-filters`.
13. Changes help supports `/find`, highlights matching text, and `/return`.
14. `/epics` from `MainState` opens `EpicsListState`, rendering `EpicsListScreen - Title: Epics List`.
15. Epic list, detail, create, update, delete, help, find, and return transitions follow the supplied navigation map.
16. `/projects` from `MainState` opens `ProjectsListState`, rendering `ProjectsListScreen - Title: Projects List`.
17. Project list, detail, create, update, delete, help, find, and return transitions follow the supplied navigation map.
18. `/select-project` opens `SelectProjectDropDown`, loads projects from `POST /api/v1/project/list`, stores the selected current project in TUI state, and returns to `MainState`.
19. `/help` from `MainState` opens `MainHelpState`; `/find` highlights matches; `/return` returns to `MainState`.
20. `/quit` from `MainState` exits to the terminal.
21. `Esc` from `MainState` behaves exactly like `/quit`.
22. `Esc` from returnable states behaves exactly like `/return`.
23. `Esc` from create, update, dropdown, confirmation, and input states behaves exactly like `/cancel`.
24. `/quit` entered outside `MainState` leaves the current state unchanged and shows a recoverable error.
25. Unknown commands leave the current state unchanged and show a recoverable error.
26. Every state transition has a focused model test.
27. Rendering tests verify each dummy screen title exactly.
28. No save, delete, filter, selector, or selection transition writes directly to the database.

## Edge Cases

1. `cli/.config/` does not exist before implementation.
2. `config.yaml` exists but is malformed YAML.
3. `backend_url` is missing or empty.
4. Backend reference data cannot be loaded for phase, type, epic, or project selectors.
5. No current project exists when opening `SelectEpicDropDown`.
6. A dropdown has no options.
7. A user types a filter that matches no dropdown options.
8. A user presses `Esc` while a backend selector request is loading.
9. A user submits empty text in `FindInput`.
10. A help search has no matches.
11. A delete confirmation receives a command other than `/yes` or `/cancel`.
12. Terminal dimensions are too small to render the full dummy screen.
13. The supplied `ChangesDetailsState` spelling is treated as `ChangeDetailsState`.

## Non-Goals

1. No real Change, Requirement, Epic, or Project create/update/delete persistence is required.
2. No backend API contract changes are required.
3. No database migration is required.
4. No full form implementation beyond navigation controls is required.
5. No Codex-assisted planning flow is required.
6. No frontend SPA changes are required.
7. No production packaging or installer changes are required.

## Dependencies And Risks

1. Depends on the existing `cli/` Bubble Tea, Bubbles, and Lip Gloss architecture.
2. Depends on `POST /api/v1/change/reference` for phase and type selector options.
3. Depends on `POST /api/v1/project/list` for project selection.
4. Depends on `POST /api/v1/epic/list` with current `project_id` for epic selection.
5. Selector loading failures must be visible and recoverable without losing the previous state.
6. Reusing one dropdown model for commands, list selection, selectors, and confirmations requires clear caller context so confirmation and cancel routes remain correct.
7. The config file introduces a local filesystem dependency under `cli/.config/`.

## Open Questions

None.
