# Current Project Context

## Purpose
The current project is the active workspace for dashboards, planning, changes, and requirements. It prevents each page from inventing its own project selection behavior.

## Selector Rules
The app shows a compact project selector in the top bar.

Selection behavior:

- restore a valid persisted project ID
- otherwise select the project with the lowest ID
- clear selection when no projects exist
- keep selection after rename
- repair selection after delete

## Project Switching
When the user changes project context:

1. Mark the selector as switching.
2. Route to `/loading`.
3. Refresh project-scoped data.
4. Return to the topic index for the previous route.
5. Clear the switching marker.

Nested change routes return to `/changes` after project switching.

## Empty State
If no project exists, project-scoped features show an empty state and direct the user to create a project.

## Safety
Project deletion is blocked when changes exist. The backend owns the final safety check; the frontend can only explain or disable actions ahead of time.
