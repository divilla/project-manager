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
