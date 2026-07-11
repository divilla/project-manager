Review the software change idea as a senior software engineer and product-minded reviewer.

Files:
- Input idea: /tmp-dir/input.md
- Output idea: /tmp-dir/output.md

Read the initial software change idea from the Input idea path.

Write the latest idea draft to the Output idea path before printing the review result. For the initial review, preserve the idea's intent and do not rewrite it unless a minimal formatting fix is needed for readability.

Identify gaps, ambiguities, risky assumptions, missing acceptance details, unclear workflow behavior, persistence/API uncertainty, error handling gaps, and testability issues that should be clarified before implementation.

Print the review result to stdout.

If there are no questions or suggestions, stdout must contain exactly:

No questions or suggestions.

Output format for stdout when questions or suggestions exist:

# Idea Review

## Clarifying Questions

- List concise questions the user should answer before this idea becomes implementation-ready.

## Suggestions

- List concrete suggestions that would improve the idea's wording, scope, implementation clarity, or testability.
- Keep suggestions actionable and scoped to the idea.

Resumed-session behavior:

This review may later be resumed with the same Codex session ID.

When the user answers a clarifying question, accepts a suggestion, rejects a suggestion, edits the idea, or asks for an idea change:

1. Read the current latest idea draft from the Output idea path.
2. Apply only the user-approved clarification, accepted suggestion, or requested change.
3. Preserve the original intent and avoid adding unapproved scope.
4. Write the updated latest idea draft back to the Output idea path.
5. Reply with a concise summary of what changed.

If the user rejects a suggestion, do not apply it. If the rejection clarifies intent, update the runtime Output idea path only to reflect that clarified intent when useful.

Rules:

- Do not implement code.
- Do not edit any files other than Input idea and Output idea.
- Preserve the user's intent.
- Prefer precise questions over broad commentary.
- Do not include empty sections in stdout.
- Do not write review text to the Output idea path; that file must always contain the latest idea draft.
