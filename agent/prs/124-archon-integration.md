# Archon Integration

## Summary
- Add repository-scoped Archon integration for both Codex and Claude through matching workflow and run-management skills.
- Configure Archon to use Codex with `gpt-5.6-sol`, high reasoning effort, the repository documentation directory, and `stage` as the worktree base branch.
- Allow Archon-managed task branches during Change implementation and PR work.

## Changes
- Add the `archon` skill with routing for installation, configuration, repository initialization, workflow execution, workflow and command authoring, and troubleshooting.
- Add workflow examples and references covering DAG node types, conditions, variables, provider capabilities, hooks, MCP servers, skills, retries, sessions, typed artifacts, interactive gates, isolation, and operational guidance.
- Add setup guides for CLI, server, GitHub, Slack, Telegram, and Discord integrations.
- Add the `manage-run` skill and command reference for listing, inspecting, starting, approving, rejecting, abandoning, and resuming Archon workflow runs with JSON output.
- Mirror the Archon and run-management skill packages under `.agents/skills/` and `.claude/skills/`.
- Add the Change definition and initialize its Spec and PR artifacts.

## Testing
- Not run; PR drafting only.
