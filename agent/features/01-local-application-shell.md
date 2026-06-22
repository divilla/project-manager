# Feature 01: Local Application Shell

## 1. Purpose
Provide the prototype with a simple local-first web application shell: a Vue/Quasar frontend, a Go/Echo backend, and a PostgreSQL database connection that can be run by a developer without account setup or hosted infrastructure.

This feature exists to make the rest of the product usable quickly. It intentionally avoids user management, authentication, task assignment, and production deployment concerns.

## 2. Prototype Scope
- Vue 3 + Quasar single-page application with top navigation.
- Frontend development server runs at `http://localhost:8000`.
- Go/Echo JSON API server running locally.
- PostgreSQL connection using `postgresql://localhost:5432/postgres` by default.
- Existing database schema used as-is, with no migrations or schema mutation by agents.
- Basic health/config checks so setup failures are obvious.
- Local development CORS support between the Vite/Quasar dev server at `http://localhost:8000` and the Go API server.

## 3. Out of Scope
- Login, signup, sessions, OAuth, RBAC, and tenant separation.
- Embedded frontend assets inside the Go binary.
- Production hosting, SSL, reverse proxy, or cloud database setup.
- Database migrations, schema changes, lookup-table seeding, or reference-data mutation.
- User profiles, teams, notifications, or task assignments.

## 4. User Experience
The first screen should be the usable application, not a landing page. The top menu contains:

- Home
- Planning
- Projects
- Help

If the backend or database is unavailable, the frontend should show a clear non-blocking error state instead of an empty page.

## 5. Backend Responsibilities
- Load configuration from environment variables and `backend/config/dev.yaml` where appropriate.
- Start the Echo server on the configured port, defaulting to `:8080`.
- Expose a basic health endpoint suitable for frontend diagnostics.
- Keep health diagnostics as `GET /api/v1/health` and `/api/health`; do not convert health checks to POST when applying the prototype's POST-based mutating-resource convention.
- Provide centralized JSON error handling.
- Use zerolog for backend logging, including Echo request logging and internal package logs.
- Enable CORS for local frontend origins.

## 6. Frontend Responsibilities
- Use Quasar layout primitives for a stable app frame.
- Configure routes for `/`, `/planning`, `/projects`, and `/help`.
- Provide a shared API client for JSON requests.
- Surface loading, empty, and error states consistently across pages.

## 7. Acceptance Criteria
- A developer can start the backend and frontend locally with documented commands.
- Running the frontend dev command serves the app at `http://localhost:8000`.
- The frontend top navigation routes to Home, Planning, Projects, and Help.
- The frontend can call the backend successfully from the local dev server.
- Database connection failures are visible in logs and understandable in the UI.
- Backend code uses zerolog only; no standard-library structured logging package usage remains in the project.
- No code or agent workflow creates migrations, changes schema, or mutates `task_phase`/`task_type`.
- No authentication or user-selection UI exists in the prototype.
