---
name: change-verify
description: Verify or review a Change spec as the contract for a Change. Use when asked to verify a Change, review a Change spec, audit a Change spec against docs or implementation, check whether a Change is complete, or produce blocking findings from a Change-file review. Do not use for implementing fixes.
---

# Change Verify

Verify the Change spec as the source of truth as a very senior engineer. Build fresh context from the repository, compare the Change spec with linked documentation and available implementation evidence, and return only actionable findings. The primary purpose of verification is to verify is to answer `Is Change spec complete, internally consistent, and implementation-ready?`, not `Has the implementation already satisfied it?`


## Preconditions

Read `AGENTS.md` first when present. Repository instructions override this skill when stricter.

Identify the Change spec:

- If the user names a Change spec, use that file.
- Otherwise, derive `<change-slug>` from the current branch:

  ```bash
  git branch --show-current
  ```

  The branch must match:

  ```text
  ^change-([0-9]+-[0-9a-z-_]+)$
  ```

  If it does not match, stop with exactly:

  ```text
  The branch must be named change/<change-slug>.
  ```

- Require:

  ```text
  specs/<change-slug>.md
  ```

  If the file does not exist, stop with exactly:

  ```text
  Missing specs/<change-slug>.md.
  ```

This is a verification/review skill. Do not edit, format, stage, commit, reset, restore, migrate, seed, post PR comments, or otherwise mutate tracked files, database files, fixtures, snapshots, generated artifacts, or application data.

## Review Scope

Review the Change spec for:

- required structure: Goal, Scope, Requirements, Acceptance Criteria, Non-Goals, Design Notes, Relevant Specs, Verification, QA Test Cases, Review Focus, and Follow-Ups
- internal contradictions, vague or unverifiable requirements, missing acceptance criteria, and scope leaks
- requirements that conflict with linked docs, `AGENTS.md`, repository rules, or the stated non-goals
- missing verification commands or QA test cases for changed behavior
- missing linked specs/docs when behavior depends on them

When implementation evidence exists, verify the Change spec against it:

- inspect `git status --short`
- inspect committed, staged, unstaged, and relevant untracked work
- inspect the full diff against `origin/stage`
- read linked specs that define intended behavior
- verify every Requirement and Acceptance Criteria item against the diff and tests
- identify changed public contracts, API behavior, data model changes, migrations, tests, docs, and workflows
- run listed verification commands only when feasible and allowed by `AGENTS.md`

If `origin/stage` freshness was not verified, include the base-ref note at the end when findings exist.

## Finding Rules

Report only actionable blocking findings. Do not report style nits, preferences, broad refactors, speculative issues, or requests to expand scope.

Use severities:

- P0: data loss, security breach, production outage, or unrecoverable corruption
- P1: broken core workflow, incorrect persisted state, API contract break, serious regression, or Change-file contradiction that makes implementation unsafe
- P2: real bug, missing required coverage, unverifiable acceptance criterion, stale/racy behavior, or edge case with user-visible impact
- P3: minor correctness issue or low-risk ambiguity that can still cause confusion

For verification failures, include the failed command and the key failure message in the finding body. If a command cannot be run for an environment-only reason, report it only when that blocks confidence in a required acceptance criterion.

## Output Format

Return findings only, in this exact format:

```markdown
Review Findings:

- [P<severity>] <short imperative title> 
  - <repo-root relative file path>:<line or line-range>
  - <one concise paragraph explaining the concrete impact>
  - <one concise paragraph explaining when it occurs>
  - <one concise paragraph explaining specific fix direction>
```

Rules:

- Start with exactly `Review Findings:`
- Use one Markdown bullet per finding.
- Use file paths relative to <repo-root> and exact line or line range.
- Put the explanation on the next line, indented by two spaces.
- Add exactly one blank line after each finding.
- Do not include praise, summaries, or non-finding commentary.

If no blocking issues exist, return exactly:

```markdown
Review Findings:

No blocking issues found.
```

If findings exist and the base ref note is required, add this exact line after the findings:

```text
Base ref note: origin/stage freshness was not verified.
```
