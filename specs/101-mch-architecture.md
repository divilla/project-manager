# Reference TUI Architecture For `mch`

## Goal

Define the reference architecture and visual baseline for the Go-based `mch` terminal UI, then add a minimal Hello World TUI under `cli/` that proves the architecture, executable name, version, styling approach, and test shape are usable.

## Scope

- Add a reference architecture document under `docs/` for the `mch` terminal UI.
- Document package boundaries for startup, model/state, screens, commands, API clients, config, markdown parsing, editor integration, styling, and tests.
- Document Bubble Tea, Bubbles, and Lip Gloss conventions for model ownership, commands, rendering, reusable components, async work, and long-running AI calls.
- Document the screen/state model for future Change planning flows, including ready, project selection, idea entry, AI running, review, save confirmation, error, and done states.
- Document how `mch` loads, overrides, saves, and validates backend URL and current project context.
- Add a new `cli/` Go module that builds an executable named `mch`.
- Implement a minimal Bubble Tea Hello World TUI using Bubbles where a reusable component is needed and Lip Gloss style tokens from the new reference document.
- Implement `mch --version` output for version `0.1`.
- Add focused tests for the Hello World TUI startup state, version output, one key transition, and rendered output.

## Requirements

- The formal app name is `Make a Change`, but product documentation, UI labels, command examples, requirements, tests, and executable references must use `mch` unless an explicitly approved about/version view requires the formal name.
- The executable name must be `mch`, and the initial version must be `0.1`.
- The reference architecture must identify Bubble Tea, Bubbles, and Lip Gloss as the TUI libraries.
- The reference architecture must keep backend APIs authoritative for Projects, Epics, Changes, reference data, validation, and persistence.
- The TUI must not write application database tables directly.
- The architecture must separate Bubble Tea model state, `Update`, `View`, `tea.Cmd`, API clients, async API calls, and long-running AI calls so future workflows can be added without rewriting the foundation.
- The architecture must define reusable UI components for prompt input, command menu, status/footer, loading indicator, error display, output viewport, confirmation prompt, and project selector.
- The style guide must define named Lip Gloss tokens for background, foreground, muted text, input band, selection highlight, error, success, accent cyan, accent purple, and border colors.
- The style guide must adapt the provided Gemini CLI reference visuals through a dark terminal surface, muted gray input/status bands, compact monospace layout, cyan/purple accents, command palette behavior, muted footer/status metadata, and minimal borders.
- The style guide must not copy Gemini branding, Gemini command names, or proprietary product copy.
- The Hello World TUI must display `mch`, version `0.1`, a Hello World message, and a visible input/status band.
- The Hello World TUI must support a deterministic quit key such as `q` or `ctrl+c`.
- Any code migrated from `cli-proto/` must follow the new package boundaries and style conventions.
- Documentation must include a test strategy covering model transitions, command parsing, API client behavior, markdown parsing, config behavior, and rendering snapshots or golden strings where practical.
- Documentation must include at least one recommended file tree for the future TUI codebase.

## Acceptance Criteria

- A reference architecture document exists under `docs/` and is linked from this Change.
- The document states that `mch` is the required product reference for UI text, command examples, requirements, and tests.
- The document names version `0.1`, executable `mch`, and the formal app name `Make a Change`.
- The document lists Bubble Tea, Bubbles, and Lip Gloss as the TUI libraries.
- The document defines recommended package boundaries for app startup, model/state, screens, commands, API client, config, markdown parsing, editor integration, styling, and tests.
- The document defines how Bubble Tea `Model`, `Update`, `View`, `tea.Cmd`, async API calls, and long-running AI calls are separated.
- The document defines the planning screen/state model for ready, project selection, idea entry, AI running, review, save confirmation, error, and done.
- The document states that backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence.
- The document states that the TUI must not write application database tables directly.
- The document defines backend URL and current project config loading, override, save, and validation behavior.
- The document defines reusable UI components for prompt input, command menu, status/footer, loading indicator, error display, output viewport, confirmation prompt, and project selector.
- The document defines Lip Gloss style tokens for background, foreground, muted text, input band, selection highlight, error, success, accent cyan, accent purple, and border colors.
- The UI style guide includes observable Gemini CLI-inspired layout references: dark background, full-width input band, compact slash-command menu, muted footer/status metadata, and minimal borders.
- The UI style guide avoids Gemini product branding, Gemini command names, and copied product copy.
- A Go module exists at repository-root `cli/`.
- The `cli/` module builds an executable named `mch`.
- The Hello World TUI is implemented with Bubble Tea.
- The Hello World TUI uses Bubbles where a reusable input, viewport, spinner, or related component is needed.
- The Hello World TUI uses Lip Gloss styles that correspond to the documented style tokens.
- Running the app displays `mch`, version `0.1`, and a Hello World message.
- Running `mch --version` prints version `0.1` and exits without starting the interactive TUI.
- Running the app displays a visible input/status band styled according to the reference.
- The app supports a deterministic quit key such as `q` or `ctrl+c`.
- Tests cover startup model state, `--version` output, at least one key transition, and rendered output containing `mch`, version `0.1`, and the Hello World message.
- The app UI does not display `Make a Change` unless an explicitly approved about/version view requires the formal name.
- Any migrated code from `cli-proto/` follows the package boundaries and style conventions defined by the reference architecture.
- The document includes a test strategy covering model transitions, command parsing, API client behavior, markdown parsing, config behavior, and rendering snapshots or golden strings where practical.
- The document includes at least one recommended file tree for the TUI codebase.
- The document identifies follow-up implementation requirements needed after the architecture baseline and Hello World app are approved.

## Non-Goals

- Do not implement a full Change planning workflow beyond the Hello World TUI.
- Do not integrate the Hello World app with backend APIs.
- Do not change database schema or seed data.
- Do not add production packaging or an installer.
- Do not add a Cobra command tree.
- Do not add GitHub integration, PR generation, or PR publishing.
- Do not copy Gemini CLI branding, command names, or proprietary UI text.
- Do not make final decisions for every future planning workflow.
- Do not require Epic assignment for this Change.
- Do not rename user-facing `mch` references to `Make a Change`.

## Design Notes

- `docs/architecture/cli.md` already defines the CLI as an optional automation surface and states that CLI code must not bypass backend rules.
- `docs/functionality/change-lifecycle.md` defines backend-owned Change creation, validation, phase, type, epic, body, and history behavior.
- `docs/functionality/current-project-context.md` defines current project selection behavior that `mch` should align with when project-scoped commands are added.
- Existing `cli-proto/` code is context only; migrate code only when it fits the new reference package boundaries.
- Local visual references are available as `gemini-cli.png` and `mch-01.png`; the implementation should adapt layout patterns without copying Gemini identity.
- The new reference document should stay within the repository documentation rules, including the 300-line limit for Markdown files.

## Relevant Specs

- `docs/architecture/mch.md`
- `docs/architecture/cli.md`
- `docs/functionality/change-lifecycle.md`
- `docs/functionality/current-project-context.md`
- `docs/functionality/agent-interaction.md`
- `docs/operations/verification.md`

## Verification

- From the repository root: `cd cli && go test ./...`
- From the repository root: `cd cli && go build -o ./bin/mch ./cmd/mch`
- From the repository root: `cd cli && ./bin/mch --version`
- From the repository root: `find docs -type f -name '*.md' -not -path 'docs/research/*' -exec wc -l {} +`

## Review Focus

- Confirm the architecture document is concrete enough for future planning workflows but does not implement those workflows in this Change.
- Confirm the Hello World TUI follows the documented package boundaries instead of preserving accidental `cli-proto/` structure.
- Confirm UI copy consistently uses `mch` and avoids Gemini product branding or copied command names.
- Confirm the backend and database boundaries are explicit and enforceable in the design.
- Confirm tests cover the actual model, rendering, version, and quit behavior rather than only package compilation.

## Follow-Ups

- Fixed PR comment `IC_kwDOTA2Xls8AAAABH8PvBw` by moving reusable Lip Gloss style tokens from `internal/app` to `internal/styles` and updating the app model to import the shared styles package.
- Add the full `mch` Change planning workflow after the architecture baseline is approved.
- Add backend API client integration for Projects, Epics, Changes, reference data, validation, and persistence.
- Decide whether `cli-proto/` should be migrated, archived, or removed after `cli/` becomes the active TUI surface.
