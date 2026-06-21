# Plan 02: Phase 1 Foundations

The objective of Phase 1 is to establish the core software engineering scaffolding. By the end of this phase, we will have a running backend API, a running frontend UI, an active PostgreSQL database link, and fully operational CRUD interfaces for Projects and Tasks.

---

## 1. Step-by-Step Developer Implementation Tasks

### Step 1.1: Database Initialization
1. Ensure a PostgreSQL server is running locally on port `5432`.
2. Connect to the server and create a dedicated database named `postgres` (or use the existing default).
3. Create the initial database schema tables for `projects` and `tasks` as outlined in `specs/03-data-model-suggestions.md`.
4. Insert 1-2 mock project rows and a few mock tasks directly into the database to serve as initial seed data.

### Step 1.2: Go/Echo Backend Scaffolding
1. **Initialize Module:** Run `go mod init project-manager` inside the backend directory.
2. **Install Dependencies:** Fetch required packages:
   - `go get github.com/labstack/echo/v4`
   - `go get github.com/labstack/echo/v4/middleware`
   - `go get github.com/jackc/pgx/v5` (or preferred DB connector).
3. **Database Driver Setup:** Create a `db` package that opens a connection pool with PostgreSQL using environment variables (e.g. `DATABASE_URL`).
4. **Echo Server Setup:** Initialize an Echo router in `main.go`, attaching the Logger, Recovery, and CORS middlewares (enabling origins like `http://localhost:9000`).
5. **Start Server:** Bind the application to port `8080`.

### Step 1.3: Core Project & Task APIs
1. Create DB repositories or service layers to list, create, and delete records for Projects and Tasks.
2. Map the resource routes:
   - `GET /api/projects` -> list projects.
   - `POST /api/projects` -> create project.
   - `GET /api/projects/:id/tasks` -> list tasks for a project.
   - `POST /api/tasks` -> create task.
3. Test endpoints using a REST client (e.g., Postman, Curl, Bruno) to verify that queries execute properly and return valid JSON content.

### Step 1.4: Vue/Quasar Frontend Scaffolding
1. **Scaffold CLI:** Create a clean Quasar installation using Vite (Vue 3, Pinia) by running `npm init quasar` or utilizing the Quasar CLI.
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
- [ ] **Vue/Quasar Layout:** Running the front-end dev watch command (`quasar dev`) compiles without errors and loads the UI in the browser.
- [ ] **Reactive Navigation:** Clicking between Home, Planning, Projects, and Help in the top menu dynamically updates the URL path and changes the page container context without triggering full-page browser reloads.
- [ ] **Integration Check:** Creating a project through the Quasar UI immediately registers the record in the PostgreSQL database, and refreshing the list renders the newly created project card instantly.