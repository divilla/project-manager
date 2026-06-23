# Frontend Architecture

The Quasar app keeps the framework-native structure for pages, layouts, router, boot files, and app-level stores. Product code is organized by feature under `src/features`.

## Boundaries

```text
shared -> no imports from features or pages
features -> may import shared
pages -> may import features and shared
```

`src/shared` contains reusable app infrastructure such as HTTP helpers and generic UI/utilities. It must not know about project, task, or requirement domains.

Feature folders own their API modules, model types, composables, and domain components. Components should not call `fetch` directly.

Route pages should stay thin. They compose feature components and route-level composables, but do not own raw API calls or large domain mutation workflows.

## Projects Area

The Projects route uses `features/projects/composables/useProjectsPage.ts` as the page-level coordinator across projects, tasks, and requirements.

Backend mutations that can recalculate ancestor task values refresh the selected project's task list after applying the immediate mutation response, so the board stays current while the open dialog remains responsive.

## Testing

Frontend checks run in layers:

1. `npm run typecheck`
2. `npm run lint:check`
3. `npm run test`

Vitest is the unit and component test runner. Vue Test Utils mounts Vue components, and `happy-dom` provides the DOM environment.

Use colocated tests for feature code:

```text
src/features/projects/components/ProjectList.test.ts
src/features/projects/composables/useProjectsPage.test.ts
src/features/tasks/components/TaskBoard.test.ts
src/shared/api/httpClient.test.ts
```

Current coverage focuses on:

- `shared/api/httpClient.ts` request, no-content, and error handling behavior
- `useProjectsPage` loading, project creation, mutation refresh, and error behavior
- extracted feature component event contracts

Use factories for complete valid test data:

```text
src/features/tasks/model/task.fixtures.ts
src/features/requirements/model/requirement.fixtures.ts
```

Browser E2E tests should be added only when a controlled backend/database or deterministic network mock is available. They should live under `e2e/` and cover critical Projects route workflows rather than duplicating unit or component coverage.
