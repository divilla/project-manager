# Change Lifecycle

## Overview
A change is the delivery unit. It can exist independently or reference one epic. It is never part of a nested change tree.

## Create
Creating a change requires:

- project ID
- title
- optional requirement body
- optional pull request body
- optional pull request URL
- workflow phase from `change_phase`
- one or more types from `change_type`
- optional epic ID

The backend validates the project and reference options before insert.

Codex-assisted planning tools may create planned changes after the user confirms the generated test cases. These changes use the `backlog` phase until the user moves them through the normal lifecycle.

## List
Project-scoped lists show active changes grouped by workflow phase. Ordering follows database-provided phase priority and change ordering rules.

## Detail
The detail view shows:

- title and requirement body
- pull request body and URL when present
- phase and type information
- linked epic when present
- test case list
- completion counters
- version and modified time

Markdown requirement body rendering is sanitized by the backend before display.

## Update
Editing a change can update title, requirement body, pull request body, type classification, epic reference, phase, and closed state. History-bearing fields must preserve the previous row before mutation.

## Delete
Deleting a change is destructive and must be confirmed. Test cases linked to the change are archived or removed according to backend history rules before the active change is removed.

## Epic Link
A change may reference one epic. Updating this reference updates aggregate completeness for the old and new epic as needed.
