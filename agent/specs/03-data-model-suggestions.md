# Spec 03: Data Model Suggestions

> **Developer Note:** These database schemas are purely advisory suggestions designed to support the completeness-tracking and AI planning modules. Since database control remains entirely under your jurisdiction, feel free to adapt, extend, or rewrite this schema to match your development preferences.

---

## 1. Conceptual Entity Relationship
The database is structured to track projects, represent their breakdown into tasks across different development phases, and measure completeness via discrete, verifiable DoD requirements.

```
  +--------------+          1             * +--------------+
  |   projects   | ------------------------> |    tasks     |
  +--------------+                           +--------------+
  | id (PK)      |                                  | 1
  | name         |                                  |
  | description  |                                  |
  | status       |                                  | *
  | created_at   |                                  v
  +--------------+                         +------------------+
                                           |   requirements   |
                                           +------------------+
                                           | id (PK)          |
                                           | task_id (FK)     |
                                           | description      |
                                           | completed        |
                                           +------------------+
```

---

## 2. Core Tables (Suggested Schema)

### A. Table: `projects`
Stores high-level metadata of individual software products or feature tracks.
- `id` (UUID or SERIAL, Primary Key): Unique identifier.
- `name` (VARCHAR, Not Null): Name of the project.
- `description` (TEXT): High-level summary of the project scope.
- `status` (VARCHAR): Current status (e.g., `'active'`, `'archived'`, `'draft'`).
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### B. Table: `tasks`
Stores the individual tasks or work items.
- `id` (UUID or SERIAL, Primary Key): Unique identifier.
- `project_id` (FK -> `projects.id`, ON DELETE CASCADE)
- `title` (VARCHAR, Not Null): Brief name of the task.
- `description` (TEXT): In-depth details or requirements.
- `task_phase` (VARCHAR, Default `'backlog'`): The workflow phase.
  - *Suggested Phases:* `'backlog'`, `'planning'`, `'in_progress'`, `'review'`, `'completed'`.
- `completeness` (NUMERIC/INT, Default `0`): A percentage integer representing task completion. This is a cached, derived value updated by database triggers or backend API calculations.
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### C. Table: `requirements`
The foundational unit of "completeness." Each single Definition of Done (DoD) item is defined as a 'requirement' in this table. Each task is broken down into one or more checkable requirements.
- `id` (UUID or SERIAL, Primary Key): Unique identifier.
- `task_id` (FK -> `tasks.id`, ON DELETE CASCADE)
- `description` (TEXT, Not Null): What needs to be done (e.g., *"Write API handler for CSV upload"*).
- `completed` (BOOLEAN, Default `FALSE`): Current binary status of this step.
- `order_index` (INT, Default `0`): For sorting the requirements in the UI.
- `created_at` (TIMESTAMP)
- `updated_at` (TIMESTAMP)

### D. Table: `planning_chats` (For AI Copilot)
To track chat threads inside the Planning screen, providing context to the LLM.
- `id` (UUID or SERIAL, Primary Key)
- `project_id` (FK -> `projects.id`, Nullable)
- `prompt` (TEXT, Not Null): The user input or directive.
- `response` (TEXT, Not Null): The AI response.
- `created_at` (TIMESTAMP)

---

## 3. Dynamic Progress Calculations
Instead of requiring developers to manually drag progress sliders, task completeness is directly derived from the requirement state.

### Recalculation Algorithm (Backend or DB Level)
For any task $T$:
$$\text{Completeness}(T) = \left( \frac{\text{Count of Completed Requirements}}{\text{Total Count of Requirements}} \right) \times 100$$

*   **Rule 1:** If a task has 0 requirements:
    *   If `task_phase = 'completed'`, completeness = `100%`.
    *   Otherwise, completeness = `0%`.
*   **Rule 2:** When a user checks a requirement, the backend recalculates `completeness` for that task and updates the row.
*   **Rule 3 (Phase Integration):** A task's completion metric is compiled alongside other tasks in the same `task_phase` to generate overall Phase Completeness on the Dashboard.

---

## 4. Key Considerations for Schema Implementation
1. **Using Triggers vs. Application Logic:**
   - *Application Logic (Go):* Easier to debug. When requirements are modified, the Go backend runs a transactional update to recalculate the task percentage and project-level aggregate statistics.
   - *Database Triggers (SQL):* Ensures integrity regardless of backend bugs. Updates to `requirements` automatically update `tasks.completeness`.
   - *Recommendation:* Start with **Application Logic** in Go for fast iteration, then move to triggers in the full-scale version if needed.
2. **Indexes:**
   - An index on `tasks.project_id` and `requirements.task_id` is highly recommended to keep dashboard page loads sub-second as the volume of tasks grows.
3. **Data Migrations:**
   - Use a migration manager in Go (such as `golang-migrate/migrate` or `pressly/goose`) to track schema evolutions, making it easy to transition from the local developer prototype to production.