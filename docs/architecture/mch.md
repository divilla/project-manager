# `mch` TUI Architecture

## Purpose

`mch` is the Go terminal UI for planning Changes. The formal app name is `Make a Change`, but product documentation, UI labels, command examples, requirements, tests, and executable references use `mch` unless an approved about or version view explicitly needs the formal name.

The first executable version is `0.1`. The executable name is `mch`.

## Libraries

`mch` uses:

- Bubble Tea for the application loop, model updates, messages, and commands
- Bubbles for reusable terminal controls such as text input, viewport, spinner, and list behavior
- Lip Gloss for rendering styles and layout tokens

## Package Boundaries

Recommended layout:

```text
make-change/
  cmd/mch/
  internal/app/
  internal/api/
  internal/commands/
  internal/config/
  internal/editor/
  internal/markdown/
  internal/screens/
  internal/styles/
  internal/testutil/
```

Responsibilities:

- `cmd/mch`: parse process arguments only far enough to call the app runner and set exit status.
- `internal/app`: own startup wiring, top-level Bubble Tea model, version output, and command-line flags.
- `internal/screens`: keep screen-specific state transitions and view helpers.
- `internal/commands`: parse slash commands and create typed `tea.Cmd` functions for async work.
- `internal/api`: call backend APIs for Projects, Epics, Changes, reference data, validation, and persistence.
- `internal/config`: load, validate, override, and save local backend URL and current project context.
- `internal/markdown`: parse and validate generated requirement and Change markdown.
- `internal/editor`: isolate external editor launches and file handoff.
- `internal/styles`: define Lip Gloss style tokens and shared components.
- `internal/testutil`: provide test messages, terminal widths, fixtures, and render assertions.

## Model And Commands

The root Bubble Tea `Model` owns current screen, window size, command menu state, current project context, visible errors, and reusable component models. It should delegate screen-specific decisions to focused helpers rather than embedding full workflows in one method.

`Update` should only translate messages into state changes and `tea.Cmd` values. It should not perform HTTP requests, file writes, editor launches, or AI calls directly.

`tea.Cmd` functions should wrap asynchronous work and return typed messages. Backend API calls and long-running AI calls must be cancellable through `context.Context` where possible. A running AI call should update the UI through loading messages and then return either a structured result message or an error message.

`View` should render current state from model data only. Rendering must not mutate state, read files, call APIs, or start processes.

## Planning States

Future Change planning flows should use these states:

- `ready`: project context is valid and the app is ready for a planning command.
- `project selection`: no current project is selected or the saved project is invalid.
- `idea entry`: the user is entering or refining a Change idea.
- `AI running`: an async AI command is active and progress metadata is visible.
- `review`: generated requirement markdown is available for review.
- `save confirmation`: parsed output is ready to persist through backend APIs.
- `error`: recoverable failure with a visible reason and next action.
- `done`: the planned Change has been saved or the flow has exited cleanly.

Slash commands should be accepted only in states that define them. Unknown commands should leave user input intact and show a recoverable error.

## Backend And Persistence

Backend APIs remain authoritative for Projects, Epics, Changes, reference data, validation, and persistence. `mch` must not write application database tables directly.

Project-scoped commands should either use the saved current project context or require an explicit project option. When the saved project no longer exists, `mch` should clear or repair selection using the same behavior documented for current project context.

## Config

`mch` should load local config at startup, then apply command-line overrides such as backend URL for the current process. Config validation should reject missing or malformed backend URLs before project-scoped API calls.

Saving config should be explicit and limited to local CLI state, such as backend URL and current project ID. Product data must be saved only through backend APIs.

## Components

Reusable components should cover:

- prompt input
- command menu
- status/footer
- loading indicator
- error display
- output viewport
- confirmation prompt
- project selector

Components should accept width and state as inputs so narrow terminals do not produce overlapping text. When width is too small, content should truncate or stack before it clips important state.

## Style Tokens

The baseline style uses a dark terminal surface, full-width muted input band, compact monospace layout, cyan and purple accents, muted footer/status metadata, and minimal borders. This adapts the local Gemini CLI reference screenshots without copying Gemini branding, command names, or product copy.

Named Lip Gloss tokens:

- `Background`: dark terminal background
- `Foreground`: primary readable text
- `Muted`: secondary metadata text
- `InputBand`: full-width prompt and status band
- `Selection`: highlighted command or project selection
- `Error`: recoverable error text
- `Success`: completion text
- `AccentCyan`: primary interactive accent
- `AccentPurple`: secondary accent
- `Border`: low-contrast border color

UI text must remain product-specific to `mch`.

## Test Strategy

Model tests should cover startup state, screen transitions, command parsing, async message handling, and cancellation paths.

Rendering tests should assert stable output for important strings, status bands, narrow widths, and no accidental `Make a Change` copy in regular UI.

API client tests should use HTTP test servers and must not inspect database tables directly.

Markdown parsing tests should cover valid generated requirements, invalid markdown, missing titles, unsupported type values, and editor round trips.

Config tests should cover missing files, malformed files, command-line overrides, saved backend URL, saved project ID, and invalid saved project repair.

## Follow-Up Work

After the Hello World baseline, add the real Change planning workflow, backend API integration, markdown validation, editor handoff, and a retirement plan for `cli-proto/`.
