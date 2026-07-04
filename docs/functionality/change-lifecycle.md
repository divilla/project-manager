# Change Lifecycle

## Overview
A change is the delivery unit. It can exist independently or reference one epic. It is never part of a nested change tree.

Each change may have a backend-owned `ref` and `slug`. The `ref` is a project-scoped numeric reference allocated from the owning project. The `slug` is a branch identifier generated from the assigned reference and current title.

Users and clients must not create, edit, or overwrite `ref`, `slug`, or the project's reference counter. New Changes may exist without a `ref` or `slug`; these fields are assigned or refreshed only by the backend reference flow and returned to clients as read-only identity data.

## Create
Creating a change requires:

- project ID
- title
- idea

The title and idea must be non-blank after validation. `spec`, `pr_body`, `pr_url`, and `epic_id` are optional and default to null when omitted. `change_types` defaults to an empty array when omitted or explicitly empty. `change_phase` defaults to `backlog` when clients do not provide a phase.

After a successful create, the returned change includes its database ID and may have unassigned `ref` and `slug` values. Creating a Change does not advance the project reference sequence.

Codex-assisted planning tools create Changes after the user confirms `Create Change?`, then may run an agent rewrite and save the rewritten idea through the dedicated agent-edit update flow. These Changes use the backend default phase until the user moves them through the normal lifecycle.

## Reference
`POST /api/v1/change/reference` assigns or returns identity for an existing Change identified by numeric `id`.

For an unreferenced Change, the endpoint allocates the next project-scoped `ref`, generates `slug`, preserves existing artifacts, and returns refreshed Change data. For a referenced Change, the endpoint must not allocate a new `ref`; it may refresh `slug` from the current title for explicit branch reconciliation.

## List
Project-scoped lists show active changes for one project. Lists include list-appropriate fields only: identity, phase and type data, linked epic identity and `epic_name` when present, title, `agent_edit`, open state, completion counters, and modified time.

List items include `ref` and `slug` when assigned and must represent unassigned identity without deriving it locally. Clients must use the backend response order.

## Detail
The detail view shows:

- project-scoped reference and slug
- title, idea, and spec
- `pr_body` and `pr_url` when present
- phase and type information
- linked epic and `epic_name` when present
- test case list ordered by test case ID
- completion counters
- `agent_edit`
- open state
- version, created time, and modified time

Markdown `spec` and `pr_body` rendering is sanitized by the backend before display.

## Update
Editing a change can update title, `idea`, `spec`, `pr_body`, `pr_url`, type classification, epic reference, phase, open state, and `agent_edit`. History-bearing fields must preserve the previous row before mutation. Open-state updates use an explicit boolean value and return refreshed backend state.

Focused updates return the refreshed change with its existing `ref` and `slug`. Updates must work before and after reference assignment. Updating the title does not let clients supply replacement identity values.

## Delete
Deleting a change is destructive and must be confirmed. Test cases linked to the change are archived or removed according to backend history rules before the active change is removed.

## Epic Link
A change may reference one epic. Updating this reference updates aggregate completeness for the old and new epic as needed.
