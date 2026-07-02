# CLI Improve Changes and Agentic Workflow

## Summary

- Updates `mch` Changes list chrome: `Make a change v<version>` header, footer shortcut hints, `/filter-*` row aligned to the table, Ctrl+N for new Change, and Ctrl+F for find filter.
- Updates Change detail behavior: Body/PR labels, boolean icons, Complete color, body/test-case divider, Open/test-case done toggles, and backend-refreshed test case create/edit/delete flows.
- Changes selector and confirmation contracts: Change type pending toggles with Space and no-op Return, renamed Change phase/type selector titles, phase option prefixes, `Are you sure?` confirmations, and Ctrl+C as No.
- Extends CLI DTO/client contracts for Change `test_cases`, `update-open`, and test case create/update/update-done/delete endpoints.
- Orders backend Change detail test cases by `test_case.id`; no schema, migration, seed, frontend, or database files are changed.
- Updates `docs/architecture/mch.md`, the Change contract, CLI/backend tests, and PR workflow tooling so `change-pr.pl` uses a prewritten `agent/prs/<change>.md` body file.

## Notes For Reviewers

- Inspect detail-row routing for `<return>`, `<space>`, and `<del>`, especially refresh-after-mutation behavior and recoverable errors.
- Inspect selector pending-state handling for type toggles, no-op Return, cancel, and backend failure paths.
- Inspect `scripts/change-pr.pl` separately from runtime behavior; it now stages/commits/pushes, reads the PR title from the first line of `agent/prs/<change>.md`, creates the PR with `--body-file`, then records and commits the PR URL.

## Verification

- Passed in `backend`: `make lint`
- Passed in `backend`: `make test`
- Passed in `backend`: `make api-test`
- Passed in `cli`: `make lint`
- Passed in `cli`: `go test ./...`
- Passed in `cli`: `go build -o /tmp/mch ./cmd/mch`
- Passed: `git diff --check`

## References

- `agent/changes/110-cli-improve-changes-and-agentic-workflow.md`
- `docs/architecture/mch.md`
- `docs/architecture/backend-api.md`
- `docs/concepts.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/requirements-and-acceptance.md`
- `docs/operations/verification.md`
