# Requirements And Test Cases

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

## Display And Mutation Behavior
Change detail views show linked test cases in numeric ID order. Test case create, update, done toggle, reassignment, and delete actions run through the backend. Responses should provide enough current data for clients to refresh visible completeness and done state without guessing.

## Requirements
Requirements define the user-visible or system-visible behavior a change must deliver. They should be written as testable obligations rather than vague goals or implementation tasks.

## Planning Output
LLM-assisted planning first rewrites the user's idea. The user confirms creation before the rewritten idea is saved as a Change.

Spec generation is a separate flow. Until that flow is implemented, choosing the temporary spec-writing prompt should leave the saved idea intact and route to Change details without creating local-only spec content.
