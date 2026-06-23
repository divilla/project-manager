# ToDo

- wire task view page directly to the backend api - at the moment it's taking stuff from memory or local storage - IMPORTANT
- enable closing of tasks - IMPORTANT

## Chat-GPT opinion on composing PRs

I’d use a hybrid, with the PR body as a concise map and the markdown spec files as the source of detail.

For your workflow, that is better than putting everything in the PR body.

The PR body should be short but complete enough to understand the change without opening every file:

## Summary
- Adds task detail hierarchy/requirements view.
- Adds backend-rendered task markdown descriptions.
- Adds scoped GitHub markdown styling with dark mode support.

## Specs
- Task frontend behavior: `agent/features/08-frontend-tasks.md`

## Notes for review
- Markdown is rendered server-side through `pkg/markdown`.
- Frontend renders only sanitized `description_html`.
- Parent descriptions are shown in reverse hierarchy order below the current task.

## Verification
- `go test ./...`
- `pnpm --dir frontend typecheck`
- `pnpm --dir frontend test -- --runInBand`
- `pnpm --dir frontend build`

Why I prefer this:

- PR body stays reviewable. Long PR descriptions become noisy and people stop reading them.
- Specs stay versioned near the code. The markdown file becomes the durable feature record, not a one-time GitHub comment.
- AI review works better with explicit references. If the PR body says “spec: agent/features/08-frontend-tasks.md”, an AI reviewer can inspect that file directly and compare implementation against it.
- Less duplication. If the full spec lives both in the PR body and in markdown, they will drift.
- Better future context. Six months later, the repo still contains the spec; the PR body is harder to discover from the working tree.

The key is: do not make the PR body too thin. “See spec file” alone is bad. The PR body should explain what changed, why, where the spec is, and what deserves reviewer attention. The detailed acceptance rules belong in the markdown spec.

### Maybe add requirements - definition of done to PR...
