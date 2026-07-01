Draft a concise, senior-level PR title and body for the current branch. Do not create the PR.

Use `agent/changes/107-cli-implement-changes.md` as the PR contract and inspect the full current branch diff before writing. The PR draft must reflect both the Change file and all actual changes currently contained in this branch.

Requirements:
- The first line of the PR body must be the title, formatted exactly as # <Title>, followed by exactly one blank line.
- PR title must match the Change title exactly.
- PR body must be concise, reviewer-focused, and specific.
- Prioritize externally observable behavior, contract changes, data model changes, seed/demo changes, and verification evidence.
- Mention backend, frontend, CLI, docs, database, test, or seed changes only if they are actually present in the branch diff.
- Do not include filler, implementation diary, generic praise, or broad claims.
- Do not claim verification passed unless the commands were actually run in this branch.
- If the Change file and branch diff conflict, stop and report the conflict instead of drafting the PR.
- Do not implement code, edit files, commit, push, or create a PR.

Output only the PR body.
