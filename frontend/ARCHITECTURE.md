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

Project selection is global app-shell state owned by `features/projects/model/projectSelection.store.ts`.
Use `currentProjectId` and `currentProject` for this shared context. Do not introduce route-local project selectors for project-scoped screens.

When the user changes the current project from the top menu, the app-shell flow is:

1. Set `isSwitchingProject` so the project selector is disabled and shows loading.
2. Redirect immediately to `/loading`.
3. Run `features/projects/services/projectScopeRefresh.ts` to reload shared project-scope data. Today that reloads projects and all tasks for `currentProjectId`; future project-scoped caches should be added there.
4. Redirect to the current topic index, for example `/tasks/123` and `/tasks/create/123` both return to `/tasks`.
5. Clear `isSwitchingProject` so the selector is enabled again.

Nested project-scoped routes must not keep stale entity context across a project switch. Add new topic routes to `src/router/projectChangeRedirect.ts` when they need the same topic-index behavior.

Direct route entry is handled separately from explicit selector changes. If a user opens a task URL whose task belongs to a different project than `currentProjectId`, the task page asks whether to switch from the selected project to the task's project. Accepting the switch stores a route-driven target path in `projectSelection.store.ts`, selects the task's project, runs the same `/loading` refresh flow, and returns to the original task URL instead of collapsing to `/tasks`. Declining the switch keeps the selected project and leaves the mismatched route.

Do not implement this as a loose global boolean. Keep the intended route in Pinia so the app shell can consume and clear it after the project refresh flow.

## Task Descriptions

Task descriptions are stored and edited as raw markdown. Backend task detail and mutation responses include sanitized `description_html` rendered through `backend/pkg/markdown`.

When a frontend screen needs rendered descriptions for known task IDs but does not need full task detail records, use `POST /api/v1/task/rendered-descriptions` with `{"ids":[...]}`. The response is scoped to rendered description fragments:

```json
{
  "descriptions": [
    { "id": 1, "description_html": "<p>...</p>" }
  ]
}
```

This keeps board/list payloads lean while still letting detail pages render ancestor markdown safely.

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
