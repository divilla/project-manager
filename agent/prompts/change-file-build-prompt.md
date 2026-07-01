Turn the initial ideas in `agent/changes/108-cli-improve-changes-and-agentic-workflow.md` into a complete, implementation-ready Change specification.

Use `agent/prompts/change-file-structure.md` as the exact required structure. Preserve every heading, heading order, and section name from that template. Replace vague notes, placeholders, and brainstorming text with concise product requirements.

Before writing the final Change, read the relevant `docs` files and use the project's established vocabulary for Change workflow, backend API, CLI behavior, persistence, and verification. Treat documentation as the source of truth. If the draft conflicts with docs, resolve the conflict in favor of docs and record any important assumption in Design Notes.

Ask clarifying questions only when missing information would make the Change untestable or could produce the wrong product behavior. If the intended outcome is clear enough, make conservative assumptions and state those assumptions in Design Notes instead of stopping.

Keep the Change scoped to one outcome. Requirements must describe required behavior and boundaries. Acceptance Criteria must describe observable success conditions that a reviewer or QA tester can verify. Avoid implementation tasks disguised as product requirements unless the implementation detail is part of the contract.

Include realistic Verification commands for every affected area. Include QA Test Cases that cover happy paths, validation failures, backend or command failures, cancellation or no-op paths, persistence behavior, and any important boundary cases. Do not invent verification that the repository cannot run.

Move related but non-essential ideas into Non-Goals or Follow-Ups. Do not expand scope to make the Change feel complete.

Do not implement code, edit tests, change docs outside the Change file, run migrations, or mutate any database. Only produce the final Change specification content.
