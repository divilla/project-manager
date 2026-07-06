# PR Integration

## Purpose
The product acts as a structured PR builder. A Change spec defines the PR contract before implementation starts.

## Change Spec
A Change spec records:

- goal
- scope
- requirements
- acceptance criteria
- non-goals
- design notes
- relevant docs
- verification commands
- review focus
- follow-ups

## Branch Naming
Change branches use:

```text
change/<change-slug>
```

The matching Change spec lives at:

```text
specs/<change-slug>.md
```

Workflow automation must reject `changes/<change-slug>` as the active Change branch namespace. Application routes such as `/changes` and API paths such as `/api/v1/change/*` remain unchanged.

## Checkpoint Commits
Planning checkpoints use:

```text
Change <change-slug> edit by user
Change <change-slug> edit by agent
```

Implementation uses:

```text
Implement change <change-slug>
```

## PR Body
The Change spec becomes the basis for the PR body. It should be complete enough that reviewers can understand intent, scope, evidence, and review focus without reconstructing context from chat.
