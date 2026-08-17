# `internal/execution` StepRunner

## Status and authority

- Package: `internal/execution`
- Shared models: `internal/domain`
- Presentation: `internal/reporter`
- External commands: `pkg/runner`
- Error builder and exit codes: `pkg/errs`
- Binding reference: `skeleton/internal/execution/steprun.go`
- Status: skeleton-aligned specification

## Purpose

`StepRunner` prepares resolved steps and executes the suite. It validates the
directory graph, schedules directories by stage, preserves sequential file and
step order within each directory, and reports nonfatal validation failures
through an injected Reporter.

## Public contract

```go
var ErrInvalidDirectoryTree = errors.New("invalid directory tree")
var ExecutionCanceledError = errors.New("execution canceled")

type StepRunner struct {
    // private retained collaborators
}

func NewStepRunner(
    variableProcessor *VariableProcessor,
    validator *Validator,
    report *reporter.Reporter,
) *StepRunner

func (s *StepRunner) Prepare(
    ctx context.Context,
    suite *domain.Suite,
) error

func (s *StepRunner) Execute(
    ctx context.Context,
    suite *domain.Suite,
) (int, error)
```

The package and constructor names above are exact. Legacy
`internal.runtime.StepRunner`, raw `io.Writer` injection, and differently named
parse phases are not part of the contract.

## Construction

`NewStepRunner` retains the exact supplied `VariableProcessor`, `Validator`,
and `*reporter.Reporter`. Construction performs no preparation, execution, or
output.

## `Prepare`

`Prepare` traverses directories from `suite.Root`. For every resolved step, in
the order represented by `Directory.ResolvedSteps`, it calls:

1. `VariableProcessor.Load(ctx, step)`
2. `VariableProcessor.ParseRequestBody(ctx, step)`

It returns nil after every step is prepared. A non-nil error stops preparation
and is returned. The exact runtime-copy policy and transactional behavior are
not defined by the current skeleton and must not be added as requirements
without first changing the reference.

## `Execute` input validation

Before starting goroutines, `Execute` validates the reachable tree:

- `suite` is non-nil;
- `suite.Root` is non-nil;
- the root stage is `0`;
- every child pointer is non-nil;
- no directory pointer is reachable more than once;
- every child identifies its containing directory through `Parent`;
- every child stage equals its parent stage plus one.

Cycles are rejected by the repeated-pointer check. Invalid input returns a
built error matching `ErrInvalidDirectoryTree` and exit code `102`; it must not
panic.

Valid directories are grouped by their declared stage. The implementation
supports arbitrary valid depth rather than a fixed stage limit.

## Stage scheduling

Execution follows this exact schedule:

1. Process stages in ascending numeric order beginning at stage `0`.
2. Start exactly one goroutine for every directory in the active stage.
3. Allow those directory goroutines to run concurrently.
4. Join every active-stage goroutine before starting the next stage.
5. Create no file or step goroutines.

Within a directory, process the outer step groups one at a time in the
alphabetical file order established by `StepsFiles`. Process every step in one
group sequentially in declaration order before starting the next group.

No ordering is promised between different directories in the same stage.

## Per-step execution

For each runtime step, StepRunner invokes the current skeleton phases in this
order:

1. `runner.Curl`
2. `VariableProcessor.ParseResponseExpected`
3. `Validator.ValidateTypes`
4. `Validator.ValidateExpected`
5. `VariableProcessor.Capture`

Only `pkg/runner` may invoke the external command. StepRunner passes the
resolved request values required by `runner.Curl` and retains the response in
the runtime step for later phases.

`Load` and `ParseRequestBody` belong to `Prepare`, not `Execute`. Capture runs
after both validation methods in the binding skeleton. Legacy phase names
`ParseRequest` and `ParseResponse`, capture-before-validation behavior, HTTP
status validation, and Git comparison are not part of this contract.

## Validation results

`ValidateTypes` can return multiple errors. `ValidateExpected` can return one
error. A nonfatal mismatch from either validator is reported through the
injected Reporter:

- type failures use `Reporter.FailureTypes`;
- expected-response failures use `Reporter.FailureExpected`.

StepRunner continues all remaining steps, files, directories, and stages after
nonfatal validation failures. If at least one occurred, the complete run
returns an error matching `ValidationError` with exit code `101`.

The current skeleton also exposes `Reporter.Success` and `Reporter.Debug` as
presentation entry points, but does not specify their text or add debug-based
scheduling behavior.

## Fatal errors and cancellation

The first fatal directory error:

- records that original error and its exit code;
- cancels the shared execution context;
- prevents later work from being started once cancellation is observed;
- joins every directory goroutine already started for the active stage;
- prevents all later stages from starting.

The first fatal error wins over sibling errors caused by shared cancellation.
If a failing operation returns exit code `0`, StepRunner derives a non-zero code
from the built error and otherwise uses internal code `103`. `Execute` never
returns a non-nil error with exit code `0`.

If the caller's context is canceled between otherwise successful stages,
StepRunner returns a built error matching `ExecutionCanceledError`, preserves
the context error, and uses code `103`.

## Error ownership

The static errors `ErrInvalidDirectoryTree`, `ExecutionCanceledError`, and
`ValidationError` originate in `internal/execution`. Contextual errors are
built only by `pkg/errs`. StepRunner does not use `fmt.Errorf`, decorate errors
locally, or print fatal diagnostics.

External tool errors retain the exact non-zero code returned by `pkg/runner`.
Configuration failures use `102`, aggregate validation failure uses `101`, and
uncoded internal failures use `103`.

## Output boundary

StepRunner never retains an `io.Writer`, writes to `os.Stdout`, or formats
human output directly. All standard output goes through the retained Reporter.
Reporter errors are ordinary fatal execution errors and follow the same
cancellation and join rules.

## Non-responsibilities

StepRunner does not:

- discover, decode, validate, or resolve definitions;
- define variable substitution, capture, or response-validation semantics;
- invoke a command outside `pkg/runner`;
- write fatal diagnostics;
- define Reporter colors, layouts, summaries, or machine-readable events;
- add filtering, HTTP-status behavior, or debug-stop behavior absent from the
  skeleton.

## Acceptance criteria

1. The constructor retains the processor, validator, and Reporter and performs
   no work.
2. Preparation invokes `Load` then `ParseRequestBody` for each resolved step.
3. Every invalid graph shape listed above returns code `102` and an error
   matching `ErrInvalidDirectoryTree` without panic.
4. Stages have complete barriers; same-stage directories can overlap; steps in
   one directory never overlap.
5. Files and steps retain their established deterministic order.
6. Runtime phases use the exact skeleton order.
7. Every nonfatal validation failure is routed through Reporter, remaining
   work completes, and the final result matches `ValidationError` with code
   `101`.
8. A fatal error cancels siblings, joins the active stage, prevents later
   stages, and retains the first fatal cause and code.
9. StepRunner performs no direct standard-output write or external command
   invocation outside `pkg/runner`.
10. Stage scheduling and tree validation remain race-detector clean.
