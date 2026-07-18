# Agent Interaction

## Purpose
Agents help refine planning, maintain documentation, implement scoped changes, and run verification. They operate against the Change spec as the contract.

## Commands
Supported workflow prompts:

- `make-change new <change-slug-or-path>`
- `make-change commit`
- `make-change implement`
- `make-change pr`

Repository Change workflow branches use `change/<change-slug>`. The matching idea and Change spec live at `agent/ideas/<change-slug>.md` and `specs/<change-slug>.md`.

Workflow prompts that operate on an existing Change must fail fast with a clear error when the current branch is not `change/<change-slug>`.

## Change Artifacts
A Change is the full delivery flow. Its artifacts include the branch, idea, spec, docs, code, PR body, PR, review, and follow-up fixes.

Only `IdeaCreate` initializes the Change title from an artifact H1: it parses the canonical Idea H1 and sends it as the explicit create `title`. Every later Idea, Spec, or PR H1 is validation-only and is independent from the Change title and every other artifact title.

Every submitted Idea, Spec, or PR requires a body containing at least one non-whitespace character, then validates and canonicalizes its own optional metadata before persistence. Title-only, whitespace-body, and metadata-only documents are invalid. Types render as `Types: <type-slugs>` with valid slugs joined by `|` and no spaces. Epic renders as `Epic: <epic-title> #<epic-id>` using the canonical title and ID returned for the current project. Each artifact's Types and Epic are independent from the Change and from every other artifact; saving them never updates Change fields.

Only explicit focused Change operations mutate the Change title, type set, or linked Epic. The first Flow assignment allocates `ref` and derives `slug` from the Change title at that time. Later focused Change edits and artifact saves preserve the assigned slug.

The canonical Change spec structure template is `.mch/default/prompts/spec-file-structure.md`. Active workflow prompts and agent instructions must use that path, not legacy prompt locations under `agent/prompts`.

## Planning Behavior
During planning, the agent:

- creates or checks out the matching branch
- commits rough user edits
- rewrites the Change spec into the standard structure
- updates or links relevant docs
- commits the agent checkpoint

## Implementation Behavior
During implementation, the agent:

- reads the current Change spec
- reads referenced docs
- verifies readiness
- changes only files needed for the Change
- records follow-ups instead of silently expanding scope
- runs verification when feasible
- runs `make lint` and fixes all findings after code changes in `backend` or `cli`
- commits with the implementation message

## Autonomy
The agent may edit code, docs, and tests within the active Change. It should stop when a product decision is missing, when docs conflict with requested behavior, or when unrelated worktree changes make the workflow unsafe.

## Database Safety
Agents must treat the repository-root `db` folder as read-only unless the user explicitly requests a specific database-file change. This applies to every file and subfolder under `db`.

Agents must not run PostgreSQL commands that alter database structure, including create, alter, drop, truncate, grant, revoke, migration, or restore operations, unless the user explicitly requests that exact structural change.

When a database function, procedure, schema object, seed file, or backup appears incorrect or blocks implementation, the agent must report the blocker and adapt only application or test code that is within scope.

## Text Quality
Generated Change and documentation text should be grammar-checked, readable, and concise.
