# Test Cases And Acceptance

## Purpose
Test cases define what complete means for a change. They convert broad intent into binary checks that can be reviewed by a human or agent.

## Test Case Rules
A test case should be:

- Binary: complete or incomplete.
- Verifiable: evidence can prove the result.
- Concrete: names a behavior, artifact, test, or decision.
- Small: can be evaluated independently.

## Completeness
Change completeness is derived from linked test cases:

```text
completed test cases / total test cases * 100
```

If a change has no test cases, it should not appear complete unless explicit product rules say otherwise.

## Mutation Behavior
Test case create, update, done toggle, reassignment, and delete actions run through the backend. Responses should provide enough current data for the frontend to refresh visible completeness without guessing.

## Acceptance Criteria
Acceptance criteria define the user-visible or system-visible outcomes required for a change to be considered done. They should be written as testable statements, not vague goals.

## Planning Output
LLM-generated planning output must produce concrete test cases. The user reviews and edits suggestions before anything is saved.
