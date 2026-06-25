> The app becomes a structured PR builder.

More precisely:

> It becomes a system for turning intent into reviewed, verifiable PR-sized changes.

The flow would be:

Backlog intent
  ↓
Structured work item
  ↓
Requirements / acceptance criteria
  ↓
Agent-readable implementation context
  ↓
Branch / PR
  ↓
Verification and review
  ↓
Merge / close / follow-up

In that model, the app is not just a task manager. It is a change-management system for AI-assisted development.

The core entities might become:

- Work item: the unit of intent
- Requirement: what must be true
- Acceptance criterion: how to verify it
- Specification reference: source of rules/context
- Implementation task: concrete work step
- PR: the artifact proving the work
- Review finding: feedback that may reshape the PR
- Decision: why scope changed
- Verification run: evidence that checks passed/failed

The PR is the centerpiece, but not the only thing. It is the delivery container.

A strong conceptual model:

Work item owns intent.
Requirements define correctness.
PR carries implementation.
Verification proves readiness.
Review controls acceptance.

If you go this route, I’d stop thinking of the app as generic project management. It’s more specific and more valuable:

> an AI-native PR planning and verification system.

That specificity helps a lot. It tells you what belongs in the app and what does not.



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
