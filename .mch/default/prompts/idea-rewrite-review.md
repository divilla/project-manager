Rewrite and then review a software change idea as a senior software engineer and product-minded reviewer.

Files:
- Input idea: /tmp-dir/input.md
- Output idea: /tmp-dir/output.md

First, read the initial software change idea from the Input idea path.

Rewrite the idea for clarity, structure, grammar, and readability without changing its intent, scope, or product decisions.

Write the rewritten idea to the Output idea path.

Then review the rewritten idea from the Output idea path. Use the original Input idea only as a source-of-truth check to ensure the rewrite preserved intent and did not add or remove meaningful scope.

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
- Keep suggestions actionable and scoped to the rewritten idea.

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

- Preserve the original intent exactly.
- Do not add new requirements, features, workflows, APIs, persistence behavior, or test expectations unless they are already clearly implied by the input.
- Do not remove meaningful details from the original idea.
- Improve wording, organization, headings, and bullet structure where useful.
- Keep fenced code blocks unchanged.
- Do not implement code.
- Do not edit any files other than the Input idea and Output idea.
- Do not write review text to the Output idea path; that file must always contain the latest idea draft.
- Prefer precise questions over broad commentary.
- Do not include empty sections in stdout.
