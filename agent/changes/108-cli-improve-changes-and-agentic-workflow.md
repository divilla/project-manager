# Improve CLI Changes Section and Agentic Workflow

## Goal

Make the Change workflow documentation, agent prompts, `mch` Change screens, and git automation scripts clearer and more consistent so agents and developers can operate the workflow with less ambiguity.

## Scope

- Update the documented `mch` Change behavior for create, update, list, detail, and filter flows.
- Implement the documented `mch` Change behavior for backend-backed refreshes, focused detail edits, selectors, filters, and requirement markdown parsing.
- Improve Change workflow prompts so agents produce complete, reviewable Change files and senior-engineer reviews.
- Improve git automation scripts for Change build, code, docs, fix, and master promotion steps.
- Clarify AGENTS.md Change workflow instructions where explicitly approved.

## Requirements

- `mch` must use backend-provided `ref` and `slug` as read-only Change identity and must not prompt for, submit, or derive those values locally.
- `mch` must reload backend-backed Change list and detail screens from backend APIs when entering those screens and after successful create, update, delete, or focused field update actions.
- `mch` Change create and full edit flows must open the external editor, parse strict requirement markdown metadata, validate title and type slugs, resolve optional epics, preserve the full markdown body, and call only the backend endpoints required for changed fields.
- `mch` Change detail must render the documented row order and support focused edits for phase, epic, types, title, requirement body, and pull request body.
- `mch` Change filters must remain list-local, load backend reference data, render filter options according to the documented option format, support field-specific `/clear`, and support find filtering across loaded Change fields.
- Change workflow prompt files must preserve the standard Change structure and guide agents to produce testable requirements, acceptance criteria, verification, QA cases, and review focus.
- Git automation scripts must validate branch context, use clear usage messages, commit scoped workflow steps, and protect master promotion from stale or divergent refs.

## Acceptance Criteria

- `/changes` loads `POST /api/v1/change/list` for the current numeric project ID and renders backend rows in response order with documented columns, formatting, and filters.
- Selecting a Change list row reloads details through `POST /api/v1/change/get` before rendering `ChangeDetailsState`.
- `/new-change` and `/edit` use external-editor markdown flows, validate metadata before mutation, and preserve `requirement_body`.
- Focused Change detail edits for phase, epic, types, title, requirement body, and pull request body call the matching backend update endpoint, reload the Change through `POST /api/v1/change/get`, and keep the selected detail row.
- `/phase`, `/epic`, and `/types` from `ChangeDetailsState` use the same focused-save behavior as pressing Enter on those detail rows.
- Phase, epic, and type filter dropdowns render normal options with a leading `-` and render `/clear` as the final clear action.
- Prompt files reference the active Change file and the standard Change file structure.
- Script usage strings and commit messages match the workflow step they automate.
- `scripts/change-master.pl` verifies a clean `stage` branch, synchronized remote refs, and master fast-forward safety before promoting `origin/master`.

## Non-Goals

- Do not add new backend database schema, migrations, or foreign keys.
- Do not implement new non-interactive `mch` automation commands.
- Do not expand Epics, Projects, or Test Cases beyond the workflow consistency changes needed by this Change.
- Do not replace the existing Bubble Tea application architecture.

## Design Notes

- `docs/architecture/mch.md` is the behavioral source of truth for the TUI contract.
- Backend APIs remain authoritative for persistence and validation; `mch` must not write database tables directly.
- Change create and update use backend type slugs and current-project epic lookup instead of local reference derivation.
- Focused selector saves reuse the same backend update and reload path whether opened from slash commands or detail-row selection.
- The Change workflow scripts intentionally operate through Git commands and should fail fast on branch or ref safety violations.

## Relevant Specs

- `agent/changes/108-cli-improve-changes-and-agentic-workflow.md`
- `docs/architecture/mch.md`
- `docs/architecture/cli.md`
- `docs/architecture/backend-api.md`
- `docs/functionality/change-lifecycle.md`
- `docs/operations/verification.md`

## Verification

- `GOCACHE=/tmp/project-manager-go-build go test ./...` from `cli`
- `make lint` from `cli`
- `go build -o /tmp/mch ./cmd/mch` from `cli`

## QA Test Cases

- Open `/changes`, verify list rows load for the current project, apply each filter type, use `/clear`, use `/clear-filters`, and verify no-match filtering renders a no-results state.
- Open a Change detail from the list and verify the detail row order, scroll behavior, read-only identity fields, and formatted timestamps.
- Use `/phase`, `/epic`, and `/types` from Change details and verify each selection persists, reloads details, and keeps the focused row.
- Press Enter on Phase, Epic, Types, Title, Requirement, and Pull Request detail rows and verify each focused edit persists through the expected backend endpoint and reloads details.
- Create a Change with valid requirement markdown and verify the created Change detail opens with backend `ref` and `slug`.
- Attempt create or update with missing title, missing types, invalid type, and unknown epic and verify no mutation endpoint is called.
- Cancel Change create, full edit, focused title edit, focused selector edit, and editor-backed text edits and verify no persistence call occurs.
- Run the Change workflow scripts from valid and invalid branches and verify usage, branch validation, and failure messages.

## Review Focus

- Focused Change edit paths, especially command-opened selectors versus detail-row Enter flows.
- Requirement markdown parsing and selective backend endpoint calls.
- Backend reload behavior after mutations and stale response handling.
- Filter option rendering and selection state clamping.
- Git automation safety checks for branch names, clean worktrees, stale refs, and master promotion.

## Follow-Ups

- Add broader end-to-end TUI coverage when a stable terminal interaction harness is available.
- Review fixes applied: focused title `/cancel` avoids mutation, and confirmed Change delete now persists before list reload.
