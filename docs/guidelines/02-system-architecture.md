# Spec 02: System Architecture

## 1. High-Level Topology
The system is constructed as a modern, decoupled web application. Because this is a developer prototype designed for single-user or trusted local team usage, all services can run seamlessly on a local developer machine (`localhost`).

```
 +-----------------------------------------------------------+
 |                    Developer Machine                      |
 |                                                           |
 |  +-----------------------+     HTTP       +------------+  |
 |  |  VueJS / Quasar UI    | -------------> | Go / Echo  |  |
 |  |  (Vite Dev Server)    | <------------- | Backend    |  |
 |  +-----------------------+     JSON       +------------+  |
 |                                                 |    ^    |
 |                                      SQL/TCP    |    |    |
 |                                                 v    |    |
 |                                           +------------+  |
 |                                           | PostgreSQL |  |
 |                                           | Database   |  |
 |                                           +------------+  |
 +-------------------------------------------------|---------+
                                                   | HTTP / JSON
                                                   v (External)
                                             +------------+
                                             | LLM Provider|
                                             | (OpenAI/   |
                                             | Anthropic) |
                                             +------------+
```

---

## 2. Component Breakdowns

### A. Frontend (VueJS / Quasar Framework)
- **Engine:** Vue 3 powered by Vite for instant hot-module reloading and high development velocity.
- **UI Framework:** Quasar Framework (v2) utilizing material design principles. It provides highly polished UI components (tables, dialogs, progress bars, layout layouts) out of the box, reducing front-end development time by 80%.
- **Client Communication:** Axios (or standard Fetch) sending asynchronous HTTP requests.
- **State Management:** Pinia or Vue Reactivity (Composition API) for managing current project state, tasks, and copilot chat sessions.

### B. Backend (Go / Echo Framework)
- **Language:** Go (Golang) for fast compilation, type safety, and minimal memory overhead.
- **Web Framework:** Echo (v4) chosen for its high performance, simple routing, and extensible middleware stack.
- **API Server Responsibilities:**
  - Serving REST API endpoints for projects, tasks, and requirements.
  - Coordinating database interactions using a standard Go database package (`database/sql`, `pgx`, or an ORM like `GORM`).
  - Handling prompt construction, structured payload generation, and communication with the LLM API.
  - Managing application configuration (e.g., database connection string, LLM API keys) via environment variables.

### C. Database (PostgreSQL)
- **Instance:** PostgreSQL v14+ running locally (`postgresql://localhost:5432/postgres`).
- **Role:** Storing all application state persistently.
- **Approach:** Use the existing relational schema exactly as it is. The prototype must not create migrations, alter tables, seed lookup values, or rename database concepts. The existing `task_phase` and `task_type` reference tables are authoritative and already contain all allowed options.

### D. AI Service Layer
- **LLM Engine:** Configurable connector supporting major providers (OpenAI GPT-4o, Anthropic Claude, or local Ollama instances running Mistral/Llama3).
- **Communication:** Standard HTTPS REST calls with API keys passed securely through server-side environment variables.
- **Workflow:** The user communicates with the copilot in the UI; the Go backend wraps user intent in a structured context prompt, invokes the LLM, parses the returned JSON payload, and returns clean, structured data to the client.

---

## 3. Data Flow Scenarios

### Scenario A: Real-Time Completeness Calculation
1. **User Action:** The developer checks off a requirement on a task in the Vue UI.
2. **Client Event:** The frontend sends an HTTP `POST` request to `/api/requirements/:id/toggle` with `{ "completed": true }`.
3. **Backend Action:** 
   - The Go server receives the request and starts a transaction.
   - It copies the current `requirement` row to `requirement_history` with `deleted = false`.
   - It updates the current `requirement` row and modified timestamp, then queries all requirements for that task.
   - It recalculates the task's completeness percentage based on active vs. completed requirements.
   - If the task row changes, it first copies the current `task` row to `task_history` with `deleted = false`, then updates the task state and modified timestamp in the DB.
4. **Aggregation:**
   - The server aggregates the task-level progress up to the project-level and phase-level completeness.
5. **Response:** The server returns the updated task status and project-wide phase statistics as JSON. The Vue UI dynamically updates the progress bars.

### Scenario B: AI Task Decomposition
1. **User Action:** The developer enters a high-level feature description in the Planning screen: *"Add CSV Export to project tables"*.
2. **Client Event:** Frontend sends an HTTP `POST` request to `/api/planning/decompose` containing the prompt.
3. **Backend Processing:**
   - The server fetches project metadata to provide context (existing modules, technology stack, etc.).
   - It compiles a structured system prompt, injecting guidelines on breaking down tasks into 3-5 phase-grouped, verifiable sub-requirements using the phase/type options loaded from the existing database.
   - It invokes the LLM API requesting a structured JSON response matching the layout of our task list.
4. **Parsing & Refinement:**
   - The server parses the LLM's response.
   - It saves the proposed plan in a temporary state in the DB and returns the proposed list to the frontend.
5. **Approval:** The user reviews the list in the UI, checks/unchecks items, and clicks "Create Tasks". This writes the approved requirements into the main task tables.

---

## 4. Key Design Considerations for Prototyping
1. **Strict Decoupling:** Keep frontend and backend strictly separated. Do not embed frontend files into Go binary assets during the prototype phase; run separate watch servers (Vite at `http://localhost:8000` / Go on port `8080`) to ensure fast feedback loops.
2. **Simplified Security:** In this developer-only version, we bypass authentication filters and CORS restrictions on local networks, allowing rapid prototyping of actual product features.
3. **Graceful Degradation of AI:** If the LLM call fails (e.g. timeout or token-limit reached), the backend must return a friendly error and allow the user to manually input their sub-tasks without crashing the application.
4. **Backend Layout & AI Logic Directive:** The skeleton backend application is built within the root `backend/` directory. AI agents or planning algorithms operating in this codebase must strictly apply these exact same design standards (POST-only mutating actions, `/create` suffixes, and `'requirement'` naming) when generating task lists or generating new application code.
5. **Database Ownership Directive:** Agents must treat the existing database as read/write application storage only through the current schema. They must not generate migration files, mutate `task_phase` or `task_type`, or assume hardcoded phase/type values beyond what the database returns.
6. **History Directive:** Any user or AI update/delete of `task` or `requirement` must archive the current row to `task_history` or `requirement_history` in the same transaction before changing the active row. Delete operations must write the history row with `deleted = true`.
