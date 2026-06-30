# Reference TUI Architecture For `mch`

Types: feature|docs|spike|test

## Problem Statement

Developers need a clear architecture and visual design baseline before expanding `mch` into a Go-based terminal UI for planning Changes. The app’s formal name is `Make a Change`, but product documentation, UI text, commands, and requirements must refer to it as `mch`. The baseline must define how the `mch` executable, version `0.1`, should structure Bubble Tea models, commands, API clients, config, state transitions, styling, and tests so later requirements can add workflows without reworking the foundation.

A minimal Hello World TUI must also be built in the `cli` folder using the proposed reference architecture. Useful code from `cli-proto/` may be migrated only when it fits the reference architecture.

The TUI should visually align with the Gemini CLI reference style where practical: dark terminal surface, muted gray input/status bands, compact monospace layout, cyan/purple accents, command palette behavior, and footer/status metadata. This must be a product-specific adaptation, not a copy of Gemini branding or product text.

## Primary Workflows

1. An engineer reads the architecture reference under the existing docs path and can identify where to add a new planning workflow, API integration, screen state, reusable component, style token, and test.
2. An engineer reads the UI reference and can reproduce the intended `mch` look using Bubble Tea, Bubbles, and Lip Gloss.
3. An engineer runs the Hello World TUI from the `cli` folder and sees a working Bubble Tea application using the reference architecture.
4. An engineer runs `mch --version` and sees version `0.1`.
5. An engineer can compare the current `cli-proto` implementation against the reference architecture and migrate only compatible code.
6. Future requirements can cite the architecture document and Hello World app for app naming, executable naming, versioning, package boundaries, command handling, state ownership, and rendering conventions.

## Acceptance Criteria

1. A reference architecture document exists under the existing `docs/` path.
2. The document states that the app’s formal name is `Make a Change`, but all product references, UI labels, command examples, and requirement text use `mch`.
3. The document names version `0.1` and executable `mch`.
4. The document explicitly lists Bubble Tea, Bubbles, and Lip Gloss as the TUI libraries.
5. The document defines recommended package boundaries for app startup, model/state, screens, commands, API client, config, markdown parsing, editor integration, styling, and tests.
6. The document defines how Bubble Tea `Model`, `Update`, `View`, `tea.Cmd`, async API calls, and long-running AI calls should be separated.
7. The document defines a screen/state model for planning Changes, including ready, project selection, idea entry, AI running, review, save confirmation, error, and done states.
8. The document defines how backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence.
9. The document states that the TUI must not write application database tables directly.
10. The document defines how current project context and backend URL config should be loaded, overridden, saved, and validated.
11. The document defines reusable UI components for prompt input, command menu, status/footer, loading indicator, error display, output viewport, confirmation prompt, and project selector.
12. The document defines a Lip Gloss style guide with named style tokens for background, foreground, muted text, input band, selection highlight, error, success, accent cyan, accent purple, and border colors.
13. The UI style guide includes observable references to the Gemini CLI-inspired layout: dark background, full-width input band, compact slash-command menu, muted footer/status metadata, and minimal borders.
14. The UI style guide avoids Gemini product branding, Gemini command names, and any copied product copy.
15. A Go module exists under repository-root `cli/`.
16. The `cli/` module builds an executable named `mch`.
17. The Hello World TUI is implemented with Bubble Tea.
18. The Hello World TUI uses Bubbles where a reusable input, viewport, spinner, or related component is needed.
19. The Hello World TUI uses Lip Gloss styles from the documented style tokens.
20. Running the app displays `mch`, version `0.1`, and a Hello World message.
21. Running `mch --version` prints version `0.1` and exits without starting the interactive TUI.
22. Running the app displays a visible input/status band styled according to the Gemini CLI-inspired reference.
23. The app supports a deterministic quit key, such as `q` or `ctrl+c`.
24. The app has tests for startup model state, `--version` output, at least one key transition, and rendered output containing `mch`, version `0.1`, and the Hello World message.
25. The app UI does not display `Make a Change` unless an explicitly approved about/version view requires the formal name.
26. Any migrated code from `cli-proto/` follows the package boundaries and style conventions defined by the reference architecture.
27. The document includes a test strategy covering model transitions, command parsing, API client behavior, markdown parsing, config behavior, and rendering snapshots or golden strings where practical.
28. The document includes at least one recommended file tree for the TUI codebase.
29. The document identifies follow-up implementation requirements needed after the architecture baseline and Hello World app are approved.

## Edge Cases

1. Terminal width is too narrow for the command menu or status footer.
2. Terminal height is too short to show output, input, and footer at the same time.
3. Terminal does not support full color or uses a light theme.
4. Backend URL is missing, malformed, or unreachable.
5. Current project config points to a deleted project.
6. Backend reference data cannot be loaded.
7. AI command runs for a long time, fails, or returns output that is not valid requirement markdown.
8. Slash command is unknown or entered during a state where it is not allowed.
9. User cancels during planning, review, confirmation, or save.
10. UI rendering changes cause text overlap, clipped input, or unreadable selection states.
11. Gemini CLI reference visuals change upstream after this requirement is written.
12. `cli/` cannot build because Go dependencies are missing or incompatible.
13. `mch` is run in a terminal that does not support interactive TUI behavior.
14. `mch --version` is run from a non-interactive shell or script.
15. Code, docs, or UI copy accidentally refer to the app as `cli` or `Make a Change` where `mch` is required.

## Non-Goals

1. No full Change planning workflow is required beyond the Hello World TUI.
2. No backend API integration is required in the Hello World app.
3. No database schema changes are required.
4. No production packaging or installer is required.
5. No Cobra command tree is required.
6. No GitHub integration is required.
7. No PR generation or publishing is required.
8. No copying of Gemini CLI branding, names, or proprietary UI text is allowed.
9. No final decision is made here about every future planning workflow.
10. No Epic assignment is required for this Change.
11. No user-facing rename from `mch` to `Make a Change` is included.

## Dependencies And Risks

1. Depends on Go, Bubble Tea, Bubbles, and Lip Gloss.
2. Depends on existing Project Manager backend API contracts for Projects, Epics, Changes, and reference data.
3. Depends on repository CLI guidance in `docs/architecture/cli.md`.
4. Depends on the existing `cli-proto` implementation as current context, not necessarily as the final architecture.
5. Uses Gemini CLI as visual inspiration: https://github.com/google-gemini/gemini-cli
6. Local screenshot references exist as `gemini-cli.png` and `mch-01.png`.
7. The local sandbox could not call `POST http://localhost:8080/api/v1/change/reference`; selected type slugs should be verified against the live endpoint before saving.
8. Mimicking Gemini CLI too closely could create UX confusion or branding risk, so the design should adapt patterns rather than clone identity.
9. Creating a new `cli/` folder while `cli-proto/` already exists may create duplicate TUI surfaces unless follow-up requirements define migration or retirement behavior.
10. Migrating `cli-proto/` code may preserve accidental prototype constraints unless each migrated piece is checked against the reference architecture.
11. The formal name `Make a Change` and required user-facing reference `mch` may create inconsistent copy unless naming rules are tested in docs and UI output.
