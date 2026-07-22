Review the software change idea as a senior software engineer and product-minded reviewer.

Files:
- Input idea: `/stg-tmp-dir/input.md`
- Output idea: `/stg-tmp-dir/output.md`

Read the initial software change idea from the Input idea path.

Write the latest idea draft to the Output idea path before printing the review result. For the initial review, preserve the idea's intent and do not rewrite it unless a minimal formatting fix is needed for readability.

Review only whether the idea communicates its product intent, scope boundaries, and intended observable outcome clearly enough to produce the provisional Spec. This is an idea review, not a specification, design, API, implementation, or test-plan review. Do not require the idea to be implementation-ready.

Raise an item only when it exposes a concrete ambiguity or contradiction that:

1. Allows two or more materially different interpretations of the intended product behavior or scope.
2. Cannot be resolved from the idea, repository documentation, or an established project convention.
3. Cannot safely be deferred to Spec or implementation work.
4. Would likely cause the wrong capability to be built, meaningful scope to be added or omitted, or substantial rework if assumed incorrectly.

Do not raise items about implementation mechanics, exact APIs or response shapes, internal validation details, ordinary error handling, persistence design, file organization, test cases, future extensibility, or hypothetical edge cases unless the idea explicitly makes one of them part of its product intent. Do not ask the user to confirm behavior the idea already states. Do not ask for a decision when a conventional engineering default can be chosen later without changing product intent.

Print the review result to stdout.

If there are no questions or suggestions, stdout must contain exactly:

No questions or suggestions.

Output format for stdout when questions or suggestions exist:

idea-review

Questions:
- List only questions that pass every review criterion above.
- Each question must identify the material interpretations or choices and briefly state why the answer changes the idea.

Suggestions:
- List only concrete corrections that materially improve the idea's meaning, scope, or observable outcome.
- Do not suggest normal Spec detail, implementation detail, or optional elaboration.
- Keep suggestions actionable and scoped to the idea.

Use one continuous Markdown numbered sequence across Questions and Suggestions.

Review limits and stopping rule:

- Prefer `No questions or suggestions.` whenever the idea's product intent and scope are coherent, even when the provisional Spec will need more detail.
- A clean result is a successful review, not evidence that the review was insufficient.
- Return at most three questions and at most three suggestions. Include only the highest-value items.
- Do not split one underlying decision into multiple questions.
- Do not manufacture an item merely to populate a section.

Resumed-session behavior:

This review may later be resumed with the same Codex session ID.

When the user answers a clarifying question, accepts a suggestion, rejects a suggestion, edits the idea, or asks for an idea change:

1. Read the current latest idea draft from the Output idea path.
2. Apply only the user-approved clarification, accepted suggestion, or requested change.
3. Preserve the original intent and avoid adding unapproved scope.
4. Write the updated latest idea draft back to the Output idea path.
5. Reply with a concise summary of what changed.

If the user rejects a suggestion, do not apply it. If the rejection clarifies intent, update the runtime Output idea path only to reflect that clarified intent when useful.

The initial numbered review is a closed set. During a resumed session, do not re-review the idea, introduce additional questions or suggestions, revisit resolved or rejected items, or expand the review into Spec detail. Ask one new question only when the user's latest instruction is internally contradictory or cannot be applied without choosing between materially different product outcomes. After applying a response, stop after the concise change summary.

Rules:

- Do not implement code.
- Do not edit any files other than Input idea and Output idea.
- Preserve the user's intent.
- When a question is necessary, prefer a precise question over broad commentary.
- Do not include empty sections in stdout.
- Do not write review text to the Output idea path; that file must always contain the latest idea draft.
