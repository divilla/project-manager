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

- `ref`: optional project-scoped numeric reference allocated by backend Flow assignment.
- `ref_uuid`: backend-generated, globally unique UUID identity returned on Change responses.
- `slug`: optional backend-owned branch identifier generated from the Change reference and title.
- `title`: short human-readable name.
- `idea`: required user intent captured when the Change is created and replaceable by an agent rewrite.
- `spec`: optional markdown-capable requirement or implementation specification.
- `pr`: optional markdown text for the eventual PR body.
- `pr_url`: optional link to the published PR.
- `agent_edit`: marker that the current active version was produced by an agent-assisted edit.
- `open`: marker that the change remains active for work.
- `change_phase`: current workflow phase.
- `change_types`: zero or more classification slugs.
- `epic_id`: optional link to one epic.
- `flow_stages`: ordered Flow Steps copied onto the Change when Flow is assigned.
- `flow_stage_modes`: ordered per-Step execution modes copied onto the Change when Flow is assigned.
- `run_claim_id`: active Worker claim for the current Run, when claimed.
- `run_flow_stage`: current Flow Step for the Run, when started.
- `run_task_step`: current Task Step within the Run, when started.
- `run_task_status`: current Task status, when started.
- `run_error`: latest Run or active Task error, empty when no current error is recorded.
- `run_is_completed`: whether the latest Run completed.

`ref` is unique only inside its project. Two projects may both have a change with the same `ref`, but a single project must not.

Users and clients cannot set or edit `ref`, `ref_uuid`, `slug`, Flow snapshot fields, or Run state fields directly. New changes may have no assigned `ref`, `slug`, or Flow snapshot; clients must render that state without deriving identity locally. The backend assigns or refreshes `ref` and `slug` only through Flow assignment and returns identity on Change responses.

## Flow and Run
A Flow is a reusable automation definition for moving a Change through ordered Steps. A Step is one named stage inside that Flow. A Run is one execution attempt of the Flow copied onto a Change. A Task performs one Step within a Run. A Worker is the executor, tool, or process that performs the Task.

Default Flow Steps are ordered as `idea`, `spec`, `ready`, `docs`, `code`, `polish`, `pr`, `review`, `fix`, `sync`, `merge`, `stage`, and `master`.

Default Flow Step modes are `skip`, `prompt`, and `exec`. The default repository Flow help config includes those stage modes; task statuses `queued`, `running`, `paused`, `stopped`, `waiting`, `completed`, and `failed`; and task steps `none`, `entry`, `prompt`, `agent`, `exit`, and `done`. These help options are configuration data rather than validation lists for custom help display.

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
History stores reviewable versions of changes, epics, and test cases. Change artifact history is document-specific for `idea`, `spec`, and `pr`; epic and test case history preserve prior active row context for their supported update and delete flows.

## Completeness
Completeness is calculated from test cases:

```text
completeness = completed test cases / total test cases * 100
```

Epic completeness is derived from linked changes. Project summaries are derived from the active project data.
