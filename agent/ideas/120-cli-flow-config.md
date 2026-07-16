# CLI Flow Config

Replace the temporary built-in Flow definitions with a YAML-configured pipeline
that constructs and runs the same generic Screens and Step runtime.

The working Idea Stage from CLI Flow 2 is the reference behavior. Loading its
definition from YAML must not change its Screens, transitions, artifact
lifecycle, process execution, persistence, or error handling.

## Configuration Direction

The Flow configuration should describe the full pipeline hierarchy and all
static Screen behavior:

- Stage lifecycle groupings
- Ordered Steps
- Task combinations within each Step
- Editor, Exec, Interactive, Script, and Preview configuration
- Artifact identifiers
- Prompt and script paths
- Expected terminal output
- Available commands
- Transition destinations
- Preview next Steps
- Step enablement and execution mode

The loader must produce the same typed Flow definition currently returned by
the built-in Go definition. The runtime and generic Screens must remain
independent of the definition source.

After the YAML definition produces equivalent Idea Stage behavior, remove the
temporary built-in definition.

## Pipeline Control

The configurable pipeline should prepare the CLI for users to:

- Skip or jump over Steps.
- Enable or disable Steps.
- Switch a Step between Exec and Interactive operation where supported.
- Change ordered transitions without recompiling the CLI.
- Add later Spec, implementation, review, and finalization stages using the
  same runtime.

These controls must operate on loaded definitions rather than introduce
state-name checks or hardcoded Idea behavior.

## Loading And Validation

The definition hierarchy and validation inherited from CLI Flow 1 are
provisional. Before making that model the YAML contract, revisit whether Stage,
Step, Task, Artifact, and Screen are separate concepts, whether Stages exist at
all, whether a Screen is equivalent to a Task, and which relationships belong
in static configuration. Revise the types, runtime boundary, and validator when
that design work produces a simpler or more accurate model; compatibility with
the temporary Go definitions must not freeze an uncertain hierarchy into the
long-term configuration format.

Flow YAML should be parsed into typed in-memory definitions before execution.
Validation should reject definitions that cannot be executed safely, including:

- Missing or duplicate identifiers
- Unsupported Screen or task types
- Missing artifacts
- Missing prompt or script paths required by a task type
- Missing expected output where an Exec task requires it
- Unknown command or transition destinations
- Invalid next-Step references
- Task combinations unsupported by the runtime

Relative prompt and script paths resolve from the active Flow directory.
Runtime values such as the active Change, temporary directory, session IDs,
originating navigation state, and task results remain Flow context and are not
stored as static YAML values.

The same definition validation must apply to YAML-loaded definitions and any
remaining test definitions.

## Compatibility And Inspection

The resolved Flow shown by `/config` should reflect the executable in-memory
definition rather than dump raw YAML. It should expose enough information to
understand Step order, task modes, artifacts, prompts, expected outputs,
commands, and destinations.

Existing `.mch/default` configuration should be migrated to the new typed
shape. Startup must fail clearly when the active Flow configuration is missing,
malformed, or invalid rather than silently falling back to built-in behavior.

The first YAML-backed pipeline may cover the Idea Stage only. Later Changes can
add the remaining stages after the generic runtime and configuration model have
been proven through real use.
