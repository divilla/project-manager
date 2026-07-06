# Changes Into Specs To Docs

## Summary
- Moves repository Change specs from `agent/changes/<change-slug>.md` to repo-root `specs/<change-slug>.md` without changing the moved spec contents.
- Renames the active repository workflow branch namespace from `changes/<change-slug>` to `change/<change-slug>` across workflow scripts, agent prompts, repository guidance, and `mch` branch reconciliation.
- Adds `scripts/rename-branches.pl` with dry-run support to migrate existing local and remote `changes/` branches to `change/` branches while creating replacement remote refs before deleting old ones.

## Behavior
- Active workflow scripts now derive the Change name only from `change/<change-slug>` branches, push `change/<change-slug>`, and use `specs/<change-slug>.md` for Change specs.
- `scripts/extract.sh` now returns `specs/<change-slug>.md` and rejects non-`change/` branches with a clear branch-pattern error.
- `mch` `/reference` now checks, checks out, renames, creates, pushes, and deletes `change/<slug>` and `change/<ref>-*` branches instead of `changes/...` branches.
- Product-facing Change terminology remains unchanged: frontend `/changes` routes, backend `/api/v1/change/*` paths, package names, and domain object names are not renamed.

## Docs
- Updates `AGENTS.md`, workflow prompts, PR integration docs, agent interaction docs, verification docs, CLI architecture, and `mch` architecture to describe Change specs under `specs/` and `change/<change-slug>` workflow branches.
- Keeps old `changes/` and `agent/changes/` references only where they are historical, part of the migration helper, or unrelated product route/package names.

## Verification
- Passed: `git diff --name-status origin/stage...HEAD`
- Passed: `rg -n "agent/changes|Change file|changes/<|changes/\\$|changes/" agent docs scripts cli backend frontend`
- Passed: `(cd cli && make lint)`
- Passed: `(cd cli && go test ./...)`
- Passed: `(cd cli && go build -o /tmp/mch ./cmd/mch)`
- Passed: `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`
- Passed: `pnpm --dir frontend test`
- Passed: `pnpm --dir frontend typecheck`

## References
- `specs/113-changes-into-specs-to-docs.md`
- `AGENTS.md`
- `.mch/default/prompts/spec-file-structure.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/pr-integration.md`
- `docs/operations/verification.md`
