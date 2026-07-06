# Change Into Spec

## Summary
- Moves active Change workflow spec-generation guidance to the canonical `.mch/default/prompts/spec-file-structure.md` template path.
- Removes the obsolete `agent/prompts/change-file-structure.md` prompt after repointing active prompt, script, docs, spec, and test references.
- Updates workflow wording from ambiguous `<change-name>` placeholders to `<change-slug>` for branch, idea, and spec artifact filenames.

## Behavior
- Active workflow prompts and legacy `agent/prompts` instructions now read the spec structure template from `.mch/default/prompts/spec-file-structure.md`.
- Repository docs describe Change as the parent delivery flow, with branch, idea, spec, docs, code, and PR artifacts sharing the Change title where supported.
- Branch and artifact examples now use `change/<change-slug>`, `agent/ideas/<change-slug>.md`, and `specs/<change-slug>.md`.
- Existing product `Change` terminology, application routes, API paths, and data concepts remain unchanged.

## Docs
- Updates CLI and `mch` architecture docs for repository-root `.mch/default` spec template loading.
- Updates agent interaction, PR integration, and verification docs to use the new slug placeholder and canonical spec template path.
- Refreshes earlier Change specs and PR notes that materially referenced the old spec template path or placeholder wording.

## Verification
- `rg -n "agent/prompts/change-file-structure[.]md|agents/prompts/change-file-structure[.]md" .` returned no matches.
- `rg -n "<change-name[>]" .mch agent docs specs cli backend frontend Makefile` returned no matches.
- `git diff --check origin/stage...HEAD` passed.
- Not run: `(cd cli && make lint)`.
- Not run: `(cd cli && go test ./...)`.
- Not run: `(cd cli && go build -o /tmp/mch ./cmd/mch)`.
- Not run: `(cd backend && GOCACHE=/tmp/project-manager-go-build make test)`.
- Not run: `(cd backend && GOCACHE=/tmp/project-manager-go-build make lint)`.
- Not run: `pnpm --dir frontend test`.
- Not run: `pnpm --dir frontend typecheck`.
- Not run: `pnpm --dir frontend build`.

## References
- `specs/116-spec-file-structure.md`
- `docs/concepts.md`
- `docs/architecture/cli.md`
- `docs/architecture/mch.md`
- `docs/functionality/agent-interaction.md`
- `docs/functionality/pr-integration.md`
- `docs/operations/verification.md`
- `.mch/default/prompts/spec-file-structure.md`
