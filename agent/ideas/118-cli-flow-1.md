# CLI Flow 1: Reusable Flow Runtime

The first CLI Flow Change establishes the reusable runtime and generic Screens
needed by later Flow stages. It does not compose the Idea Stage yet.

## Runtime Direction

Editor, Exec, Interactive, and Preview Screens must be generic components that
are configured and reused by individual flow states. Flow states must be
configured uses of those Screens rather than separate hardcoded Screen
implementations.

The complete Flow will eventually be loaded from YAML and executed
automatically. Every static parameter that defines a Screen or transition must
therefore be representable in YAML, including:

- Screen and task type
- Artifact
- Prompt or script
- Expected output
- Available commands
- Transition destinations
- Preview next Step
- Other Screen-specific options

Runtime values such as the temporary directory, active Change, session,
originating navigation state, and current execution result belong to the Flow
context rather than the static Screen definition.

## Temporary Built-In Definition

This Change does not implement the complete YAML pipeline loader. Instead, it
introduces typed Flow definition values and constructs the initial definitions
in Go.

The temporary built-in definitions must:

- Use the same types that the future YAML loader will produce.
- Keep all static behavior in one definition instead of scattering constants
  through Screen code.
- Use YAML-representable fields and YAML tags.
- Pass the same validation that future loaded definitions will use.
- Be returned as fresh values rather than mutable package-level state.

The runtime receives a Flow definition without knowing whether it came from Go
or YAML. The future config Change will replace the built-in definition at that
composition boundary without changing the generic Screens or runtime.

## Artifact Context

Every artifact-processing state receives an artifact identifier.

Supported artifact identifiers are:

- `idea`
- `spec`
- `pr`
- `implement`
- `review`
- `finalize`

`implement` represents implementation work. `finalize` represents finalization
work.

Each artifact has an isolated working directory:

    <tmp-dir>/<artifact>/

A state resolves the resources it needs from that directory, including:

    <tmp-dir>/<artifact>/session-id
    <tmp-dir>/<artifact>/input.md
    <tmp-dir>/<artifact>/output.md
    <tmp-dir>/<artifact>/agent-output.md

Missing required resources or unsupported artifact identifiers transition to
an Error State.

## Step Lifecycle

A Step operates on one artifact and may contain one or more tasks. The task
combinations initially supported are:

- Exec
- Exec followed by Interactive
- Interactive

Tasks within the same Step share the same artifact files and runtime context.

### Start

Every Step starts by loading its artifact from the database:

- Write the persisted artifact to `input.md`.
- Write the same persisted artifact to `output.md`.
- Keep `input.md` unchanged for the entire Step.
- Allow the Step's tasks to modify only `output.md`.

An Interactive task that follows an Exec task continues the same Step and does
not reload the artifact.

### Completion

A Step completes successfully when:

- Exec finishes with its expected terminal output; or
- Interactive returns control to the application.

On successful completion:

- Compare `input.md` and `output.md`.
- If they differ, save `output.md` to the artifact's database field.
- If they are identical, do not perform a database write.
- If a save is required, transition to Preview only after it succeeds.
- If no save is required, transition directly to Preview.

Load and save failures transition to an Error State. A stopped, cancelled, or
failed Step does not complete and does not save `output.md`.

## Editor Screens

Editor Screens open `output.md` in the configured external editor and wait for
the editor to return control. Their artifact, commands, and transition
destinations come from the Screen definition.

Editor behavior must remain generic so the same Screen can edit an Idea, Spec,
or PR artifact.

## Exec Screens

All Exec Screens:

- Run the configured prompt using
  `scripts/codex-exec-restore-session.sh`.
- While execution is running, offer only `/stop`.
- Compare the last line of `agent-output.md` with the configured expected
  output after successful execution.
- Complete the Step when the expected output matches.
- Continue the same Step in its configured Interactive Screen when the output
  does not match.
- Return to the originating navigation state when `/stop` terminates the
  process.
- Transition to an Error State when execution fails.
- Supply the configured next Step to Preview.

The lifecycle is:

    Exec -> expected output   -> save if changed -> Preview
    Exec -> unexpected output -> Interactive

## Interactive Screens

An Interactive Screen may follow Exec in the current Step or be the first task
of an Interactive-only Step. Entering the Screen does not start Codex
automatically.

All Interactive Screens:

- Display the preceding agent output when available.
- Offer `/interactive`, `/edit`, and `/cancel`.
- Run `scripts/codex-resume-session.sh` for `/interactive` using the artifact's
  `session-id`.
- Give terminal control to Codex, like an external editor.
- Allow the user to leave Codex using any exit mechanism supported by Codex.
- Complete the Step and transition to Preview after Codex returns control.
- Do not compare Interactive output with an Exec expected value.
- Open `output.md` in the configured Editor Screen for `/edit`.
- Return to the originating navigation state for `/cancel`.
- Transition to an Error State when the session cannot be resolved or Codex
  cannot be started.

Interactive Screens do not offer `/stop`.

## Preview Screens

A Preview Screen is an artifact checkpoint between two Steps:

    previous Step -> Preview Screen -> next Step

Preview initially supports `idea`, `spec`, and `pr`. Other artifacts remain
valid runtime contexts but do not yet have Preview support.

All Preview Screens:

- Perform no database loads or saves.
- Render `output.md` in Preview mode.
- Render the difference between `input.md` and `output.md` in Diff mode.
- Receive the next Step from the completed Step definition.
- Start the next Step when the user continues; that Step performs its own
  database load.
- Switch between Preview and Diff modes when the user presses either the Left
  Arrow or Right Arrow key.

```shell
artifact_dir="$tmp_dir/$artifact"

# Preview mode
bat -pp --theme 'Coldark-Dark' "$artifact_dir/output.md"

# Diff mode
git --no-pager diff \
  --no-index \
  --no-ext-diff \
  --color=never \
  -- "$artifact_dir/input.md" "$artifact_dir/output.md" |
  bat -pp --theme 'Coldark-Dark' --language diff
```

The application interprets `git diff` exit statuses as follows:

- `0`: files are identical
- `1`: files differ
- Greater than `1`: execution failure

In Bash, a pipeline's `$?` normally reports `bat`'s status. Use
`${PIPESTATUS[0]}` or execute Git separately to capture its exit status.
