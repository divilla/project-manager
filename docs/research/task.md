# Task

Goal:
- What outcome we want.

Requirements:
- Functional requirement 1
  - Acceptance criteria
  - Edge cases
- Functional requirement 2
  - Acceptance criteria

Constraints:
- Architecture constraints
- Existing APIs/specs to follow
- Things not to change

References:
- Design/spec document
- API contract
- Existing module/file examples
- Tests that should pass

Why this helps:

- Hierarchy gives scope control. The AI can see the big picture and the smaller obligations.
- Specifications reduce ambiguity. They say what “correct” means.
- References anchor the work in reality. They prevent the AI from inventing APIs, styles, or behavior.
- Acceptance criteria make verification easier. The AI can translate them into tests or manual checks.

But there is a tradeoff: overly long or contradictory requirements can hurt. The best form is not “maximum detail”; it is structured, relevant detail.

A useful hierarchy for AI work is:

Objective
Context
Requirements
Acceptance criteria
Non-goals
Relevant files/specs
Verification commands

Example:

Objective:
Add task deletion confirmation to the task board.

Requirements:
1. Clicking delete must not immediately call the delete API.
2. A persistent confirmation dialog must open.
3. Cancel closes the dialog without deleting.
4. OK calls the existing delete endpoint.
5. After success, reload the current project tasks.

Constraints:
- Reuse DeleteConfirmationDialog.vue.
- Do not change backend APIs.
- Keep existing board layout.

Verification:
- Add/update component tests.
- Run pnpm --dir frontend test.
