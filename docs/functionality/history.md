# History

## Purpose
History preserves reviewable versions of planning data. It supports audit, review, and revert-oriented workflows for both user and AI changes.

## Change History
Change artifact history is document-specific. Creating a Change stores the active `idea` and creates the first `change_history` row for `doc_type = 'idea'`.

Updating `idea`, `spec`, or `pr` updates the active Change artifact, increments the active version, stores the supplied `agent_edit` value, updates `modified`, and inserts a matching history row in the same operation. Each Change history row records:

- id
- version
- doc_type
- body
- agent_edit
- modified
- deleted

PR URL is a non-null string artifact field but does not create document-specific history in the active contract. Title, type, epic, phase, open, PR URL, Flow, Run, and Test Case behavior must continue without generic Change history capture.

Change history does not store Flow snapshot or Run state fields, including `flow_stages`, `flow_stage_modes`, or any `run_*` column.

## Epic History
Before updating or deleting an epic, the backend records the current epic row in `epic_history`. Epic name updates increment the active epic version after history capture so repeated updates and later deletes create distinct history rows.

Epic history supports review of planning container changes and preserves previous aggregate context.

## Test Case History
Before updating or deleting a test case scenario, the backend records the current test case row in `test_case_history`. Scenario updates increment the active test case version after history capture so repeated edits and later deletes create distinct history rows.

Done toggles update active completion state without changing the test case scenario version. Test case history rows store the associated `change_id` so historical scenario changes remain tied to the change they helped define.

## Transaction Rule
History insert and active-row mutation must happen in one transaction. If history capture fails, the active row must not change.

## AI Changes
AI-initiated artifact updates follow the same document-specific history rules as user-initiated artifact updates. The artifact update payload supplies `agent_edit`, and the active row plus inserted history row must preserve that value for review.
