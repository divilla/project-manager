# Review Request

Review the current branch against `origin/stage` as a 10x senior engineer.

Change contract: `agent/changes/108-cli-improve-changes-and-agentic-workflow.md`

## Review Steps

1. Read and use the Change contract as the source of truth and implementation contract.
2. Read and use the relevant branch documentation under `docs/` as the behavioral reference.
3. Inspect the full diff against `origin/stage`.
4. Verify behavior against the Change file and current docs, not assumptions.
5. Pay close attention to regressions from the latest fixes.

## Review Rules

- Treat the Change file as the PR contract.
- Treat current `docs/` changes as the behavioral reference.
- Report only actionable findings.
- Prioritize correctness bugs, data loss risks, duplicate writes, broken state transitions, contract violations, missing required tests, and user-visible regressions.
- Do not report style nits, preferences, unrelated cleanup, or scope expansions.
- Do not broaden the Change scope.
- If behavior is ambiguous, ask a question instead of inventing a requirement.

## Findings Format

For each finding, include:

- Severity: `P0`, `P1`, `P2`, or `P3`
- File and line
- Concrete impact
- Specific fix direction

Do not add blank lines between bullet points within a finding. Add one blank line between separate findings.

## No Blocking Issues

If there are no blocking issues, say exactly:

`No blocking issues found.`
