# Plan

Flow State     Meaning                                           AI/PR Artifact
━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Backlog        Intent exists, not being changed yet              Epic/story/requirement spec
─────────────  ────────────────────────────────────────────────  ───────────────────────────────────
  In Progress    Work is being implemented                         Branch + draft PR
─────────────  ────────────────────────────────────────────────  ───────────────────────────────────
  Staging        Implementation done, under verification/review    Ready PR with tests, review notes
─────────────  ────────────────────────────────────────────────  ───────────────────────────────────
  Closed         Accepted or abandoned                             Merged PR or closed PR

The PR becomes the container that ties together:

- requirement scope
- implementation
- tests
- review findings
- verification evidence
- final decision

A good AI-oriented flow:

1. Backlog
    - Define goal, requirements, acceptance criteria, non-goals.
    - Link specs and examples.
    - No code yet.

2. In Progress
    - Create branch and draft PR early.
    - AI implements against the PR spec.
    - Commit incrementally.
    - Update PR description as decisions change.

3. Staging
    - PR is feature-complete.
    - AI runs verification scripts.
    - AI performs self-review.
    - Human or another AI reviews the diff.
    - Findings become follow-up commits on the same PR.

4. Closed
    - If merged: requirements are considered satisfied.
    - If rejected/abandoned: record why, maybe split back into backlog items.

This works especially well with AI because it gives the agent a bounded mission:

> Make this PR satisfy these requirements and pass these checks.

The main risk is making PRs too large. AI can produce broad changes quickly, but review quality drops if the PR combines too many concerns. So I’d enforce:

- One PR = one coherent requirement slice.
- PR description must include acceptance criteria.
- PR cannot enter staging without automated verification.
- Review findings either block the PR or become explicit follow-up backlog items.
- Avoid “misc cleanup” inside feature PRs.

For your states, I’d define them like this:

Backlog
A requirement package exists, but no implementation branch is active.

In Progress
A branch/draft PR exists and may be unstable.

Staging
The PR is believed complete and is being validated. No new scope should be added here except fixes.

Closed
The PR has reached a terminal decision: merged, rejected, superseded, or abandoned.

The key rule:

> Backlog items define intent; PRs prove intent was satisfied.



## Task Types

feature      new capability
fix          behavior correction
refactor     no behavior change
chore        maintenance
upgrade      dependency/runtime update
docs         documentation only
test         test-only
ci           build or pipeline
migration    schema/data evolution
security     vulnerability or hardening
revert       undo change

Useful distinction:

- Feature changes what users can do.
- Fix changes behavior to match intended requirements.
- Refactor changes how code is structured, not what it does.
- Upgrade changes external dependency versions.
- Chore keeps the project healthy but is not directly user-visible.
