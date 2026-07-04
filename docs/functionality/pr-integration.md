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
change/<change-name>
```

The matching Change spec lives at:

```text
specs/<change-name>.md
```

Workflow automation must reject `changes/<change-name>` as the active Change branch namespace. Application routes such as `/changes` and API paths such as `/api/v1/change/*` remain unchanged.

## Checkpoint Commits
Planning checkpoints use:

```text
Change <change-name> edit by user
Change <change-name> edit by agent
```

Implementation uses:

```text
Implement change <change-name>
```

## PR Body
The Change spec becomes the basis for the PR body. It should be complete enough that reviewers can understand intent, scope, evidence, and review focus without reconstructing context from chat.
