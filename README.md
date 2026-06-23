# project-manager

Simple, agentic, project management solution

## Project API Shape

Project read and mutation endpoints return rows from `public.vw_project`:

```json
{
  "id": 1,
  "name": "Example project",
  "created": "2026-06-23T00:00:00Z",
  "modified": "2026-06-23T00:00:00Z",
  "task_count": 3
}
```

The view owns project ordering and task counts. Clients should consume the returned order directly.

## Local Testing Servers

After app-affecting development work, leave both local servers running so the UI can be tested immediately:

- Backend: run `go run ./cmd/server` from `backend/`; expected API URL is `http://localhost:8080`.
- Frontend: run `pnpm dev` from `frontend/`; expected app URL is `http://localhost:8000`.
