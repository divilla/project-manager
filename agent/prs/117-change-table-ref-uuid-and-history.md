# Change Table `ref_uuid` and Redesigned History

## Summary
- Introduces generated, unique `ref_uuid` identity for Changes and exposes it as read-only data across backend, frontend, and CLI response models.
- Replaces generic Change history capture with document-specific `idea`, `spec`, and `pr` history owned by the corresponding database procedures.
- Aligns database initialization, demo data, APIs, clients, tests, and documentation around non-null artifact strings and artifact-owned `agent_edit` provenance.

## Behavior
- Change create, list, detail, Flow assignment, and focused mutation responses include a non-null `ref_uuid`; create and update payloads do not accept it.
- Missing optional `spec`, `spec_html`, `pr`, `pr_html`, and `pr_url` values are returned as empty strings instead of null.
- Idea, spec, and PR updates require a non-empty artifact string plus an explicit `agent_edit` boolean; null or empty artifact updates are rejected without mutation.
- PR URL updates require a non-empty absolute `http` or `https` URL.
- Removes the independent `update-idea-agent-edit` and `update-agent-edit` routes. Agent rewrites now use `update-idea` with `agent_edit: true`, while user-driven artifact saves send `agent_edit: false`.

## Data Model
- Adds `change.ref_uuid` with generated UUID values, a non-null constraint, and a unique index.
- Makes `idea`, `spec`, `pr`, and `pr_url` non-null text fields, using empty text for never-submitted optional artifacts.
- Redesigns `change_history` around `id`, `version`, `doc_type`, `body`, `agent_edit`, `modified`, and `deleted`.
- Adds document-specific update procedures that update the active artifact, increment its version, persist provenance, and insert the matching history row.
- Updates initialization and demo seed data for the renamed PR field, non-null artifacts, and document-specific history procedures.

## API and Clients
- Replaces `update-pr-body` and `pr_body` with `update-pr` and `pr` across backend, frontend, and CLI contracts.
- Updates backend DTOs, rendering, repositories, and services to use string artifact fields and reload refreshed Changes after artifact mutations.
- Updates frontend detail/edit behavior to render empty artifacts safely, reject clearing submitted artifacts, validate PR URLs, and send user edits with `agent_edit: false`.
- Updates CLI request builders and agent rewrite saves for the consolidated artifact endpoints and string response contract.
- Expands backend API, service, frontend component, and CLI client coverage for UUID propagation, artifact validation, removed routes, provenance payloads, and repeated Epic/Test Case version updates.

## Verification
- Passed: `! rg -n "update-idea-agent-edit|update-agent-edit|sp_change_to_history|ChangeUpdateIdeaAgentEditRequest" backend frontend cli docs`.
- Not run during PR drafting: backend `make lint`, `make test`, `make api-test`, and `make race`; frontend `pnpm --dir frontend test`, `pnpm --dir frontend typecheck`, and `pnpm --dir frontend build`; CLI `make lint`, `go test ./...`, and `go build -o /tmp/mch ./cmd/mch`.

## References
- `specs/117-change-table-ref-uuid-and-history.md`
- `docs/concepts.md`
- `docs/architecture/backend-api.md`
- `docs/architecture/frontend-spa.md`
- `docs/architecture/mch.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/history.md`
- `docs/operations/verification.md`
