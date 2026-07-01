Using `agent/changes/109-db-alters-and-views.md` as the source of truth, update or create only the documentation needed to precisely describe the desired external behavior for this Change.

Before editing anything:
1. Read `agent/changes/109-db-alters-and-views.md`.
2. Read `docs/docs-rules.md`.
3. Read every existing `docs/` file that describes affected behavior, including Change fields, backend API payloads, frontend behavior, history behavior, local development, and verification.

Documentation rules:
- Follow `docs/docs-rules.md` exactly.
- Treat the Change file as the contract for this documentation pass.
- Keep docs concise, product-focused, and testable.
- Describe intended external behavior, user-visible/API-visible contracts, persistence constraints, validation behavior, and verification expectations.
- Do not describe implementation internals unless they are part of the observable product or API contract.
- Resolve conflicts between existing docs and the Change file in favor of `agent/changes/109-db-alters-and-views.md`.
- Preserve established project vocabulary.
- Keep each doc under the repository’s documented line limit.
- Do not create duplicate documentation if an existing doc is the right home for the behavior.

Scope:
- Update only files under `docs/`.
- Update all affected docs enough that a future implementer can align code, tests, frontend, CLI, and seed data with the Change without relying on chat history.
- Remove or revise stale references to the old active Change field names when they conflict with the Change.
- Add concise notes for any new external contract introduced by the Change, such as new Change field names, history behavior, API payloads, frontend display behavior, and verification expectations.

Hard constraints:
- Do not implement code.
- Do not edit database files.
- Do not edit backend code.
- Do not edit frontend code.
- Do not edit CLI code.
- Do not edit tests.
- Do not edit seed files.
- Do not edit generated artifacts.
- Do not run migrations or mutate any database.
- Only update documentation under `docs/`.

After editing, report:
- Which docs changed.
- The behavior contracts clarified.
- Any documented follow-up or unresolved ambiguity.
