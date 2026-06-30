# System Architecture

## Topology
The application is a local-first web system:

```text
Vue / Quasar SPA  <---- JSON HTTP ---->  Go / Echo API  <---- SQL ----> PostgreSQL
                                      |
                                      +---- optional LLM provider
```

The frontend and backend run as separate local development processes. The database is authoritative for application state.

## Backend
The backend is a Go service using Echo. It owns:

- JSON route handling.
- DTO validation and normalization.
- Repository access through SQL.
- Markdown rendering and sanitization.
- LLM proxy behavior.
- Centralized JSON error handling.
- Structured logging with zerolog.

Mutating resource actions use POST endpoints for prototype consistency. Health endpoints remain GET diagnostics.

## Frontend
The frontend is a Vue 3 and Quasar single-page app. It owns:

- Application shell and navigation.
- Current project selection.
- Project CRUD.
- Change board and detail views.
- Test case editing flows.
- Planning and help screens.
- Client-side loading, empty, error, and confirmation states.

Feature folders own domain API wrappers, types, stores, composables, and components.

## Database
PostgreSQL stores projects, epics, changes, test cases, reference options, and history rows. Application code must use the database contract as supplied. Backend and frontend implementation changes must not invent new schema behavior.

## AI Integration
The backend mediates LLM calls. Prompts receive project context and database-provided reference options. Model output is parsed, validated, and shown to the user before persistence.
