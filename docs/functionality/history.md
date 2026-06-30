# History

## Purpose
History preserves prior active row states before update or delete behavior. It supports audit, review, and revert-oriented workflows for both user and AI changes.

## Change History
Before updating or deleting a history-bearing change field, the backend records the current change row in `change_history`.

History-bearing change data includes:

- project ID
- epic ID
- change types
- title
- requirement body
- pull request body
- pull request URL
- modified time
- delete marker

## Epic History
Before updating or deleting an epic, the backend records the current epic row in `epic_history`.

Epic history supports review of planning container changes and preserves previous aggregate context.

## Test Case History
Before updating or deleting a test case scenario, the backend records the current test case row in `test_case_history`.

Done toggles may update active state without changing the test case scenario version, depending on backend contract. Test case history rows store the associated `change_id` so historical scenario changes remain tied to the change they helped define.

## Transaction Rule
History insert and active-row mutation must happen in one transaction. If history capture fails, the active row must not change.

## AI Changes
AI-initiated updates follow the same history rules as user-initiated updates. The product should make AI changes reviewable and reversible through the same history model.
