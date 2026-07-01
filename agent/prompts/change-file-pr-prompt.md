Draft a concise, senior-level PR title and body for the current branch. Do not create the PR.

Use `agent/changes/109-db-alters-and-views.md` as the PR contract and inspect the full current branch diff before writing. The PR draft must reflect both the Change file and all actual changes currently contained in this branch.

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

codex exec -C /home/vito/go/src/project-manager "$(cat <<'PROMPT'
Draft a concise, senior-level PR title and body for the current branch. Do not create the PR.

Read the Change file from stdin and use it as the PR contract and inspect the full current branch diff before writing. The PR draft must reflect both the Change file and all actual changes currently contained in this branch.

Requirements:
- The first line of the PR body must be the title, formatted exactly as # <Title>, followed by exactly one blank line.
- PR title must match the Change title exactly.
- PR body must be reviewer-focused, and specific.
- Prioritize externally observable behavior, contract changes, data model changes, seed/demo changes, and verification evidence.
- Mention backend, frontend, CLI, docs, database, test, or seed changes only if they are actually present in the branch diff.
- Do not include filler, implementation diary, generic praise, or broad claims.
- Do not claim verification passed unless the commands were actually run in this branch.
- If the Change file and branch diff conflict, stop and report the conflict instead of drafting the PR.
- Do not implement code, edit files, commit, push, or create a PR.

Output only the PR body.
  PROMPT
  )" < agent/changes/109-db-alters-and-views.md

(
    printf '%s\n' 'Draft a concise, senior-level PR body. Do not create the PR.'
    printf '%s\n' 'Use only the stdin provided Change file and diff. Do not inspect additional files.'
    printf '%s\n' 'Output only the PR body using markdown. Title must precisely match the one in Change file.'
    printf '%s\n\n<change-file>'
    cat agent/changes/109-db-alters-and-views.md
    printf '%s\n\n<diff-stat>' '</change-file>'
    git diff --stat "$(git merge-base HEAD stage)"..HEAD
    printf '%s\n\n<diff>' '</diff-stat>'
    git diff --find-renames "$(git merge-base HEAD stage)"..HEAD
    printf '%s\n' '</diff>'
) | codex exec -C /home/vito/go/src/project-manager --sandbox read-only --ephemeral -

