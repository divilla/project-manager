# Frontend SPA

## Application Shell
The frontend uses Vue 3, Quasar, Vite, Pinia, and TypeScript. The first screen is the usable application shell, not a landing page.

Top navigation includes:

- Home
- Planning
- Projects
- Epics
- Changes
- Help

The right side of the top bar contains the current project selector.

## Feature Structure
Product code is organized by feature:

```text
frontend/src/
  features/
    projects/
    epics/
    changes/
    test-cases/
  shared/
    api/
    ui/
    lib/
```

Route pages compose features and keep orchestration thin. Shared code must not depend on feature or page modules.

## State
Pinia stores durable client state such as current project selection and project-scoped cached data. Local form state stays in components or composables.

## Change Board
The Changes route shows change cards grouped by workflow phase. Search filters by title, type, and phase. Creating a new change opens a dedicated create route. Detail and edit routes must work from pasted URLs.

Change cards may display the backend-provided `ref` and must carry the backend-provided `slug` when present. The frontend must not derive, edit, or submit `ref`, `slug`, or project reference counters.

## Epic Management
The Epics route owns epic list, create, edit, and delete workflows. The list uses a Quasar markup table and relies on backend epic response data, including linked change counts, to disable unsafe deletes. Epics do not have a detail route.

## Detail View
The change detail view shows the opened change, linked test cases, and sanitized markdown requirement body. Test case create, edit, done toggle, and delete actions update visible completeness from backend responses.

The change detail view may render `ref` and `slug` as read-only identity data. Change create and edit forms must not expose inputs for `ref`, `slug`, or project reference counters.

## Confirmations
Destructive operations use persistent confirmation dialogs. Buttons are consistently labeled `Cancel` and `OK`; dangerous `OK` actions use negative styling.

## Testing
Frontend verification uses:

- `vue-tsc` for types.
- ESLint for static checks.
- Vitest and Vue Test Utils for unit and component coverage.
- Browser-level tests only for workflows that need full routing and rendering.
