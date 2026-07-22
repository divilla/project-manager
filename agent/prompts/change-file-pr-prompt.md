Draft a senior-level PR title and body for the current branch. Do not create the PR.

Use Change spec `specs/<change-slug>.md` as the intended Change instructions. Before writing,
inspect the full current branch diff, including untracked files. The PR draft must summarize what
the code changed, why it changed, verification evidence, and any Spec instructions intentionally
deferred.

Documentation is outside the default Change Flow. Do not inspect it by default or add documentation
work to the PR unless it is present in the diff. Read a retained ADR or operational runbook only
when the Spec explicitly cites it or the diff changes it.

Requirements:

- Write the PR body to `agent/prs/<change-slug>.md`.
- The first line of the PR body must be the title, formatted exactly as # <Title>, followed by exactly one blank line.
- The PR title must match the Change title exactly.
- Keep the PR body reviewer-focused, and specific.
- Prioritize externally observable behavior, contract changes, data model changes, seed/demo changes, and verification evidence.
- Mention backend, frontend, CLI, docs, database, test, or seed changes only if they are actually present in the branch diff.
- Include a References section listing the Change spec. Add an ADR or runbook only when it was
  explicitly part of the Change.
- Do not include filler, implementation diary, generic praise, or broad claims.
- Do not claim verification passed unless those commands were actually run in this branch.
- If a Spec instruction is absent from the diff, identify it as deferred instead of presenting it
  as implemented. Stop only when the diff contains unresolved material behavior outside the Spec.
- Do not implement code, edit files other than the PR draft file, commit, push, or create a PR.
