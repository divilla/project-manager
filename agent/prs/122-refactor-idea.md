# Refactor Artifact Terminology and Retain Initial Flow Configuration

## Summary

- Rename the Change definition artifact from `idea` to `def` in machine-facing contracts and to
  `definition` in user-facing text across the database, backend, frontend, CLI, scripts, prompts,
  fixtures, and tests.
- Rename the shared definition, Spec, and PR write stage from `idea` to `artifact`, including CLI
  workspaces, `MCH_STAGE`, Make targets, and default Flow configuration.
- Move the definition archive from `agent/ideas` to `agent/defs` while preserving historical
  artifact contents, and update current decisions and active prompts to the new terminology.

## Behavior

- Replace the backend `idea` field and `POST /api/v1/change/update-idea` route with `def` and
  `POST /api/v1/change/update-def`; the old field and route are not retained as compatibility
  aliases.
- Display and edit definitions as `Definition` in the frontend and CLI while preserving existing
  trimming, validation, type parsing, provenance, save ordering, cancellation, and failure
  behavior.
- Run `def-write`, `def-review`, `spec-write`, `pr-write`, and `artifact-chat` in the shared
  `.mch/tmp/<ref_uuid>/artifact` workspace with `MCH_STAGE=artifact` and reusable saved sessions.
- Initialize new and existing Change branches without publishing them, and create the definition,
  Spec, and PR artifact files under `agent/defs`, `specs`, and `agent/prs`.

## Data Model

- Rename `public.change.idea` to `def`, expose it through the Change view, and update backend
  scanners and test-case projections to consume the renamed column.
- Store initial and updated definition history with `doc_type='def'`, using `_def` in
  `public.fn_change_insert` and `public.sp_change_def_update`.
- Keep `def`, `spec`, and `pr` as writable artifact types while using `artifact` as their shared
  Flow stage.

## Flow and Scripts

- Retain the expanded default Flow graph, prompt paths, routing metadata, hooks, and output-switch
  declarations as configuration-only data; this Change does not add runtime support for those
  unfinished fields or resources.
- Remove the legacy prompt-selector script and generated `*-prompt` Make targets so configured Flow
  prompt paths remain the prompt-selection contract.
- Rename `scripts/change-idea.pl` to `scripts/change-def.pl` and preserve its explicit commit-and-push
  workflow separately from local-only Change initialization.

## Testing

- Add and update backend service and API coverage for definition creation, trimming, updates,
  history, provenance, validation, missing records, and rejection of the old API contract.
- Add frontend coverage for definition API payloads and create, edit, detail, validation, and
  failure behavior.
- Add CLI model, client, program integration, PTY, Flow configuration, and script integration
  coverage for renamed operations, shared artifact workspaces, session reuse, cancellation,
  failures, local-only initialization, and definition terminology.

## Verification

- `git diff --check origin/stage...HEAD` — passed.
- Not run: the Spec-listed backend, CLI, and frontend verification suites were not executed during
  PR drafting, and no exact successful results were available in the branch or conversation.

## Risks

- The database column and API field/route renames are intentionally breaking and provide no legacy
  `idea` compatibility path.
- Several expanded Flow declarations reference future resources and behavior; they remain inert
  configuration until a later Change implements the generic Flow engine.

## Docs

- Update the code-first documentation decision to describe definitions as the source artifact.
- Update the Flow stage ownership decision to document the shared `artifact` workspace.

## References

- `specs/122-refactor-idea.md`
- `docs/decisions/code-first-documentation.md`
- `docs/decisions/flow-stage-ownership.md`
