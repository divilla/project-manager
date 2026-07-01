Turn the initial ideas in `agent/changes/109-db-alters-and-views.md` into a complete, implementation-ready Change specification.

Use `agent/prompts/change-file-structure.md` as the exact required structure. Preserve every heading, heading order, and section name from that template. The output must be the full final Markdown content for `agent/changes/109-db-alters-and-views.md`, not a summary or plan.

Before drafting the Change, read `agent/prompts/change-file-structure.md`, `agent/changes/109-db-alters-and-views.md`, and all relevant files under `docs/`. Use the project’s documented vocabulary and contracts for Change workflow, backend API
behavior, CLI behavior, persistence, verification, and QA. Treat documentation as the source of truth. If the current draft conflicts with docs, resolve the conflict in favor of docs and record any important assumption in `Design Notes`.

Clarification policy:
- Ask clarifying questions before producing the final Change whenever any requirement, scope boundary, API contract, database contract, CLI behavior, verification expectation, or QA expectation is ambiguous.
- Do not guess missing product decisions.
- Do not silently fill gaps with assumptions unless the documentation directly supports them.
- If clarification is needed, stop after asking the questions and do not draft the Change yet.

Scope control:
- Keep the Change scoped to one coherent outcome.
- Convert vague notes, placeholders, and brainstorming text into concise product requirements.
- Move related but non-essential ideas into `Non-Goals` or `Follow-Ups`.
- Do not expand scope to make the Change feel more complete.

Requirement quality:
- `Requirements` must describe required behavior, user-visible/API-visible contracts, persistence expectations, and boundaries.
- `Acceptance Criteria` must describe observable success conditions that a reviewer or QA tester can verify.
- Avoid implementation tasks disguised as product requirements unless the implementation detail is explicitly part of the project contract.
- Include database expectations only as product/persistence contract. Do not instruct the agent to mutate any database.

Verification and QA:
- Include realistic verification commands for every affected area.
- Do not invent verification commands that this repository cannot run.
- Include QA Test Cases covering:
    - happy paths
    - validation failures
    - backend or command failures
    - cancellation or no-op paths
    - persistence behavior
    - important boundary cases

Hard constraints:
- Do not implement code.
- Do not edit tests.
- Do not change docs outside the Change file.
- Do not run migrations.
- Do not mutate any database.
- Only produce the final Change specification content, or clarifying questions if required.
