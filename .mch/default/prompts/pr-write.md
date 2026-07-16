# Change PR

Draft a reviewer-focused GitHub PR body for the current Change. Do not create, update, push, merge, or comment on a pull request.

## Hard Limits

- Read repository instructions first when `AGENTS.md` is present. Follow repository instructions when they are stricter than this skill.
- Do not edit code, docs, tests, database files, fixtures, snapshots, generated artifacts, or application data.
- Only create or update the PR draft file at `agent/prs/<change-slug>.md`. Creating `agent/prs/` for that file is allowed.
- Do not stage, commit, reset, restore, format, migrate, seed, or mutate any database.
- Do not post PR comments.

## Identify The Change

- If the user names a Change spec or Change name, use it.
- Otherwise derive `<change-slug>` from the current branch:

  ```bash
  git branch --show-current
  ```

- The branch must match:

  ```text
  ^change/([0-9]+-[0-9a-z-_]+)$
  ```

- If the branch does not match, stop with exactly:

  ```text
  The branch must be named change/<change-slug>.
  ```

- Require `specs/<change-slug>.md`. If it does not exist, stop with exactly:

  ```text
  Missing specs/<change-slug>.md.
  ```

## Build Context

Read the Change spec as the PR contract before drafting. Then inspect the actual branch contents:

- `git status --short`
- full tracked diff against the PR base branch
- staged and unstaged diffs
- untracked files that would be relevant to the PR

Use the repository's workflow base branch when documented. In this Project Manager repository, default to `origin/stage` when available.

Read only the docs and source files needed to understand changed public behavior, contracts, verification, and references. For documentation references, prefer files named in the Change spec and files touched by the diff.

## Contract Check

Compare the Change spec with the branch diff before writing the PR draft.

Stop instead of drafting if:

- the Change spec requires behavior that is absent from the diff
- the diff includes material behavior outside the Change scope
- the PR title cannot be determined from the Change spec
- verification claims in the Change spec conflict with available evidence
- database or generated-file changes appear and repository instructions prohibit acting on them

Report the exact conflict with file paths and the specific Change requirement or diff item involved.

## Draft Requirements

Write the PR body to `agent/prs/<change-slug>.md`.

The first line must be the title, formatted exactly:

```markdown
# <Title>
```

Follow it with exactly one blank line. The title must match the Change title exactly. If the Change spec has no explicit title, derive the title from the first `# ...` heading in the Change spec. If no title can be found, stop and report that the Change title is missing.

Use this structure unless the repository gives a stricter PR template:

```markdown
# <Title>

## Summary
- ...

## Behavior
- ...

## Verification
- ...

## References
- `specs/<change-slug>.md`
- ...
```

Omit `## Behavior` only when the branch has no externally observable behavior or contract change. Include additional sections only when they help reviewers, such as `## Data Model`, `## API`, `## Frontend`, `## CLI`, `## Docs`, or `## Risks`.

## Writing Standards

- Be specific and concise. Write for reviewers who need to understand what changed, why it satisfies the contract, and how it was verified.
- Reflect the branch diff, not intentions. Mention backend, frontend, CLI, docs, database, tests, seed/demo data, or generated artifacts only if they are present in the diff.
- Prioritize externally observable behavior, public contracts, API payloads, data model changes, migrations, seed/demo changes, and verification evidence.
- Do not include implementation diary, generic praise, filler, broad claims, or unexplained internal details.
- Do not claim verification passed unless the exact command was run in this branch and the result is known.
- If verification was not run, say `Not run` and give the reason if known.
- If verification failed, include the exact command and the relevant failure summary. Do not soften or reinterpret failures as passing.
- Include a `References` section listing `specs/<change-slug>.md` and every relevant docs file used to understand or validate the branch.

## Final Response

After writing the draft, respond with:

- the PR draft path
- a brief summary of what the body covers
- verification commands observed or noted as not run
- any conflicts or residual risks that remain
