# Spec 04: API Design Concepts

The Go/Echo backend exposes a stateless RESTful JSON API. For the prototype phase, the endpoints are designed to support rapid frontend-backend data transfer without authentication. 

This document serves as a high-level conceptual specification outlining the active resources, endpoints, and behavioral intents of each API node. Detailed JSON request/response payloads must be mapped to the existing database schema rather than a new proposed schema.

---

## 1. Base URL & Common Guidelines
- **Base Path:** `/api`
- **Port:** Configured by default to `:8080` (or as set via `PORT` environment variable).
- **Format:** All requests containing bodies should expect `application/json` format.
- **CORS:** Cross-Origin Resource Sharing (CORS) is enabled globally for local development origins, including `http://localhost:8000` for the Vite/Quasar frontend.
- **Health Check Exception:** Health diagnostics are intentionally GET endpoints. Keep `GET /api/v1/health` and `GET /api/health` as GET-only checks; do not change them to POST when implementing resource APIs.

---

## 2. Resource Endpoints

### Health

| Endpoint | Verb | Conceptual Intent / Behavior |
| :--- | :--- | :--- |
| `/api/v1/health` | `GET` | Returns API/database health for frontend diagnostics. |
| `/api/health` | `GET` | Compatibility health endpoint with the same response shape. |

### A. Projects Resource (`/api/projects`)
Manages high-level projects.

| Endpoint | Verb | Conceptual Intent / Behavior |
| :--- | :--- | :--- |
| `/api/project/list` | `POST` | Retrieves a list of all active projects, including an aggregated overall project completeness score. |
| `/api/project/get` | `POST` | Retrieves a single project's details, including a summary of tasks grouped by their phases. |
| `/api/project/create` | `POST` | Creates a new project workspace. Takes a name and optional description. |
| `/api/project/update` | `POST` | Updates project details (name, description, or status). |
| `/api/project/delete` | `POST`| Permanently deletes a project and cascades deletion to all associated tasks and requirements. |

Project delete handlers must archive every affected current `task` row to `task_history` with `deleted = true` and every affected current `requirement` row to `requirement_history` with `deleted = true` before deleting active rows.

### B. Tasks Resource (`/api/tasks`)
Manages individual tasks and maps them to projects and workflow phases.

| Endpoint | Verb | Conceptual Intent / Behavior |
| :--- | :--- | :--- |
| `/api/task/list`| `POST` | Retrieves all tasks belonging to a specific project. Tasks should contain metadata detailing their phase and calculated completeness. |
| `/api/task/get` | `POST` | Retrieves a single task and includes all its associated requirements. |
| `/api/task/create` | `POST` | Creates a new task under a specific project using a valid phase/type from the existing `task_phase` and `task_type` tables. |
| `/api/task/update` | `POST` | Updates task fields (title, description). |
| `/api/task/phase` | `POST`| Updates the workflow phase of the task to one of the existing `task_phase` options. |
| `/api/task/delete` | `POST`| Deletes a task and all associated requirements. |

Task update, phase-change, and delete handlers must archive the current `task` row to `task_history` before changing the active row. For updates and phase changes, history records use `deleted = false`. For deletes, history records use `deleted = true`. If deleting a task removes child requirements, each affected current `requirement` row must also be archived to `requirement_history` with `deleted = true`.

### C. Requirements Resource (`/api/requirements`)
Provides granular requirement actions that drive the completeness progress calculation.

| Endpoint | Verb | Conceptual Intent / Behavior |
| :--- | :--- | :--- |
| `/api/requirement/list`| `POST` | Retrieves all requirements for a specific task. |
| `/api/requirement/create` | `POST` | Appends a new requirement (a Definition of Done step) to a task. |
| `/api/requirement/update` | `POST`| Updates requirement text or toggles the completion status (`completed = true/false`). *Note: Triggering this endpoint causes the server to instantly recalculate the parent task's completeness percentage.* |
| `/api/requirement/delete` | `POST`| Deletes a requirement and triggers parent task completeness recalculation. |

Requirement update and delete handlers must archive the current `requirement` row to `requirement_history` before changing the active row. For updates, history records use `deleted = false`. For deletes, history records use `deleted = true`. If an update payload omits completion state, the active row's current `done` value must be preserved rather than defaulting to false.

Requirement create, update, and delete responses should include the recalculated parent task and current requirement list so the frontend can refresh task completeness without an extra fetch. Requirement ordering must use existing columns only, currently `created` and `definition`; do not add schema migrations or synthetic ordering fields from API work.

### D. AI Planning & Copilot Resource (`/api/planning`)
Interfaces with the language model provider to process user intent and produce project assets.

| Endpoint | Verb | Conceptual Intent / Behavior |
| :--- | :--- | :--- |
| `/api/planning/decompose` | `POST` | Takes a natural language prompt (e.g. a feature name or high-level project goal) and returns a structured suggestion of task phases validated against the existing database, task titles, task types, and individual requirements generated by the LLM. |
| `/api/planning/chat` | `POST` | Takes a prompt and optional project context, returning conversational assistance from the planning copilot. Useful for brainstorming workflows, identifying architecture patterns, or refining existing task requirements. |

---

## 3. Conceptual Response Patterns & Error Handling

### Standard Success Response Pattern
Successful GET requests generally return a standard response wrapped in a root object:
- For singular resources: Direct resource object with its fields.
- For collections: An array of objects.
- To ensure ease of rendering, calculated calculations (like `completeness_percentage` or `tasks_count_by_phase`) are computed server-side and injected directly into response bodies.

### Standard Error Handling
Echo's central error handler should catch all database or service failures and map them to standard JSON errors:
- **400 Bad Request:** Occurs on invalid parameters or malformed JSON payloads.
- **404 Not Found:** Occurs when looking up a project, task, or requirement that does not exist.
- **500 Internal Server Error:** Occurs on database connection drops or LLM API timeouts. The backend should return a sanitized error message to the client while logging the detailed raw error server-side for the developer.

---

## 4. Iterative API Roadmap Suggestions
1. **Endpoint Testing:** In Phase 1, these endpoints can be rapidly simulated and validated using standard curl requests, Bruno, or Postman before the frontend UI is connected.
2. **Phase and Type Updates:** Task phases and task types must be validated against the existing `task_phase` and `task_type` tables. Do not hardcode an independent list or mutate those reference tables from API code.
3. **SSE/WebSockets (V2):** If live multi-user dashboard updates are required in the future, Echo can easily support Server-Sent Events (SSE) or WebSockets on these same endpoints to broadcast changes instantly.
4. **Audit Transactions:** Any endpoint that mutates or deletes `task` or `requirement` must write the matching history row and active-row change in one transaction. A failed history write must fail the whole request.
