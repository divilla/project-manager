Draft a senior-level PR title and body for the current branch. Do not create the PR.

Use Change file `agent/changes/110-cli-improve-changes-and-agentic-workflow.md` as the PR contract. Before writing, inspect the full current branch diff, including untracked files. The PR draft must accurately reflect both the Change file and every
actual change currently contained in the branch.

Requirements:

- Write the PR body to `agent/prs/110-cli-improve-changes-and-agentic-workflow.md`.
- The first line of the PR body must be the title, formatted exactly as # <Title>, followed by exactly one blank line.
- The PR title must match the Change title exactly.
- Keep the PR body reviewer-focused, and specific.
- Prioritize externally observable behavior, contract changes, data model changes, seed/demo changes, and verification evidence.
- Mention backend, frontend, CLI, docs, database, test, or seed changes only if they are actually present in the branch diff.
- Include a References section listing the Change file as mandatory, plus all relevant docs files used by the branch.
- Do not include filler, implementation diary, generic praise, or broad claims.
- Do not claim verification passed unless those commands were actually run in this branch.
- If the Change file and branch diff conflict, stop and report the exact conflict instead of drafting the PR.
- Do not implement code, edit files other than the PR draft file, commit, push, or create a PR.
