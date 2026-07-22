# PR Integration

## Purpose
The product acts as a structured PR builder. A Change spec defines the PR contract before implementation starts.

## Change Spec
A Change spec records:

- goal
- scope
- requirements
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

The Change slug is valid if and only if it matches `^[0-9]+-[0-9A-Za-z_-]+$`. The `change/`
namespace is not part of the slug. Validation uses that exact expression everywhere, while
extraction from a full branch name uses `^change/([0-9]+-[0-9A-Za-z_-]+)$`.

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
