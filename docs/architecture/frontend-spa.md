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

Change cards may display the backend-provided `ref_uuid`, `ref`, `slug`, `epic_name`, `agent_edit`, `open`, and completion fields. New Changes may have unassigned `ref` and `slug`; the frontend must render that state without deriving identity locally. The frontend must not derive, edit, or submit `ref_uuid`, `ref`, `slug`, or project reference counters.

Frontend Flow assignment, per-Change stage mode editing, Run controls, and claim reset controls are out of scope until a dedicated frontend Change adds them. The frontend must not derive, edit, or submit Flow snapshot fields or Run state fields locally.

Project-scoped Change board state uses the backend list item shape. Detail and edit routes must load the selected Change through `POST /api/v1/change/get` before rendering or submitting detail-only fields such as `idea`, `spec`, `spec_html`, `pr`, `pr_html`, `pr_url`, and rendered HTML.

Change create forms require only a title and idea. They may allow no selected types, no epic, and no spec or PR fields. When no types are selected, API clients send or accept an empty `change_types` array instead of manufacturing a placeholder type.

## Epic Management
The Epics route owns epic list, create, edit, and delete workflows. The list uses a Quasar markup table and relies on backend epic response data, including linked change counts, to disable unsafe deletes. Epics do not have a detail route.

## Detail View
The change detail view shows the opened change, linked test cases in backend response order, backend-provided `epic_name`, `idea`, `spec`, `spec_html`, `pr`, `pr_html`, `pr_url`, `agent_edit`, `open`, and sanitized markdown rendered from Change spec and PR fields. Empty `change_types`, empty artifact strings, empty rendered artifact strings, and unassigned `ref` or `slug` must not crash the view or display `null`. PR URLs render as links only when they are absolute `http` or `https` URLs; other stored values remain visible as plain text. Test case create, edit, done toggle, and delete actions update visible completeness and done state from backend responses.

The change detail view may render `ref_uuid`, `ref`, and `slug` as read-only identity data. Change create and edit forms must not expose inputs for `ref_uuid`, `ref`, `slug`, or project reference counters. Forms and API clients must use `idea`, `spec`, `pr`, and `pr_url` field names exactly.

Frontend artifact update calls send strings, not nulls. User-driven `idea`, `spec`, and `pr` saves send `agent_edit: false`; agent-produced saves send `agent_edit: true`. Empty `spec`, `pr`, and `pr_url` values render as empty values after reload when those optional fields have not been submitted yet; focused update payloads for those fields must be non-empty.

Detail API responses may include backend-owned Flow snapshot and Run state fields. The frontend may ignore those fields until user-facing Flow and Run controls are added, but it must not crash when the fields are present.

## Confirmations
Destructive operations use persistent confirmation dialogs. Buttons are consistently labeled `Cancel` and `OK`; dangerous `OK` actions use negative styling.

## Testing
Frontend verification uses:

- `vue-tsc` for types.
- ESLint for static checks.
- Vitest and Vue Test Utils for unit and component coverage.
- Browser-level tests only for workflows that need full routing and rendering.
