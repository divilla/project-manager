Implement all the review findings like a 10x senior engineer.

Change contract: `agent/changes/108-cli-improve-changes-and-agentic-workflow.md`

Before coding:
1. Read and use Change contract as the source of truth and implementation contract.
2. Read and use the relevant branch documentation under `docs/` as the behavioral reference.
3. Read each review comment carefully.
4. Map every review comment to the exact expected behavior and affected files.

Rules:
- Implement only the requested review fixes.
- Do not broaden scope beyond the Change file, docs, and review comments.
- If the docs and Change file conflict, stop and report the conflict before coding.
- If any implementation detail is unclear, do not guess or silently choose an approach. Stop and ask one specific clarifying question.
- Preserve unrelated local changes.
- Do not refactor unrelated code.
- Add or update focused tests for the fixed behavior.
- Keep behavior aligned with existing project patterns and vocabulary.

After implementation:
1. Run the relevant focused tests.
2. Run the required verification commands for every touched area.
3. Update the Change file’s `Follow-Ups` section with a concise note about the review fixes applied.
4. Report which review comments were addressed, what changed, and which verification commands passed or failed.
