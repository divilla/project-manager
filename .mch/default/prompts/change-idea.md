---
name: change-idea
description: Rewrite the active repository Change idea file in place. Use when the user asks to clean up, polish, clarify, or rewrite the active Change idea without turning it into a Change spec, requirements document, or implementation plan.
---

Rewrite the active Change idea into clear Markdown while preserving the user's intended scope and level of detail.

## Preconditions

If the user asks only to explain, discuss, summarize, or review the idea without rewriting it, answer normally and do not edit files.

Extract `<change-name>` from the current Git branch using `^change/([0-9]+-[a-z0-9-_]+)$`.

If the branch does not match, stop with: `Branch must be change/<change-name>.`

If `agent/ideas/<change-name>.md` does not exist, stop with one concise error.

## Instructions

Read `agent/ideas/<change-name>.md`.

Rewrite the idea in place:

- Preserve all meaningful user intent.
- Preserve the original level of detail.
- Keep the document an idea draft.
- Fix grammar, wording, ordering, and readability.
- Remove repetition when it does not remove meaning.
- Clarify vague wording only when the intended meaning is obvious.
- Do not add Change spec sections, requirements, acceptance criteria, implementation plans, product decisions, or new scope.

Do not modify any other file.

After successfully rewriting the idea, output exactly:

Idea
