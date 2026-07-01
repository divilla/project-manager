The patch has several contract-level regressions in Change validation, epic clearing, filtered selection behavior, and includes out-of-scope database schema changes. These should be addressed before the patch is considered correct.

Full review comments:

- [P2] Validate markdown before fetching references — /home/vito/go/src/project-manager/cli/internal/app/change_save.go:46-46
  For structurally invalid editor bodies such as a missing H1 or missing `Types:` line, this fetches reference data before running local validation, so a backend/reference failure is shown instead of the required validation error and the flow still calls the backend on inputs that should be rejected locally. Parse the H1/Types structure before loading reference options, then fetch references only when needed for slug/epic resolution.

- [P2] Let omitted Epic clear existing epic links — /home/vito/go/src/project-manager/cli/internal/app/change_save.go:164-164
  When the backend response has `epic_id` but no `epic_name` and the edited markdown omits `Epic:`, this treats the omission as unchanged, so users cannot clear an existing epic by omitting the line even though the Change contract says omitted or blank epic must persist `null`. Resolve the current epic name before edit, or otherwise distinguish a no-op from an explicit omission so the clear request is sent.

- [P2] Clamp the selected row after filtering — /home/vito/go/src/project-manager/cli/internal/changes/view.go:66-66
  If the user has a later row selected and applies a filter that leaves fewer rows, `m.Selected` can remain beyond the filtered row count; `clampOffset` only adjusts the local offset, so no row is highlighted while Enter later selects a different clamped row. Clamp the selected index when filters change or before rendering/selection against the filtered rows.

Before coding:
1. Read and use `agent/changes/107-cli-implement-changes.md` as the source of truth and implementation contract.
2. Read and use the current branch documentation under `docs/` as the behavioral reference.
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
