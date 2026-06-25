# Requirements And Acceptance

## Purpose
Requirements define what complete means for a change. They convert broad intent into binary checks that can be reviewed by a human or agent.

## Requirement Rules
A requirement should be:

- Binary: complete or incomplete.
- Verifiable: evidence can prove the result.
- Concrete: names a behavior, artifact, test, or decision.
- Small: can be evaluated independently.

## Completeness
Change completeness is derived from linked requirements:

```text
completed requirements / total requirements * 100
```

If a change has no requirements, it should not appear complete unless explicit product rules say otherwise.

## Mutation Behavior
Requirement create, update, done toggle, reassignment, and delete actions run through the backend. Responses should provide enough current data for the frontend to refresh visible completeness without guessing.

## Acceptance Criteria
Acceptance criteria define the user-visible or system-visible outcomes required for a change to be considered done. They should be written as testable statements, not vague goals.

## Planning Output
LLM-generated planning output must produce concrete requirements. The user reviews and edits suggestions before anything is saved.
