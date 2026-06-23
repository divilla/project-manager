# Feature 06: Frontend Architecture

## 1. Purpose

Refactor the Projects frontend into a feature-first architecture that keeps Vue pages thin, moves product behavior into domain modules, and establishes conventions for future Quasar, Pinia, and TypeScript work.

The goal is not a visual redesign. The goal is to make the Projects area easier to change without turning a page component into the place where API calls, state coordination, form behavior, and domain rules all accumulate.

## 2. Scope

Allowed implementation areas:

- `frontend/src/pages/index/projects.vue`
- `frontend/src/services/api.ts`
- new files under `frontend/src/features/`
- new files under `frontend/src/shared/`
- Pinia stores under `frontend/src/stores/` or feature-local model folders
- frontend tests and type/lint configuration needed to enforce the architecture
- frontend architecture documentation

Out of scope:

- backend API behavior changes
- database changes
- visual redesign unrelated to component extraction
- introducing a new frontend framework, router, or state library
- replacing Quasar project structure

## 3. Architectural Direction

Keep the existing Quasar application structure and layer a feature-first organization inside `frontend/src`.

Use Quasar's native top-level concepts for application infrastructure:

- `pages/` for route entry components
- `layouts/` for layout components
- `router/` for route configuration
- `boot/` for application initialization
- `stores/` only for truly shared stores, if feature-local stores are not used

Use feature folders for product domains:

```text
frontend/src/
  features/
    projects/
      api/
      model/
      composables/
      components/

    tasks/
      api/
      model/
      composables/
      components/

    requirements/
      api/
      model/
      composables/
      components/

  shared/
    api/
    ui/
    lib/
    types/
```

## 4. Boundary Rules

Pages:

- Route pages are orchestration shells.
- Pages may compose feature components and call feature composables or stores.
- Pages must not contain raw backend request functions.
- Pages must not contain large domain mutation workflows when those workflows can live in a feature composable or store.

Features:

- A feature owns its domain API wrapper, types, store/composable behavior, and domain-specific components.
- Feature code may import from `shared/`.
- Feature code should avoid importing from another feature unless the dependency is an intentional product relationship.
- Cross-feature coordination should live in a page-level composable or a dedicated orchestration module.

Shared:

- `shared/` contains framework-neutral or app-wide utilities.
- `shared/` must not import from `features/` or `pages/`.
- `shared/api/httpClient.ts` owns common HTTP behavior.
- `shared/ui/` contains reusable non-domain UI components only.

Services:

- `frontend/src/services/api.ts` should not remain a monolithic API surface.
- Split backend calls into feature API modules.
- Keep request and response types close to the API module or feature model that owns them.

Pinia:

- Pinia is the default state tool for shared, durable client state.
- Local component state stays in Vue components or composables.
- Stores should coordinate state and expose actions, not become broad service classes.
- Stores may call feature API modules.
- Components should not call `fetch` directly.

## 5. Proposed Projects Area Shape

The Projects page currently coordinates projects, tasks, requirements, dialogs, board grouping, form state, and backend calls in one component. Refactor it toward:

```text
frontend/src/features/projects/
  api/projectApi.ts
  model/project.types.ts
  model/project.store.ts
  composables/useProjectSelection.ts
  components/ProjectList.vue
  components/ProjectCreateForm.vue
  components/ProjectRenameDialog.vue

frontend/src/features/tasks/
  api/taskApi.ts
  model/task.types.ts
  model/task.store.ts
  composables/useTaskBoard.ts
  components/TaskBoard.vue
  components/TaskCard.vue
  components/TaskCreateForm.vue
  components/TaskDetailDialog.vue

frontend/src/features/requirements/
  api/requirementApi.ts
  model/requirement.types.ts
  composables/useRequirementMutations.ts
  components/RequirementList.vue
  components/RequirementCreateForm.vue
  components/RequirementListItem.vue

frontend/src/shared/api/
  httpClient.ts
  apiError.ts
```

`projects.vue` should become a thin route-level composition of the project list, task board, and active dialogs.

## 6. Data Refresh Rules

Task requirement mutations can update ancestor task counters in the backend. Task phase changes can update ancestor phases in the backend.

Frontend behavior must preserve these invariants:

- After a requirement create, done toggle, definition update, task move, or delete, the task board must show current task aggregate values.
- If the backend response contains only the directly affected task, the frontend must refresh the selected project's task list after applying the immediate response.
- If a future backend response returns all affected task rows, the frontend may merge those rows instead of reloading the whole list.
- Open dialogs should remain responsive and should continue showing the mutation response for the directly edited task.

## 7. Confirmation Dialogs

Confirmation dialogs must use consistent Quasar button labels and styling across the application:

- Never use a button labeled `Confirm` in a confirmation dialog.
- Confirmation dialogs always use `Cancel` for the cancel action and `OK` for the accepting action.
- `Cancel` is always flat and has no explicit color.
- `OK` is always not flat.
- `OK` uses `color="negative"` for dangerous actions such as delete.
- `OK` uses `color="primary"` for non-dangerous confirmations.
- Confirmation dialogs that guard destructive actions must be persistent, so clicking the modal surrounding area does not close the dialog.
- Destructive delete confirmation UI is provided by `frontend/src/shared/ui/DeleteConfirmationDialog.vue` so any page can reuse the same title, persistence, labels, and button styling.

## 8. Migration Steps

1. Add `shared/api/httpClient.ts` and move the common `post` behavior out of `services/api.ts`.
2. Split DTO and API functions from `services/api.ts` into feature API and model files.
3. Introduce feature composables or Pinia stores for project list loading, task board loading, and requirement mutation flows.
4. Extract project list and project create/rename controls from `projects.vue`.
5. Extract task board, task cards, task create form, and task detail dialog.
6. Extract requirement list, requirement item editing, and requirement create form.
7. Reduce `projects.vue` to route composition and high-level error/loading presentation.
8. Remove or shrink `services/api.ts` after all callers use feature modules.

Each step should keep the UI functional and typecheckable.

## 9. Enforcement

Add lightweight enforcement after the first extraction is complete:

- TypeScript must pass with `npm run typecheck`.
- ESLint must pass with the existing frontend lint command.
- Add import-boundary lint rules if the current ESLint stack can support them without heavy tooling churn.
- At minimum, document import boundaries in `frontend/ARCHITECTURE.md`.

Suggested import boundary:

```text
shared -> no imports from features or pages
features -> may import shared
pages -> may import features and shared
```

## 10. Frontend Testing Strategy

Frontend tests should verify behavior at the right architectural level without duplicating Quasar, Vue, or browser behavior.

Use this test pyramid:

1. Type and lint checks for every frontend change.
2. Unit tests for pure feature logic and composables.
3. Component tests for reusable feature components with meaningful user interactions.
4. End-to-end tests for the Projects route's critical workflows.

### Required Tooling Direction

Prefer these tools:

- `vue-tsc` for type checks.
- ESLint for static code quality and import-boundary enforcement.
- Vitest for unit and component tests.
- Vue Test Utils for Vue component mounting.
- Playwright for end-to-end route workflows.
- MSW or a small fetch-mocking layer for frontend tests that exercise API clients without requiring a live backend.

Do not use a browser E2E test where a composable or component test would provide the same confidence with less setup and less flakiness.

### Unit Tests

Unit tests should cover:

- pure helpers in `shared/`
- feature model mappers, if added
- composables that coordinate state transitions
- error handling and refresh rules in route-level composables

For `useProjectsPage`, focused tests should verify:

- loading references and projects initializes default task type and phase
- selecting a project loads its tasks
- project create selects the new project
- requirement mutations apply the immediate mutation response and refresh the selected project task list
- task phase changes refresh the selected project task list because ancestor phases can change
- failures set a user-facing error message without clearing existing state unnecessarily

Mock feature API modules at this level. Do not make real HTTP calls in unit tests.

### Component Tests

Component tests should cover extracted feature components where props, events, and visible state matter:

- `ProjectCreateForm` emits create only through form submit and disables empty submit
- `ProjectList` emits select, rename, and delete with the expected project
- `TaskCreateForm` emits typed task form changes and create events
- `TaskBoard` renders tasks grouped by phase and emits open, move, and delete events
- `TaskDetailDialog` wires requirement create, toggle, edit, save, delete, and task save events
- `RequirementListItem` switches between display and edit states correctly

Keep component tests focused on public props/events and rendered behavior. Avoid asserting internal implementation details.

### API Module Tests

Feature API modules should be covered with lightweight tests around request shape and error handling when they contain non-trivial behavior.

If API modules remain thin wrappers over `shared/api/httpClient`, test `httpClient` behavior once instead of repeating the same test for every endpoint:

- sends JSON POST requests
- returns parsed JSON on success
- handles `204 No Content`
- throws backend-provided `error` or `message`
- throws a stable fallback error when the backend response has no message

### End-to-End Tests

Use Playwright for critical user workflows only:

- Projects page loads projects and task references
- user creates, renames, selects, and deletes a project
- user creates a task, moves it between phases, opens it, edits it, and deletes it
- user creates, toggles, edits, and deletes a requirement
- requirement and task phase mutations keep the visible board state current after backend recalculation

E2E tests should run against either:

- a controlled local test backend and database, or
- a deterministic mocked network layer

Do not point E2E tests at a developer's normal local database.

### Test Placement

Use colocated tests for feature code:

```text
frontend/src/features/projects/components/ProjectList.test.ts
frontend/src/features/projects/composables/useProjectsPage.test.ts
frontend/src/features/tasks/components/TaskBoard.test.ts
frontend/src/features/requirements/components/RequirementListItem.test.ts
frontend/src/shared/api/httpClient.test.ts
```

Use a top-level E2E folder for browser workflows:

```text
frontend/e2e/projects.spec.ts
```

### Test Data

Test data should be built by small factories rather than copied inline across tests:

```text
frontend/src/features/tasks/model/task.fixtures.ts
frontend/src/features/requirements/model/requirement.fixtures.ts
```

Factories should create complete valid objects and allow partial overrides.

### CI Expectations

Frontend CI should run, in order:

1. `npm run typecheck`
2. frontend ESLint check without auto-fix
3. Vitest unit and component tests
4. Playwright E2E tests for critical workflows, when the controlled backend or network mock is available

## 11. Acceptance Criteria

- `projects.vue` no longer owns raw API calls for projects, tasks, or requirements.
- Project, task, and requirement API calls are split into feature modules.
- Project, task, and requirement TypeScript types are owned by feature model files.
- Reusable project, task, and requirement UI pieces are extracted into feature components.
- Requirement and task phase mutations keep ancestor-derived task board state current.
- Frontend testing strategy is documented in this feature and reflected in `frontend/ARCHITECTURE.md` when tests are introduced.
- No backend or database changes are introduced.
- `npm run typecheck` passes.
- The Projects page remains functionally equivalent after the refactor.
