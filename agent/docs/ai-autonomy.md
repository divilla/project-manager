## AI autonomy

1. AI can freely explore and change within a PR branch.
2. Every meaningful change is captured as semantic history and can be reverted.

That is a strong model.

The risky part is:

> “AI can do anything, change anything.”

I’d constrain that to:

> “AI can change anything inside the declared PR scope, and every change must be attributable, inspectable, and reversible.”

That gives freedom without losing control.

The model could be:

PR workspace
  changes
  semantic history
  checkpoints
  decisions
  verification runs
  review comments
  revert points

Git already gives you raw history, but not smart history. Smart history should capture why and what it meant.

Example smart-history event:

{
  "type": "scope_change",
  "actor": "agent",
  "summary": "Added backend rendered descriptions endpoint",
  "reason": "Task detail page needed sanitized parent markdown without rendering all task descriptions in list payload.",
  "files": [
    "backend/internal/task/api.go",
    "backend/internal/task/service.go",
    "frontend/src/features/tasks/api/taskApi.ts"
  ],
  "requirements_affected": ["REQ-08-MARKDOWN-02"],
  "verification": ["make api-test", "pnpm --dir frontend test"],
  "revert_group": "markdown-rendering"
}

That is much more useful than a raw commit message.

A good architecture would use layers of reversibility:

  Layer                  Purpose
━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Git commit             Exact file-level revert
─────────────────────  ──────────────────────────────────────────
  Patch/checkpoint       Revert a work session
─────────────────────  ──────────────────────────────────────────
  Smart-history event    Explain intent and affected requirements
─────────────────────  ──────────────────────────────────────────
  Requirement trace      Know what behavior a change served
─────────────────────  ──────────────────────────────────────────
  Verification run       Know whether revert is safe

I would still keep hard safety boundaries:

- Never change outside workspace.
- Never run destructive commands without explicit approval.
- Never mutate production-like data directly.
- Never alter schema unless PR scope allows it.
- Never silently expand PR intent.
- Never bypass tests if verification is available.

Within those boundaries, yes: let the AI work freely.

The key is to make “current PR” a sandboxed change context:

current_pr
  branch
  objective
  allowed scope
  protected files / forbidden actions
  smart history
  checkpoints
  verification policy

Then the agent can say:

> I’m about to make a checkpoint before changing routing and task cache behavior.

And after:

> Checkpoint created. Event recorded: “Refactor task loading into shared cache.” Verification pending.

The best version of your idea is not “AI can do anything.” It is:

> AI can act with high autonomy inside a PR workspace because every meaningful action is scoped, logged, testable, and reversible.

That is a very promising architecture for AI coding.



## Change classification

Yes. I would classify AI changes by intent, risk, and reversibility, not only by “feature/fix/refactor”.

For smart-history, every AI change event should answer:

> Why was this changed, what kind of change is it, how risky is it, and how can it be undone?

A practical classification:

  Dimension              Values
━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Intent                 implement, fix, refactor, verify, document, configure, investigate
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Scope Area             backend, frontend, database, tests, docs, tooling
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Semantic Impact        behavior, structure, contract, presentation, verification, metadata
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Risk Level             low, medium, high
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Reversibility          easy, moderate, hard
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Origin                 planned, discovered, review-requested, agent-initiated
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Requirement Link       requirement IDs or none
─────────────────────  ─────────────────────────────────────────────────────────────────────
  Verification Status    unverified, passed, failed, not-applicable

Example:

{
  "intent": "fix",
  "scope_area": ["frontend"],
  "semantic_impact": ["behavior"],
  "risk": "medium",
  "reversibility": "easy",
  "origin": "review-requested",
  "requirements": ["REQ-project-switch-guardrail"],
  "summary": "Prevent edit route from saving tasks outside current project context."
}

Useful intent types:

implement      adds required behavior
fix            corrects incorrect behavior
refactor       changes structure without intended behavior change
adapt          adjusts implementation to fit existing architecture
verify         adds or changes tests/checks
document       updates specs/docs/comments
configure      changes build/dev/runtime configuration
cleanup        removes unused or obsolete code
investigate    exploratory read-only or prototype work
revert         undoes prior change

Useful semantic impact types:

behavior       changes runtime behavior
contract       changes API/types/schema/interface
structure      reorganizes code
state          changes persistence/cache/session handling
presentation   changes UI layout/visuals/copy
verification   changes tests or validation
operations     changes scripts, CI, dev workflow
documentation  changes docs/specs

Useful origin types:

planned          directly from PR requirements
derived          necessary sub-change inferred from planned work
discovered       found during implementation
review_requested from PR review comment
user_requested   explicit human instruction
agent_suggested  autonomous improvement proposed by agent

Then you can group changes into change sets:

Change Set: task-detail-routing
  - implement frontend route
  - adapt task cache
  - verify route tests

Change Set: markdown-rendering
  - implement backend renderer
  - contract add description_html
  - verify markdown sanitizer tests

That gives you better revert behavior than reverting individual files.

For each change set, store:

{
  "id": "chg-2026-06-25-001",
  "summary": "Add backend markdown rendering",
  "origin": "derived",
  "intent": "implement",
  "impact": ["contract", "behavior"],
  "areas": ["backend", "frontend"],
  "requirements": ["REQ-MARKDOWN-HTML"],
  "files": [...],
  "depends_on": [],
  "verification": {
    "commands": ["make test", "make api-test"],
    "status": "passed"
  },
  "revert": {
    "strategy": "git_patch",
    "checkpoint": "abc123"
  }
}

My recommendation:

> Classify individual changes lightly, but classify change sets carefully.

Individual file edits are too noisy. The meaningful unit is:

> “I changed these files together to achieve this reason.”

That is the right level for smart-history and revert.



# The Flow

Not necessarily. A smart-history change and a Git commit are related, but they should not be identical concepts.

A Git commit is a storage/version-control unit.

A smart-history change is a semantic unit.

Sometimes they align:

Smart change: Add task detail route
Git commit:   Add task detail route

But often they won’t.

One smart change may require several commits:

Smart change: Add markdown rendering
  commit 1: backend renderer
  commit 2: API contract
  commit 3: frontend rendering
  commit 4: tests

One commit may contain several smart changes:

Commit: implement feature 08
  smart change A: task detail route
  smart change B: create route
  smart change C: markdown rendering
  smart change D: project switch guardrail

For your system, I’d separate them like this:

Checkpoint
  raw Git state / patch snapshot

Smart change
  semantic description of why something changed

Commit
  optional VCS packaging for review/merge

A good model:

  Concept         Purpose
━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Smart change    Meaning: why and what changed
──────────────  ───────────────────────────────
  Checkpoint      Revert point during AI work
──────────────  ───────────────────────────────
  Commit          Git history unit
──────────────  ───────────────────────────────
  PR              Review/delivery unit

If you want easy revert, every smart change should have a checkpoint or patch boundary. But that does not require every smart change to become a permanent Git commit.

Possible workflow:

1. AI starts change set.
2. System records pre-change checkpoint.
3. AI edits files.
4. System records affected files and semantic metadata.
5. Tests run.
6. Later, human decides how to squash/split commits.

So during AI work:

many smart changes + checkpoints

At PR finalization:

clean Git commits

This is often better than making every AI action a Git commit, because AI may make noisy intermediate edits.

My recommendation:

> Treat smart changes as semantic checkpoints first, then compile them into clean commits later.

That gives you both:

- fine-grained reversibility while developing
- readable Git history when merging.
