# Feature 04: Home Dashboard

## 1. Purpose
Give the developer an immediate read on project progress using completeness grouped by task phase. The dashboard replaces time estimation views with empirical progress based on requirements.

## 2. Prototype Scope
- Select an active project.
- Show overall project completeness.
- Show completeness grouped by `task_phase`.
- Show task counts per phase.
- Highlight bottlenecks and stale-looking work.

## 3. Dashboard Metrics
The dashboard should include:

- Overall project completeness.
- Phase completeness for each phase returned by the existing `task_phase` table.
- Task count per phase.
- Requirement count and completed requirement count where useful.
- Bottleneck list for tasks that need attention.

## 4. Aggregation Rules
Project completeness should be derived from task completeness, which is derived from requirements.

Phase completeness:

```text
phase_completeness = average completeness of tasks in that phase
```

Project completeness:

```text
project_completeness = average completeness of all project tasks
```

If a project has no tasks, the dashboard should show an empty state rather than 0% as a failure condition.

## 5. Bottleneck Signals
The prototype can start with simple heuristics:

- Tasks in a database-provided active/progress phase with 0% completeness.
- Tasks in a database-provided review/verification phase below 100% completeness.
- Tasks with no requirements.
- Tasks in a database-provided completion phase with incomplete requirements.

## 6. User Experience
The Home page should be scan-friendly:

- A project selector at the top.
- Overall completeness near the top of the page.
- Phase cards or rows with progress bars.
- Bottleneck watch list below the phase summary.

Progress visualizations should be useful without pretending to forecast dates.

## 7. API Notes
The dashboard can be served by either a dedicated summary endpoint or by project/task list responses that include aggregate data. A dedicated endpoint is preferable once the board grows:

- `POST /api/dashboard/project-summary`

Prototype payload should include:

- project metadata
- overall completeness
- phase summaries
- bottleneck tasks

## 8. Acceptance Criteria
- User can select a project and see its completeness overview.
- Completeness is grouped by phase.
- Phase task counts match the project task list.
- Bottleneck list identifies at least the initial simple heuristics.
- Dashboard uses requirement-derived completeness only.
- Phase labels and ordering come from the existing `task_phase` table.
