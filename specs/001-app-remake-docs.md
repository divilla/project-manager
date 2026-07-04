# Application Remake Docs

## Goal
Rewrite the project documentation and agent-facing planning artifacts from the old feature documents into the current `docs` structure, using the change-based product model.

## Scope
- Treat every relevant file in `agent/old/features` as source documentation to convert.
- Rewrite the `docs` folder so it reads as current product documentation, not as historical migration notes.
- Use `docs/docs-rules.md` as the target schema and file organization for the rewritten documentation.
- Update `.agent/AGENTS.md` so repository agent guidance points at the current Change workflow and verification commands.
- Add rough follow-up Change files for backend and frontend implementation work.
- Apply the task-to-change rename throughout the rewritten documentation.
- Apply the `name` to `title` and `description` to `body` field rename throughout the rewritten documentation.
- Document the new epic model with `epic` and `epic_history`.
- Document the fixed non-hierarchical change structure: a change is standalone or references one epic.
- Use `history`, not `smart history`, as the product terminology.
- Remove the old task model from the rewritten documentation.

## Requirements
- The `docs` folder is rewritten from the old documentation currently stored in `agent/old/features`.
- Rewritten documentation follows the schema, file organization, and constraints from `docs/docs-rules.md`.
- Rewritten documentation applies every product, schema, and vocabulary change listed in this Change file.
- Rewritten documentation describes projects, epics, changes, requirements, history, agent interaction, and PR-building behavior.
- Rewritten documentation uses `title` and `body` when describing change fields.
- Rewritten documentation uses `history` terminology and does not use `smart history`.
- Rewritten documentation no longer carries forward the old task hierarchy model.
- Rewritten documentation no longer contains the old task vocabulary.
- Personal research files under `docs/research` are left untouched.
- Repository agent guidance describes the current `make-change` workflow commands and current verification commands.
- Follow-up backend and frontend Change files capture only rough implementation intent for later workflow refinement.
- Documentation remains concise, readable, and suitable as the product source of truth.

## Acceptance Criteria
- `docs` contains the rewritten version of the old `agent/old/features` documentation.
- The rewritten docs are organized according to `docs/docs-rules.md`.
- The rewritten docs describe the current change-based product model without task-based terminology.
- The rewritten docs explain that changes are not hierarchical and may only optionally reference one epic.
- The rewritten docs explain requirement-driven completeness against changes and epics.
- The rewritten docs explain history behavior for user and AI changes.
- The rewritten docs describe PR-building and agent interaction behavior using the Change workflow vocabulary.
- `.agent/AGENTS.md` points agents at `make-change new`, `make-change commit`, `make-change implement`, and `make-change pr`.
- `agent/changes/002-app-remake-backend.md` and `agent/changes/003-app-remake-frontend.md` exist as follow-up Change stubs.
- No documentation file exceeds the line-count limit from `docs/docs-rules.md`.

## Non-Goals
- No backend implementation changes.
- No frontend implementation changes.
- No database schema changes.
- No API route or payload changes.
- No generated migration or seed-data updates.
- No attempt to preserve obsolete task-based documentation as current product docs.
- No refinement of the backend or frontend follow-up Change stubs beyond rough implementation intent.

## Design Notes
- `agent/old/features` is source documentation to convert, not optional reference material.
- `docs` is the target folder and final product documentation location.
- `docs/research` is personal research material and is not part of this rewrite.
- `docs/docs-rules.md` defines the target documentation structure, but terminology in the generated docs must follow this Change file.
- Existing task-based docs should be replaced or rewritten, not lightly patched.
- The output should read as if the product was always change-based.
- Use `history` for audit/revert behavior.
- Follow-up Change stubs remain rough until their own `make-change new` workflow runs.

## Relevant Specs
- `docs/docs-rules.md`
- `.agent/AGENTS.md`
- `agent/changes/002-app-remake-backend.md`
- `agent/changes/003-app-remake-frontend.md`
- `agent/old/features/01-local-application-shell.md`
- `agent/old/features/02-projects-and-tasks.md`
- `agent/old/features/03-requirement-completeness-engine.md`
- `agent/old/features/04-db-migrations-and-seed.md`
- `agent/old/features/05-code-fixes-following-code-refactor.md`
- `agent/old/features/06-frontend-architecture.md`
- `agent/old/features/07-frontend-projects.md`
- `agent/old/features/08-frontend-tasks.md`
- `agent/old/features/09-app-remake.md`
- `agent/old/features/_04-home-dashboard.md`
- `agent/old/features/_05-planning-copilot.md`
- `agent/old/features/_06-help-and-guidance.md`

## Verification
- `find docs -type f | sort`
- `awk 'FNR==1{if(n>300){print f ":" n; bad=1} f=FILENAME; n=0} {n++} END{if(n>300){print f ":" n; bad=1} exit bad}' $(find docs -path 'docs/research' -prune -o -name '*.md' -print | sort)`
- `rg "\btask\b|\bTask\b|task_" docs --glob '!docs/research/**'`
- `rg "smart history|smart-history" docs --glob '!docs/research/**'`

## Review Focus
- Whether the old feature documentation was fully converted into the new docs structure.
- Whether docs are current product documentation rather than migration notes.
- Whether task-to-change vocabulary is complete.
- Whether the epic, requirement, history, agent, and PR-building concepts are clear.
- Whether the rewritten docs are concise and readable.

## Follow-Ups
- Implement backend code changes in `agent/changes/002-app-remake-backend.md`.
- Implement frontend code changes in `agent/changes/003-app-remake-frontend.md`.
