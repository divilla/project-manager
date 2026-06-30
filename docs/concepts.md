# Concepts

## Project
A project is the workspace for delivery. It owns epics and changes and provides the current context for dashboards, planning, and the change board.

Projects can be created, renamed, selected, and deleted when empty. A project with existing changes must not be deleted by cascade from normal UI behavior.

## Epic
An epic is a planning container. It groups related changes and receives aggregate completeness from those changes.

An epic is not a parent node in a nested change tree. It is a reference target.

## Change
A change is the primary unit of delivery and PR construction. It has a fixed structure and can be either standalone or linked to one epic.

Important fields:

- `ref`: project-scoped numeric reference allocated by the backend.
- `slug`: stable backend-owned identifier derived from the change reference and title.
- `title`: short human-readable name.
- `requirement_body`: markdown-capable description of the requirement or implementation intent.
- `pull_request_body`: optional markdown text for the eventual PR body.
- `pull_request_url`: optional link to the published PR.
- `change_phase`: current workflow phase.
- `change_types`: one or more classification slugs.
- `epic_id`: optional link to one epic.
- `closed`: completion marker.

`ref` is unique only inside its project. Two projects may both have a change with the same `ref`, but a single project must not.

Users and clients cannot set or edit `ref` or `slug`. The backend assigns them when the change is created and returns them on change responses.

## Test Case
A test case is a binary Definition of Done item for a change. It must be concrete, verifiable, and small enough to evaluate independently. Its `scenario` describes the condition that must be true.

Good examples:

- "API response includes the recalculated completeness fields."
- "Frontend detail view renders sanitized markdown from the backend."
- "History row is inserted before deleting the active record."

Weak examples:

- "Improve backend."
- "Make UI better."
- "Finish planning."

## History
History stores previous active row versions before update or delete operations. It supports audit, review, and revert-oriented workflows for user and AI changes.

## Completeness
Completeness is calculated from test cases:

```text
completeness = completed test cases / total test cases * 100
```

Epic completeness is derived from linked changes. Project summaries are derived from the active project data.
