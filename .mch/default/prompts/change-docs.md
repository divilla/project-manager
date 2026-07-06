---
name: change-docs
description: Update repository documentation under docs from the active agent changes file on a changes branch, without editing code, tests, database files, or generated artifacts.
---

Extract `<change-slug>` from the current Git branch using this regex: `^change/([0-9]+-[a-z0-9-_]+)$`

If the current branch does not match, stop and output one concise error explaining that the branch must be named `change/<change-slug>`.

Stop without editing if `specs/<change-slug>.md` does not exist.

Using Change spec `specs/<change-slug>.md` as the source of truth, update or create only the documentation needed to precisely describe the desired external behavior for this Change.

Before editing anything:
1. Read the Change spec.
2. Read `docs/docs-rules.md`.
3. Use targeted search to find related `docs/` files that describe affected behavior.
4. Read only docs that materially affect this Change, including relevant Change fields, backend API payloads, frontend behavior, history behavior, local development, and verification.

Documentation rules:
- Follow `docs/docs-rules.md` exactly.
- Treat the Change spec as the contract for this documentation pass.
- Keep docs concise, product-focused, and testable.
- Describe intended external behavior, user-visible/API-visible contracts, persistence constraints, validation behavior, and verification expectations.
- Do not describe implementation internals unless they are part of the observable product or API contract.
- Resolve conflicts between existing docs and the Change spec in favor of `specs/<change-slug>.md`.
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
