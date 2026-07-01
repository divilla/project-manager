Implement the code described in Change file `agent/changes/107-cli-implement-changes.md` with senior-level discipline.

Before coding:
1. Read the full Change file and use it as the source of truth and implementation contract.
2. Read all relevant docs under `docs/` and use the current branch documentation as the behavioral reference.
3. Compare the Change file and docs for conflicts.
4. Inspect the existing implementation patterns before choosing an approach.

Hard rules:
- Do not broaden scope beyond the Change file and docs.
- If the Change file and docs conflict, stop and report the exact conflict before coding.
- If any implementation detail is unclear, do not guess or silently choose an approach. Stop and ask one specific clarifying question.
- Preserve unrelated local changes.
- Do not refactor unrelated code.
- Follow existing project architecture, naming, and test patterns.

Implementation requirements:
- Implement only the behavior required by the Change file and docs.
- Add or update focused tests for every changed behavior.
- Keep user-facing behavior observable and aligned with the documented contract.
- Avoid new abstractions unless they clearly reduce complexity or match an existing pattern.

After implementation:
1. Run focused tests for touched behavior.
2. Run the required verification commands for every touched area.
3. Report what changed, which files were touched, and which verification commands passed or failed.
