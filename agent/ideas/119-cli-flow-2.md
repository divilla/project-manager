# CLI Flow 2: Idea Stage

The second CLI Flow Change uses the reusable runtime and generic Screens to
implement the complete Idea Stage.

## Idea Flow

- The user writes an initial idea for a software change.
- The agent rewrites the idea for clarity and readability.
- The user previews the rewritten idea and may accept it, edit it, or request
  interactive changes.
- The agent reviews the rewritten idea and reports questions or suggestions.
- If findings exist, the user may answer interactively, edit the idea, or
  cancel.
- When the review is resolved, the user may proceed to Spec writing.

`IdeaCreate` and `IdeaEdit` save their document to the database before starting
`IdeaRewriteExec`. The Rewrite Step then loads the persisted Idea into both
`input.md` and `output.md` through the generic Step lifecycle.

## State Model

| State | Trigger or result | Destination |
|---|---|---|
| `IdeaCreate` | completed | `IdeaRewriteExec` |
| `IdeaEdit` | completed | `IdeaRewriteExec` |
| `IdeaRewriteExec` | expected output | `IdeaRewritePreview` |
| `IdeaRewriteExec` | unexpected output | `IdeaRewriteInteractive` |
| `IdeaRewriteExec` | `/stop` while running | originating navigation state |
| `IdeaRewriteInteractive` | `/interactive` completes | `IdeaRewritePreview` |
| `IdeaRewriteInteractive` | `/edit` | `IdeaEdit` |
| `IdeaRewriteInteractive` | `/cancel` | originating navigation state |
| `IdeaRewritePreview` | `/idea-review` | `IdeaReviewExec` |
| `IdeaRewritePreview` | `/interactive` | `IdeaRewriteInteractive` |
| `IdeaRewritePreview` | `/edit` | `IdeaEdit` |
| `IdeaRewritePreview` | `/cancel` | originating navigation state |
| `IdeaReviewExec` | expected output | `IdeaReviewPreview` |
| `IdeaReviewExec` | unexpected output | `IdeaReviewInteractive` |
| `IdeaReviewExec` | `/stop` while running | originating navigation state |
| `IdeaReviewInteractive` | `/interactive` completes | `IdeaReviewPreview` |
| `IdeaReviewInteractive` | `/edit` | `IdeaEdit` |
| `IdeaReviewInteractive` | `/cancel` | originating navigation state |
| `IdeaReviewPreview` | `/spec-write` | `SpecWrite` |
| `IdeaReviewPreview` | `/interactive` | `IdeaReviewInteractive` |
| `IdeaReviewPreview` | `/edit` | `IdeaEdit` |
| `IdeaReviewPreview` | `/cancel` | originating navigation state |

The originating navigation state is `ChangesListState` or
`ChangeDetailsState`, according to where the Idea flow started.

## Idea Step Definitions

The Idea Stage is composed from a temporary built-in Go definition using the
generic definition types introduced by CLI Flow 1. The runtime must not contain
Idea-specific transition logic.

The built-in definition provides these Exec values:

| Exec State | Artifact | Prompt | Expected last line | Preview next Step |
|---|---|---|---|---|
| `IdeaRewriteExec` | `idea` | `prompts/idea-rewrite.md` | `Done.` | `IdeaReviewExec` |
| `IdeaReviewExec` | `idea` | `prompts/idea-review.md` | `No questions or suggestions.` | `SpecWrite` |

The Rewrite prompt writes the rewritten Idea to `output.md` and finishes its
agent output with `Done.` Each Exec Screen compares its configured expected
output with the last line of the Idea artifact's `agent-output.md`.

## Review Resolution

`IdeaReviewPreview` is reached after either:

- `IdeaReviewExec` produces `No questions or suggestions.`; or
- `IdeaReviewInteractive` completes successfully.

An unexpected Review output opens `IdeaReviewInteractive`, which displays the
review output and waits for the user to choose `/interactive`, `/edit`, or
`/cancel`. Selecting `/interactive` resumes the same Idea session. When Codex
returns control, the Step saves the Idea if it changed and opens
`IdeaReviewPreview`.

`/spec-write` is available only from `IdeaReviewPreview`.

Execution, artifact load, and artifact save errors transition to an Error
State and are never treated as review findings.

## Preview Behavior

Both Idea Preview states use the generic Preview Screen:

- `IdeaRewritePreview` offers `IdeaReviewExec` as its next Step.
- `IdeaReviewPreview` offers `SpecWrite` as its next Step.
- `/interactive` starts the corresponding Interactive-only Step.
- `/edit` opens the Idea in the generic Editor Screen and then returns through
  the Rewrite Step.
- `/cancel` returns to the originating navigation state.

Preview performs no database operations. Each completed Rewrite, Review, or
Interactive Step has already saved a changed `output.md` before Preview is
shown. Starting the next Step causes that Step to load the latest persisted
Idea into a new `input.md` and `output.md` baseline.

The generic runtime and built-in definition should make the complete Idea flow
usable before the project commits the remaining Flow stages to this design.

## Legacy Flow Removal

When the new Idea Stage is connected, remove the existing hardcoded Flow code
that runs after an Idea is created or an edited Idea is saved. The new generic
runtime must replace that path rather than coexist with it.

Keep the existing Idea create, edit, and database-save behavior up to the point
where the save succeeds. After that boundary, hand control to the new Rewrite
Step.

Remove legacy code that exists only to support the old post-save workflow,
including its specialized states, agent model fields, commands, messages,
runner methods, workspace files, views, handlers, transitions, and tests. Do
not leave a compatibility path that can still start the old rewrite or Spec
generation flow.

Code shared with unrelated CLI behavior should remain, but it must no longer
contain branches for the superseded Flow. Tests for the removed behavior should
be replaced with coverage of the new generic Idea Stage rather than retained
as parallel expectations.
