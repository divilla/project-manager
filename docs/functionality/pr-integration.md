# PR Integration

## Purpose
The product acts as a structured PR builder. A Change file defines the PR contract before implementation starts.

## Change File
A Change file records:

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
changes/<change-name>
```

The matching Change file lives at:

```text
agent/changes/<change-name>.md
```

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
The Change file becomes the basis for the PR body. It should be complete enough that reviewers can understand intent, scope, evidence, and review focus without reconstructing context from chat.
