# Change Lifecycle

## Overview
A change is the delivery unit. It can exist independently or reference one epic. It is never part of a nested change tree.

## Create
Creating a change requires:

- project ID
- title
- optional body
- workflow phase from `change_phase`
- one or more types from `change_type`
- optional epic ID

The backend validates the project and reference options before insert.

## List
Project-scoped lists show active changes grouped by workflow phase. Ordering follows database-provided phase priority and change ordering rules.

## Detail
The detail view shows:

- title and body
- phase and type information
- linked epic when present
- requirement list
- completion counters
- version and modified time

Markdown body rendering is sanitized by the backend before display.

## Update
Editing a change can update title, body, type classification, epic reference, phase, and closed state. History-bearing fields must preserve the previous row before mutation.

## Delete
Deleting a change is destructive and must be confirmed. Requirements linked to the change are archived or removed according to backend history rules before the active change is removed.

## Epic Link
A change may reference one epic. Updating this reference updates aggregate completeness for the old and new epic as needed.
