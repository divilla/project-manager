# Review Request

Review the current branch against `origin/stage` as a 10x senior engineer.

Change contract: `agent/changes/109-db-alters-and-views.md`

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

Report findings only.

Each finding must be exactly one grouped block of four bullet lines:

- Severity: `P0`, `P1`, `P2`, or `P3`
- File and line: `path/to/file:line`
- Concrete impact: describe the user-visible or system impact
- Specific fix direction: describe the actionable fix

Do not add blank lines between the four bullet lines inside a finding.

Add exactly one blank line between separate findings.

## No Blocking Issues

If there are no blocking issues, say exactly:

`No blocking issues found.`

Example:

- Severity: `P1`
- File and line: `backend/internal/change/repository.go:85`
- Concrete impact: The endpoint returns detail-only fields in list responses, which violates the list contract and can cause clients to cache incomplete detail data.
- Specific fix direction: Return the dedicated list DTO from repository, service, and API layers, and add an API test that asserts list response fields.

- Severity: `P2`
- File and line: `frontend/src/pages/index/changes/[id].vue:43`
- Concrete impact: The UI ignores backend-provided epic names and can show stale linked epic display data.
- Specific fix direction: Add `epic_name` to the frontend Change type and render it before falling back to local epic lookup.
