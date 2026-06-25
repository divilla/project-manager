# Work item

I would avoid calling the top-level thing a task if you want clean planning around PRs.

Use:

> Work item

as the generic term.

Then specialize it by type:

- epic
- feature
- fix
- refactor
- upgrade
- chore
- docs
- test
- ci
- security
- migration
- revert
- spike

Why not “task”?

Because task is overloaded. It can mean:

- a tiny implementation step
- a user story
- a Jira ticket
- a background job
- a to-do item
- a PR-sized change
- a child of a feature

For your model, I’d use this hierarchy:

Epic
  Feature / Fix / Refactor / Upgrade / Chore ...
    Requirement
      Acceptance criterion
    Implementation task

Or, if planning around PRs:

Work item
  type: feature | fix | refactor | upgrade | ...
  phase: backlog | in_progress | staging | closed
  requirements
  acceptance criteria
  linked PR

Then reserve task for implementation-level steps, like:

- “Add backend endpoint”
- “Update Pinia store”
- “Add route test”
- “Run API tests”

A good naming model:

  Name                    Use
━━━━━━━━━━━━━━━━━━━━━━  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Work item               Generic planning unit
──────────────────────  ───────────────────────────────────────
  Epic                    Large business/product capability
──────────────────────  ───────────────────────────────────────
  Feature                 PR-sized or release-sized capability
──────────────────────  ───────────────────────────────────────
  Issue                   Problem/bug/report needing resolution
──────────────────────  ───────────────────────────────────────
  Task                    Concrete implementation action
──────────────────────  ───────────────────────────────────────
  Requirement             Condition the system must satisfy
──────────────────────  ───────────────────────────────────────
  Acceptance criterion    Testable evidence requirement is met

So if your app stores PR-planned units, I’d name the entity work item or change request, not task.

My preference:

> Use work item internally, and show user-facing labels like Feature, Fix, Chore, etc.
