# Plan 03: Phase 2 Core Tracking

The objective of Phase 2 is to implement the **Completeness Engine**—the defining architectural piece of this tool. By the end of this phase, the application will dynamically calculate completeness metrics for individual tasks based on requirements and aggregate those values by workflow phases to feed the analytical Home Dashboard.

---

## 1. Step-by-Step Developer Implementation Tasks

### Step 2.1: Requirement Table & Database Integration
1. Apply the advisory database table schema for `requirements` ( FK -> `tasks.id` ) to your PostgreSQL database.
2. In your Go backend, create the repository models for requirements and configure query handlers to support CRUD operations on requirement entities.

### Step 2.2: Dynamic Backend Completeness Calculator
1. In the Go backend, create a helper module or database hook that performs the following mathematical task percentage calculation:
   $$\text{Completeness}(T) = \left( \frac{\text{Completed Items}}{\text{Total Items}} \right) \times 100$$
2. Hook this calculator to the requirement mutation endpoints. When a requirement is added, toggled, or deleted:
   - Recalculate the parent task's completion percentage.
   - Run a SQL transaction updating the `tasks.completeness` database field.
3. Establish a conceptual aggregation query for a specific project. This query must sum up all requirements across all tasks within each separate `task_phase` (e.g., Backlog, Planning, In-Progress, Review, Completed) to return aggregated phase completeness:
   $$\text{PhaseCompleteness}(P) = \left( \frac{\sum \text{Completed Items in Phase } P}{\sum \text{Total Items in Phase } P} \right) \times 100$$

### Step 2.3: Task Detail Dialog & Requirements UI
1. In `ProjectsPage.vue`, configure the UI so clicking a task card triggers a modal dialog (`<q-dialog>`).
2. Inside the dialog, display the task title, description, and list its corresponding requirements.
3. Build an input box that allows the developer to quickly add a new Definition of Done step (creating a new requirement via `POST /api/requirements/create`).
4. Set up checkboxes (`<q-checkbox>`) for each item. Hook the `@update:model-value` change gesture to dispatch a `POST` request to `/api/requirements/:id/update` to toggle the completed state.
5. In the same dialog, display a highly visible `<q-linear-progress>` bar representing the task's individual completion percentage. The progress bar should dynamically animate when requirements are toggled.

### Step 2.4: Home Dashboard (Analytical Overview)
1. Build out `IndexPage.vue` (Home screen) as the analytical control center.
2. Build an async function that fetches aggregated phase completeness statistics from the backend on page mount.
3. Display a large project progress card displaying the master project completeness score.
4. Render five distinct metrics grid blocks representing the standard development phases (`backlog`, `planning`, `in_progress`, `review`, and `completed`). 
5. Each block must feature:
   - A bold percentage label.
   - A `<q-linear-progress>` indicator showing phase completeness.
   - A small text summary showing total active tasks in that phase.

---

## 2. Success Criteria & Verification Checklist

To complete Phase 2, verify the following checks pass:

- [ ] **Requirement Persistence:** Adding, checking, and deleting requirements is persistent across browser reloads.
- [ ] **Calculated Logic:** If a task features 4 requirements, marking one as completed immediately reflects a task progress of `25%` in the database and in the task modal progress bar.
- [ ] **Dynamic Re-aggregation:** When you check off a task requirement in the Projects screen and click back to the Home Dashboard, the corresponding phase progress bar (and overall project progress) immediately animates to display the recalculated progress.
- [ ] **Zero-State Resilience:** A newly created task with zero requirements defaults to `0%` progress gracefully without triggering dividing-by-zero math errors in Go or Vue.
- [ ] **Phase Alignment:** Tasks can be successfully dragged or selected to change phases, and the dashboard aggregates them under their new phase instantly.