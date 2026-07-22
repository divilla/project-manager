Implement all the review findings like a 10x senior engineer.

Change contract: `specs/<change-slug>.md`

Before coding:
1. Read the Change spec as the intended scope and inspect code as the source of current behavior.
2. Read each review comment carefully.
3. Read a retained ADR or operational runbook only when the Spec or finding explicitly cites it.
4. Map every review comment to the exact expected behavior and affected files.

Documentation is outside the default Change Flow. Do not inspect, create, update, or reconcile
documentation unless the user explicitly requests documentation work or the active Spec explicitly
includes it.

Rules:
- Implement only the requested review fixes.
- Do not broaden scope beyond the Change spec and review comments.
- If any implementation detail is unclear, do not guess or silently choose an approach. Stop and ask one specific clarifying question.
- Preserve unrelated local changes.
- Do not refactor unrelated code.
- Add or update focused tests for the fixed behavior.
- Keep behavior aligned with existing project patterns and vocabulary.

After implementation:
1. Run the relevant focused tests.
2. Run the required verification commands for every touched area.
3. Update the Change spec’s `Follow-Ups` section with a concise note about the review fixes applied.
4. Report which review comments were addressed, what changed, and which verification commands passed or failed.
