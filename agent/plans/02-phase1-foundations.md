# Plan 02: Phase 1 Foundations

The objective of Phase 1 is to establish the core software engineering scaffolding. By the end of this phase, we will have a running backend API, a running frontend UI, an active PostgreSQL database link, and fully operational CRUD interfaces for Projects and Tasks.

---

## 1. Step-by-Step Developer Implementation Tasks

### Step 1.1: Existing Database Connection
1. Ensure a PostgreSQL server is running locally on port `5432`.
2. Connect to the existing database named `postgres` or the configured database URL.
3. Inspect the existing schema and use the current table/column names exactly as-is.
4. Do not create migrations, create tables, alter tables, or seed lookup/reference data.
5. Treat `task_phase` and `task_type` as fixed reference tables that already contain all valid options.
6. Treat `task`, `requirement`, `task_history`, and `requirement_history` as existing tables. Use created/modified timestamps and history writes according to the database contract.

### Step 1.2: Go/Echo Backend Scaffolding
1. **Initialize Module:** Run `go mod init project-manager` inside the backend directory.
2. **Install Dependencies:** Fetch required packages:
   - `go get github.com/labstack/echo/v4`
   - `go get github.com/labstack/echo/v4/middleware`
   - `go get github.com/jackc/pgx/v5` (or preferred DB connector).
3. **Database Driver Setup:** Create a `db` package that opens a connection pool with PostgreSQL using environment variables (e.g. `DATABASE_URL`).
4. **Echo Server Setup:** Initialize an Echo router in `main.go`, attaching the Logger, Recovery, and CORS middlewares (enabling origins like `http://localhost:8000`).
5. **Start Server:** Bind the application to port `8080`.

### Step 1.3: Core Project & Task APIs
1. Create DB repositories or service layers to list, create, and delete records for Projects and Tasks using the existing database schema.
2. Map the resource routes:
   - `GET /api/projects` -> list projects.
   - `POST /api/projects/create` -> create project.
   - `GET /api/projects/:id/tasks` -> list tasks for a project.
   - `POST /api/tasks/create` -> create task.
3. Test endpoints using a REST client (e.g., Postman, Curl, Bruno) to verify that queries execute properly and return valid JSON content.
4. For task update/delete routes, verify that the current `task` row is copied to `task_history` before the active row changes. Delete history rows must use `deleted = true`.
5. For task/project delete routes that remove child requirements or tasks, verify every affected current row is archived to the matching history table with `deleted = true` before removal.

### Step 1.4: Vue/Quasar Frontend Scaffolding
1. **Frontend Skeleton:** Use the existing Vite/Vue/Quasar skeleton under `frontend/` as the template for further development.
2. **Navigation Shell Setup:**
   - Modify the default layout (`MainLayout.vue`) to include a clean top navigation bar.
   - Add `<q-tabs>` matching the four core views: Home, Planning, Projects, Help.
3. **Routing Configuration:**
   - Set up the Vue Router (`src/router/routes.js`) to handle the four primary page containers:
     - `/` mapping to `IndexPage.vue` (Home).
     - `/planning` mapping to `PlanningPage.vue`.
     - `/projects` mapping to `ProjectsPage.vue`.
     - `/help` mapping to `HelpPage.vue`.

### Step 1.5: Frontend-to-Backend Connectivity
1. Install Axios (`npm install axios`) in the Quasar project.
2. Create an API helper file (`src/boot/axios.js` or standard utility) that configures a base Axios instance pointing to the local Echo backend at `http://localhost:8080`.
3. In `ProjectsPage.vue`, write an asynchronous `onMounted` function that fetches active projects and displays them in a neat list layout using `<q-list>` or `<q-card>`.
4. Implement a simple "Create Project" dialog (`q-dialog`) containing form inputs that sends a `POST` request to the backend.

---

## 2. Success Criteria & Verification Checklist

To complete Phase 1, verify the following checks pass:

- [ ] **DB Link:** Go backend successfully starts without logging database connection errors.
- [ ] **Echo Server Routing:** Accessing `http://localhost:8080/api/projects` via browser or curl returns a valid HTTP `200` response containing a JSON array (even if empty).
- [ ] **Vue/Quasar Layout:** Running the frontend dev command (`npm run dev` from `frontend/`) compiles without errors and loads the UI at `http://localhost:8000`.
- [ ] **Reactive Navigation:** Clicking between Home, Planning, Projects, and Help in the top menu dynamically updates the URL path and changes the page container context without triggering full-page browser reloads.
- [ ] **Integration Check:** Creating a project through the Quasar UI immediately registers the record using the existing PostgreSQL schema, and refreshing the list renders the newly created project card instantly.
- [ ] **Schema Safety:** No migrations, schema changes, or `task_phase`/`task_type` data changes are introduced.
- [ ] **Task History Safety:** Updating or deleting a task writes the previous current version to `task_history` in the same transaction.
- [ ] **Cascade History Safety:** Project/task deletes archive all affected task and requirement rows with `deleted = true`.
