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

- `ref`: optional project-scoped numeric reference allocated by the backend reference flow.
- `slug`: optional backend-owned branch identifier generated from the Change reference and title.
- `title`: short human-readable name.
- `idea`: required user intent captured when the Change is created and replaceable by an agent rewrite.
- `spec`: optional markdown-capable requirement or implementation specification.
- `pr_body`: optional markdown text for the eventual PR body.
- `pr_url`: optional link to the published PR.
- `agent_edit`: marker that the current active version was produced by an agent-assisted edit.
- `open`: marker that the change remains active for work.
- `change_phase`: current workflow phase.
- `change_types`: zero or more classification slugs.
- `epic_id`: optional link to one epic.

`ref` is unique only inside its project. Two projects may both have a change with the same `ref`, but a single project must not.

Users and clients cannot set or edit `ref` or `slug`. New changes may have no assigned `ref` or `slug`; clients must render that state without deriving identity locally. The backend assigns or refreshes them only through the Change reference flow and returns them on Change responses.

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
