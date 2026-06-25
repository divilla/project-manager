# Agent Interaction

## Purpose
Agents help refine planning, maintain documentation, implement scoped changes, and run verification. They operate against the Change file as the contract.

## Commands
Supported workflow prompts:

- `new change <change-name-or-path>`
- `commit change`
- `implement change`

## Planning Behavior
During planning, the agent:

- creates or checks out the matching branch
- commits rough user edits
- rewrites the Change file into the standard structure
- updates or links relevant docs
- commits the agent checkpoint

## Implementation Behavior
During implementation, the agent:

- reads the current Change file
- reads referenced docs
- verifies readiness
- changes only files needed for the Change
- records follow-ups instead of silently expanding scope
- runs verification when feasible
- commits with the implementation message

## Autonomy
The agent may edit code, docs, and tests within the active Change. It should stop when a product decision is missing, when docs conflict with requested behavior, or when unrelated worktree changes make the workflow unsafe.

## Text Quality
Generated Change and documentation text should be grammar-checked, readable, and concise.
