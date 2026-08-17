# internal.runtime.Reporter Service

## Status

- Service: `internal/runtime.Reporter`
- Package: `internal/runtime`
- Package name: `runtime`
- Shared models: `internal/models`
- Validation: `internal/runtime.Validator`
- External tools: terminal `jq`, optional terminal `bat`
- Command runners: `pkg/runner`
- Execution owner: `internal/runtime.StepRunner`
- Status: implementation specification

## Base specification

### Purpose

`Reporter` is the sole owner of human-readable standard output produced while
`StepRunner.Execute` runs. It reports one successful directory, reports type
and expected-response validation failures with their distinct layouts, and
prints one deferred debug dump for the runtime step that wins debug selection.

Regular output remains concise. A directory whose complete executed scope
passes receives one green-check line. A failed step receives a red-cross block
for each failed validation dimension. A selected debug step receives its
normal failure output when applicable and then one final bug-icon block
containing the complete executed runtime step as colored, pretty JSON.

`Reporter` presents validation payloads that `Validator` has already computed.
It does not repeat type validation, expected-member projection, or Git
comparison. Type validation returns compact JSON containing the failed
selector names. Expected validation returns the headerless Git diff. Reporter
extracts those payloads from their classifying errors and applies the
established terminal presentation.

### Responsibilities

`Reporter` owns:

- Writing all nonfatal human-readable standard output triggered by
  `StepRunner.Execute`.
- Printing one success line for a fully successful non-debug directory.
- Printing type-validation and expected-response failures through separate
  methods and separate output blocks.
- Preserving the established relative-directory and exact request-path
  presentation.
- Pretty-printing failed type-selector JSON through terminal `jq` with forced
  color.
- Presenting generated Git diffs with the established TTY-only `bat` behavior
  and the established in-process color fallback.
- Marshaling the selected final runtime `models.Step` to JSON and pretty
  printing it through terminal `jq` with forced color.
- Serializing complete output blocks so concurrent directory goroutines cannot
  interleave their bytes.
- Returning output and formatter errors to `StepRunner` without printing an
  error diagnostic.

### Non-responsibilities

`Reporter` does not:

- Print fatal execution, configuration, dependency, cancellation, or internal
  errors. The caller reports a fatal error after `StepRunner.Execute` returns,
  through the application logger's fatal path.
- Decide which step wins a concurrent debug race.
- Stop scheduling steps, cancel a run, join directory goroutines, or choose a
  process exit code.
- Execute curl or populate `Step.Response.Body`.
- Validate response types or expected values.
- Select response values for type validation.
- Project actual JSON to the members declared by expected JSON.
- Generate a Git diff.
- Load, substitute, capture, or retain variables.
- Print a green success line for the directory terminated by a debug step.
- Print a normal passing status for each individual step.
- Produce the future `--json` NDJSON event stream.

## Required prerequisite contract changes

The model, runner, validator, and step-runner changes in this section are part
of this specification. This specification supersedes only the conflicting
contracts identified below; all unaffected behavior in specifications
`05-runner-pkg.md`, `07-validator.md`, and `08-step-runner.md` remains in force.

### Add JSON tags to `models.Step`

Every declarative and runtime result member of `models.Step` must have a JSON
tag matching its YAML field name. No JSON field uses `omitempty`: the debug
dump represents the complete final runtime step, including zero, empty, nil,
and false values.

The required shape is equivalent to:

```go
type Step struct {
    Vars map[string]YAMLString `yaml:"vars" json:"vars"`
    Request struct {
        Method   string            `yaml:"method" json:"method"`
        BaseURL  string            `yaml:"baseUrl" json:"baseUrl"`
        BasePath string            `yaml:"basePath" json:"basePath"`
        Path     string            `yaml:"path" json:"path"`
        Headers  map[string]string `yaml:"headers" json:"headers"`
        Timeout  int               `yaml:"timeout" json:"timeout"`
        Retries  int               `yaml:"retries" json:"retries"`
        Query    string            `yaml:"query" json:"query"`
        Body     YAMLString        `yaml:"body" json:"body"`
    } `yaml:"request" json:"request"`
    Response struct {
        Status   []int                 `yaml:"status" json:"status"`
        Body     string                `yaml:"body" json:"body"`
        Expected YAMLString            `yaml:"expected" json:"expected"`
        Types    map[string][]string   `yaml:"types" json:"types"`
        Capture  map[string]YAMLString `yaml:"capture" json:"capture"`
    } `yaml:"response" json:"response"`
    Debug      bool             `yaml:"debug" json:"debug"`
    Definition *StepsDefinition `yaml:"-" json:"-"`
    Index      int              `yaml:"-" json:"index"`
}
```

`Definition` is excluded because it is provenance, not part of the resolved
step, and links into a cyclic directory and definition graph. `Index` is
included because it identifies the step within its definition. Debug output
must marshal successfully without traversing `Definition`.

#### Acceptance criteria

##### AC-StepJSON-1: Use the YAML field names

Given a populated runtime step, when Go marshals it as JSON, then the members
use `vars`, `request`, `response`, `debug`, and `index` and the nested members
use the corresponding YAML names, including `baseUrl` and `basePath`.

##### AC-StepJSON-2: Include complete runtime state

Given zero, false, nil, and empty members, when Go marshals the step, then no
declared JSON member is omitted.

##### AC-StepJSON-3: Exclude cyclic provenance

Given a step whose `Definition` points into the complete suite tree, when Go
marshals it, then `Definition` is absent and no cyclic-value error occurs.

##### AC-StepJSON-4: Include the runtime index

Given a step at index `3`, when Go marshals it, then the output contains
`"index":3`.

### Add pretty-JSON runner support

Add this contract to `pkg/runner`:

```go
var JQPrettyError = errors.New("jq pretty error")

func JQPretty(
    ctx context.Context,
    input string,
) (string, int, error)
```

`JQPretty` starts `jq` directly, without a shell, with the equivalent argument
slice:

```text
jq -C .
```

It supplies `input` unchanged through standard input. `-C` is mandatory and
forces the established JSON colors even when Reporter writes to a non-TTY.
On jq exit `0`, return jq standard output unchanged, exit code `0`, and nil.
The successful output is jq's two-space pretty JSON and includes jq's trailing
line ending.

On jq operational failure, return an empty output, jq's exact exit code, and
an error matching `JQPrettyError` and containing jq standard error. A command
startup failure returns `"", 0` and an error matching the existing
`runner.CommandError` and identifying `jq`. Context cancellation terminates
and waits for jq and returns an error matching `ctx.Err()`.

`JQPretty` uses call-local buffers and is safe for concurrent invocations. It
does not validate an application schema, sort object members independently of
jq, remove ANSI escapes, or invoke a shell.

#### Acceptance criteria

##### AC-JQPretty-1: Pretty print with forced color

Given valid compact JSON, when `JQPretty` succeeds, then it invokes exactly
`jq -C .`, returns colored two-space pretty JSON with its trailing newline,
exit code `0`, and nil.

##### AC-JQPretty-2: Forward jq failures

Given invalid JSON or another jq operational failure, when `JQPretty` returns,
then its output is empty, its exit code is jq's exact code, and its error
matches `JQPrettyError`.

##### AC-JQPretty-3: Honor cancellation without a shell

Given jq is running when the context is canceled, when `JQPretty` returns,
then jq is terminated and waited for, the error matches the context error,
and no shell was invoked.

### Add optional diff-color runner support

Add this contract to `pkg/runner`:

```go
var BatDiffError = errors.New("bat diff error")

func BatDiff(
    ctx context.Context,
    diff string,
) (string, int, error)
```

`BatDiff` invokes `bat` directly with exactly:

```text
bat --color=always --style=plain --language=diff --paging=never
```

It supplies the complete headerless diff unchanged through standard input.
On success it returns bat standard output unchanged, exit code `0`, and nil.
On a non-zero bat exit it returns empty output, the exact exit code, and an
error matching `BatDiffError`. Startup failures use `runner.CommandError` and
identify `bat`. Cancellation terminates and waits for the process and returns
an error matching the context error.

`bat` remains an optional presentation dependency. Reporter, not `BatDiff`,
decides whether a bat error is suppressed in favor of fallback coloring.

#### Acceptance criteria

##### AC-BatDiff-1: Preserve the established invocation

Given a diff, when bat succeeds, then `BatDiff` invokes the exact argument
sequence above without a shell and supplies the diff byte for byte on standard
input.

##### AC-BatDiff-2: Expose optional-command failures

Given bat is missing or exits unsuccessfully, when `BatDiff` returns, then it
returns the classifying command or bat error and its applicable exact exit
code, allowing Reporter to select its fallback.

### Aggregate and classify validator failures

Add these public sentinels to `internal/runtime`:

```go
var TypeValidationError = errors.New("type validation failed for")
var ExpectedValidationError = errors.New("response does not match expected")
```

Every nonfatal type-validation result from one step is aggregated into one
error. The error payload is a compact JSON array containing only the selectors
whose declarations failed. Selectors retain `ValidateTypes`' established
ascending bytewise lexicographic evaluation order. For example, given:

```json
{
  "t1": ["something"],
  "t2": ["something"],
  "t3": ["something"]
}
```

when `t1` and `t2` fail, the payload is exactly:

```json
["t1","t2"]
```

The nonfatal error is created equivalently to:

```go
fmt.Errorf("%w: %s", TypeValidationError, `["t1","t2"]`)
```

It must match `TypeValidationError` through `errors.Is`. The JSON contains
selector keys only; it does not contain declarations, selected response
values, individual explanations, or fatal jq errors.

`ValidateTypes` retains its `[]error` return type. A completed validation with
one or more rejected declarations returns exactly one nonfatal aggregate error
instead of one error per selector. If an operational failure occurs after
some selectors have failed, the returned slice may contain the aggregate
nonfatal error followed by the fatal error as required to preserve the
established complete-slice fatal scan. StepRunner must not report the
nonfatal payload for an incomplete step whose returned slice also contains a
fatal error.

Every nonfatal expected-response difference is one error whose payload is
exactly the headerless Git diff already produced by `runner.GitDiff`. It is
created equivalently to:

```go
fmt.Errorf("%w: %s", ExpectedValidationError, diff)
```

It must match `ExpectedValidationError` through `errors.Is`. The existing
projection, recursive ordering, pretty formatting, expected-first comparison,
and Git difference-status behavior remain owned by `Validator`.

All fatal validator errors retain the existing `ErrValidatorFatal` wrapping
and external exit-code behavior. They are not converted to either nonfatal
sentinel.

#### Acceptance criteria

##### AC-ValidatorPayload-1: Aggregate failing type selectors

Given several declarations fail, when `ValidateTypes` completes without a
fatal error, then it returns exactly one error matching
`TypeValidationError`, and the suffix after the static prefix is a compact
JSON array of all and only failed selectors in lexicographic order.

##### AC-ValidatorPayload-2: Keep type success empty

Given every type declaration passes or no type declarations exist, when
`ValidateTypes` returns, then it returns nil and does not create an empty
type-validation error.

##### AC-ValidatorPayload-3: Classify the expected diff

Given expected validation produces a non-empty headerless Git diff, when
`ValidateExpected` returns, then it returns exactly one nonfatal error matching
`ExpectedValidationError` whose payload is that diff byte for byte.

##### AC-ValidatorPayload-4: Preserve fatal classification

Given jq, Git, cancellation, or another infrastructure operation fails, when
either validator returns, then the fatal error continues to match
`ErrValidatorFatal`, preserves its underlying cause and exit code, and is not
reported as a nonfatal type or expected mismatch.

## Public contract

```go
var ReporterError = errors.New("reporter error")

type Reporter struct {
    output io.Writer
    // private output serialization, prior-output, and terminal-detection state
}

func NewReporter(output io.Writer) *Reporter

func (r *Reporter) Success(
    ctx context.Context,
    directory *models.Directory,
) error

func (r *Reporter) FailureTypes(
    ctx context.Context,
    step *models.Step,
    failure error,
) error

func (r *Reporter) FailureDiff(
    ctx context.Context,
    step *models.Step,
    failure error,
) error

func (r *Reporter) Debug(
    ctx context.Context,
    step *models.Step,
) error
```

All four reporting methods are called only by `StepRunner.Execute`. Other
layers do not write human execution output directly.

### Dependency and lifetime

`NewReporter` retains the exact non-nil output writer. The writer is normally
standard output and is injected for byte-exact tests. Construction writes no
output, invokes no external command, and retains no directory or step.

One Reporter may serve all directory goroutines in one suite execution. It is
safe for concurrent method calls. Every method constructs its complete output
block before locking the output mutex and performs one serialized writer
operation for that block. Under the same lock, Reporter records whether an
earlier block was written so `Debug` can insert its conditional separator.
External formatting may run concurrently before the write lock is acquired.
Reporter does not require the supplied writer itself to be concurrency-safe.

Reporter retains no context, directory, step, validation error, formatted JSON,
or diff after a method returns.

An output write error is returned wrapped with `ReporterError` and preserves
the writer cause. A nil Reporter, writer, directory, step, or required failure
is a caller error: the method returns an error matching `ReporterError` and
does not panic.

### Relative directory and request paths

Reporter derives the suite root by following `Directory.Parent` pointers to
the directory whose parent is nil.

- The root directory renders as `/`.
- A descendant is its cleaned path relative to the root path, converted to
  forward slashes and prefixed with `/`.
- A step's directory comes from `Step.Definition.File.Directory`.
- A step's request path is `Step.Request.Path` exactly.
- Base URL, base path, and query are never included in a status header.

Invalid or cyclic provenance returns an error matching `ReporterError` rather
than hanging or panicking.

## `NewReporter`

```go
func NewReporter(output io.Writer) *Reporter
```

`NewReporter` stores the writer and initializes the private serialization
state. Terminal detection follows the established behavior: an `*os.File`
whose mode includes `os.ModeCharDevice` is a TTY; other writers and failed
file-stat calls are non-TTY. Tests may use an unexported injection seam for
terminal detection and optional diff formatting.

### Acceptance criteria

#### AC-NewReporter-1: Retain the writer without work

Given a writer, when `NewReporter` is called, then it retains that exact writer
and does not write, inspect a tree, or invoke jq or bat.

## `Success`

```go
func (r *Reporter) Success(
    ctx context.Context,
    directory *models.Directory,
) error
```

`Success` writes exactly one line for a directory whose complete normally
scheduled scope executed without validation failure:

```text
 ✅ <relative-directory>
```

There is exactly one ASCII space before the icon and one between the icon and
directory. The line has one trailing line ending and no following empty line.
No request path is printed. The line receives no additional ANSI wrapper.

StepRunner calls `Success` once after the directory finishes, not after each
passing step. It does not call `Success` for:

- A directory containing any type or expected validation failure.
- A directory interrupted by a fatal error.
- The directory containing the selected debug step, even when every executed
  step through the debug boundary passed.
- A directory stopped before it begins by another directory's selected debug
  step.

### Acceptance criteria

#### AC-Success-1: Report one complete directory

Given every step in a non-debug directory passes, when StepRunner reports its
completion, then Reporter writes one green-check line containing only the
relative directory.

#### AC-Success-2: Do not report incomplete or failed directories

Given a validation failure, fatal interruption, or selected debug boundary in
a directory, when execution stops or completes that directory, then Reporter
does not write a green-check line for it.

## `FailureTypes`

```go
func (r *Reporter) FailureTypes(
    ctx context.Context,
    step *models.Step,
    failure error,
) error
```

The supplied error must match `TypeValidationError`. Reporter removes exactly
the static `TypeValidationError.Error()+": "` prefix and supplies the
remaining compact JSON array to `runner.JQPretty`.

Reporter writes one block equivalent to:

```text
 ❌ <relative-directory> <request-path>
type check failed for [
  "t1",
  "t2"
]

```

The JSON portion is the exact colored jq output. The prose prefix is
`type check failed for ` on the same line as jq's opening `[`.
There is one empty line after the jq output. Reporter does not print the
validator's static error prefix a second time.

When output is a TTY, the complete red-cross status line is wrapped in the
established light-cyan sequence:

```text
ESC[38;2;160;255;255m<status line>ESC[0m
```

The detail label is not wrapped. Jq supplies its own forced JSON colors on TTY
and non-TTY output.

### Acceptance criteria

#### AC-FailureTypes-1: Pretty print all failed selectors

Given a type error containing `["t1","t2"]`, when `FailureTypes` succeeds,
then it prints one red-cross block and `type check failed for ` followed by the
colored two-space jq rendering of exactly `t1` and `t2`.

#### AC-FailureTypes-2: Reject the wrong error class

Given nil or an error that does not match `TypeValidationError`, when
`FailureTypes` is called, then it writes nothing and returns an error matching
`ReporterError`.

#### AC-FailureTypes-3: Preserve formatter and writer failures

Given jq cannot format the payload, when `FailureTypes` returns, then it writes
nothing and returns the classifying formatter error. Given the output writer
fails, Reporter returns an error matching `ReporterError`; a writer that
accepts some bytes before returning its error may leave those bytes observable.

## `FailureDiff`

```go
func (r *Reporter) FailureDiff(
    ctx context.Context,
    step *models.Step,
    failure error,
) error
```

The supplied error must match `ExpectedValidationError`. Reporter removes
exactly the static `ExpectedValidationError.Error()+": "` prefix. The
remaining bytes are the complete headerless Git diff.

Reporter writes:

```text
 ❌ <relative-directory> <request-path>
response does not match expected:
@@ -1 +1 @@
-<expected>
+<actual>

```

The status line uses the same TTY-only light-cyan wrapper as
`FailureTypes`. The label is exactly `response does not match expected:` on
its own line.

For a non-TTY writer, Reporter writes the diff unchanged and does not invoke
bat. For a TTY writer, Reporter passes the complete diff to `runner.BatDiff`.
Successful bat output replaces the raw diff. If bat is missing, cannot start,
or exits unsuccessfully, Reporter suppresses that optional presentation error
and applies the established in-process fallback to the raw diff:

- Hunk lines beginning `@@` are cyan (`ESC[36m`).
- Removed lines beginning `-` are red (`ESC[31m`).
- Added lines beginning `+` are green (`ESC[32m`).
- Every colored line ends with `ESC[0m` before its original line ending.
- Other lines are unchanged.

If bat returns because `ctx` is canceled, Reporter returns the context error
instead of treating cancellation as an optional fallback. Reporter preserves
the diff's embedded line endings, adds a line ending only when the selected
rendering lacks one, and writes one empty line after the complete diff.

### Acceptance criteria

#### AC-FailureDiff-1: Render a plain non-TTY diff

Given a non-TTY writer and an expected mismatch, when `FailureDiff` succeeds,
then it does not invoke bat and writes the raw headerless diff beneath the exact
label with one trailing empty line.

#### AC-FailureDiff-2: Use bat for a TTY diff

Given a TTY writer and successful bat, when `FailureDiff` runs, then the raw
diff is supplied byte for byte to the exact bat invocation and the returned
colored output appears beneath the label.

#### AC-FailureDiff-3: Preserve the colored fallback

Given TTY output and missing or unsuccessful bat, when `FailureDiff` runs,
then reporting succeeds and hunk, removed, and added lines receive the
established cyan, red, and green fallback colors.

#### AC-FailureDiff-4: Reject the wrong error class

Given nil or an error that does not match `ExpectedValidationError`, when
`FailureDiff` is called, then it writes nothing and returns an error matching
`ReporterError`.

## Combined failure output

Type and expected validation are distinct reporting dimensions. When both
fail for one step, StepRunner calls `FailureTypes` first and `FailureDiff`
second. Reporter writes two complete red-cross blocks; it does not merge their
headers or details:

```text
 ❌ <relative-directory> <request-path>
type check failed for [
  ...
]

 ❌ <relative-directory> <request-path>
response does not match expected:
<diff>

```

### Acceptance criteria

#### AC-CombinedFailure-1: Keep distinct layouts and order

Given one step fails type and expected validation, when StepRunner reports the
result, then the complete type block precedes the complete diff block and no
bytes from another goroutine interleave either block.

## `Debug`

```go
func (r *Reporter) Debug(
    ctx context.Context,
    step *models.Step,
) error
```

`Debug` receives the selected final runtime element from
`Directory.RuntimeSteps`. The step has been attempted through its normal
execution and validation path. Its request and expected members contain any
runtime formatting completed before the execution outcome, and
`Step.Response.Body` contains curl output only when curl succeeded and
StepRunner assigned it.

Reporter calls `json.Marshal(step)`, then supplies the resulting JSON to
`runner.JQPretty`. It does not marshal `ResolvedSteps`, reconstruct data from
the YAML file, or format selected fields separately.

Reporter writes one block:

```text
 🐛 <relative-directory> <request-path>
{
  "vars": ...,
  "request": ...,
  "response": ...,
  "debug": true,
  "index": ...
}
```

There is exactly one ASCII space before the bug icon, after the icon, and
between directory and request path. The header receives no extra ANSI wrapper.
The JSON is jq's exact `-C` colored, two-space pretty output. It contains every
JSON-tagged Step member, including `debug` and `index`, and excludes
`Definition`.

When earlier stdout exists, exactly one empty line separates the preceding
complete output block or success line from the debug header. Reporter does not
write a leading empty line when debug is the first stdout block. Jq's trailing
newline terminates the run's final stdout; Reporter adds no second empty line
after the JSON.

`Debug` prints no execution or fatal-error diagnostic. It only prints the
step's current final runtime state.

### Acceptance criteria

#### AC-Debug-1: Print the complete final runtime step

Given a selected debug step whose runtime fields were mutated by request
formatting, curl response assignment, and response formatting, when `Debug`
runs, then the JSON contains those final values rather than their resolved
source values.

#### AC-Debug-2: Include failures only through normal reporting

Given the selected debug step fails type validation, expected validation, or
both, when output is written, then its normal red-cross block or blocks appear
first and the bug-icon JSON block appears last. The debug JSON does not embed
the validation errors as additional fields.

#### AC-Debug-3: Print after an execution error

Given the selected debug step was entered and curl, parsing, validation, or
another execution phase returns a fatal error, when StepRunner unwinds, then
the deferred debug block prints the step's state reached before that error and
the execution error remains the error returned to the caller.

#### AC-Debug-4: Return marshal, jq, and writer errors

Given JSON marshaling or jq formatting fails, when `Debug` returns, then it
writes nothing and returns the classifying error. Given output writing fails,
it returns an error matching `ReporterError`; a writer that accepts some bytes
before returning its error may leave those bytes observable.

## StepRunner integration

### Dependency change

`Reporter` replaces the raw output writer as StepRunner's presentation
dependency:

```go
type StepRunner struct {
    variableProcessor *VariableProcessor
    validator         *Validator
    reporter          *Reporter
}

func NewStepRunner(
    variableProcessor *VariableProcessor,
    validator *Validator,
    reporter *Reporter,
) *StepRunner
```

The constructor retains the exact non-nil Reporter. StepRunner does not write
standard output itself. This supersedes the `output io.Writer` field,
constructor parameter, output mutex, and inline regular rendering assigned to
StepRunner by `08-step-runner.md`.

### Normal reporting

For every completed step, StepRunner scans the complete validator return
slices for fatal errors before reporting nonfatal failures.

- A type error matching `TypeValidationError` calls
  `Reporter.FailureTypes(ctx, step, err)`.
- An expected error matching `ExpectedValidationError` calls
  `Reporter.FailureDiff(ctx, step, err)`.
- When both exist, type output is emitted first.
- A fatal validator result makes the step incomplete and emits neither
  nonfatal block from that step.
- A directory that completes without any validation failure calls
  `Reporter.Success(ctx, directory)` exactly once.
- A directory with one or more failed steps has no green line; each failure
  remains attached to its step through the failure header.

Reporter errors are fatal execution errors. StepRunner records the first such
error, cancels normal execution, joins active goroutines, and returns the error
with its cause and external-tool exit code preserved where applicable.

### Debug selection and stopping

Before running the phases of a step whose final runtime `Debug` value is true,
StepRunner atomically attempts to claim that step as the run's debug step.
Exactly one step can win. With sequential debug steps, the first reached step
wins. When same-stage directory goroutines race, the winner is intentionally
nondeterministic; users requiring deterministic debug behavior must declare
only one reachable debug step.

Once a step wins:

- The winner continues its normal request execution and validation phases.
- StepRunner starts no later step in that directory.
- Other goroutines check the shared debug-stop state before starting another
  step and stop when they observe it.
- A step already in progress in another directory may finish and report its
  outcome.
- The active stage is joined completely.
- No later stage starts.
- The selected directory receives no green success line.
- Earlier and already-started steps may have emitted green or red output.
- The winner receives its red validation output when it fails validation.

Debug stop is an intentional scheduling boundary, not a fatal cancellation.
It must not cancel the context needed by the winning step or an already-started
sibling merely to stop later scheduling.

The winning debug step skips `VariableProcessor.Capture`. No later scheduled
step can consume a captured value. The remaining phase order is the established
order with that operation omitted. A response expectation in a debug step may
use values captured by earlier completed steps but cannot rely on a capture
declared by that same debug step.

When multiple debug steps race, only the selected winner's dump is guaranteed.
Whether a losing already-started debug step reaches or skips capture is not a
supported behavioral contract. The implementation must remain race-free and
must never emit more than one debug dump.

### Deferred final debug output

`StepRunner.Execute` registers a Go `defer` whose closure reads the atomically
selected debug-step pointer. The defer is registered before traversal so it
runs for every normal or error return after a debug step has been selected.
The closure invokes Reporter only after execution has stopped, the active
stage's directory goroutines have joined, and all normal Reporter calls have
finished.

The deferred call is equivalent in outcome to:

```go
defer func() {
    if selectedDebugStep == nil {
        return
    }
    debugErr := reporter.Debug(reportContext, selectedDebugStep)
    if executeErr == nil {
        executeErr = debugErr
    }
}()
```

The final reporting context must remain usable even when the shared execution
context was canceled by a fatal error. This ensures jq can render an already
selected step before the execution error reaches the terminal error path.

Error precedence is exact:

- An existing execution error is preserved when deferred debug formatting also
  fails. The debug error must not replace its cause or exit code.
- When execution otherwise succeeds and deferred debug formatting fails,
  `Execute` returns the debug error and its applicable exact exit code.
- When execution accumulated validation failures and deferred debug succeeds,
  `Execute` returns the established validation error and exit code `1`.
- A fatal error remains fatal and is printed only by the caller after
  `Execute` returns. Therefore the debug block is the final standard-output
  block even though a later fatal diagnostic may be written to standard error.

### Superseded StepRunner output rules

This specification replaces these conflicting rules from
`08-step-runner.md`:

- StepRunner no longer owns an output writer or inline output formatting.
- Passing output is one line per successful directory, not one line per
  passing step.
- Type failures are one aggregate selector-array error per step, not one
  presentation line per rejected selector.
- Type and expected failures produce two independent red-cross blocks when
  both fail, rather than one merged block.
- Human output retains the established colors described here rather than the
  plain-text-only rule.
- A selected debug step stops later scheduling, skips capture, has no green
  directory line, and produces one deferred final debug block.

Stage barriers, directory concurrency, step ordering, runtime element pointer
identity, curl response assignment, normal non-debug capture behavior, fatal
cancellation, joining, validation exit code `1`, and all other unaffected
StepRunner contracts remain unchanged.

### Integration acceptance criteria

#### AC-Integration-1: Report one success per directory

Given several passing steps in one non-debug directory, when the directory
finishes, then stdout contains one green-check directory line and no per-step
passing line.

#### AC-Integration-2: Report distinct validation dimensions

Given one step fails types and expected comparison, when it completes, then
stdout contains the type block followed by the diff block, each with its own
red-cross header, and the run ultimately returns exit code `1`.

#### AC-Integration-3: Stop at one debug winner

Given normal work before and after a reachable debug step, when the debug step
is selected, then it executes and validates, later unscheduled work does not
start, already-started sibling work joins, and exactly one debug block is
printed.

#### AC-Integration-4: Make debug final stdout

Given prior successes, prior validation failures, and already-started sibling
results, when a debug step is selected, then all of those complete output
blocks precede the bug-icon block and no stdout follows its JSON.

#### AC-Integration-5: Print debug on fatal unwind

Given the selected debug step starts and then a fatal execution phase fails,
when StepRunner returns, then its deferred debug state is printed before the
caller receives and logs the original fatal error.

#### AC-Integration-6: Preserve execution-error precedence

Given execution and deferred debug formatting both fail, when `Execute`
returns, then the original execution error and exit code remain discoverable
and are not replaced by the debug failure.

#### AC-Integration-7: Do not report debug when it was never entered

Given a fatal error stops execution before a configured debug step is reached,
when `Execute` returns, then no bug-icon block is printed.

#### AC-Integration-8: Handle concurrent debug declarations safely

Given two same-stage directories reach debug steps concurrently, when the run
stops, then exactly one step is selected and printed, output is race-free, and
the specification promises no identity for the winner.

#### AC-Integration-9: Skip the winner's capture

Given the selected debug step declares response captures, when it executes,
then StepRunner does not invoke Capture for that step and no later step starts
to consume those values.

#### AC-Integration-10: Keep fatal diagnostics out of Reporter

Given a fatal error with or without a selected debug step, when StepRunner and
Reporter finish, then Reporter writes no fatal diagnostic; the error is
returned for the caller's logger fatal path.

## Concurrency and atomicity

Reporter protects output at block granularity. A type block, diff block,
success line, or debug block is written through one serialized call and its
bytes cannot interleave with another Reporter call. StepRunner invokes type
reporting before diff reporting for one step and does not advance that
directory between those calls. A complete block from another concurrently
finishing directory may appear between the type and diff blocks; same-stage
directory output retains the established actual-completion-order freedom.

The selected debug pointer and debug-stop state use synchronization valid under
the Go memory model. The deferred reader must observe all mutations completed
on the winning runtime step before its directory goroutine joins. Reporter
must never retain or read the step concurrently with an active mutation.

## Error and exit-code contract

- Nonfatal type mismatches match `TypeValidationError`.
- Nonfatal expected mismatches match `ExpectedValidationError`.
- A completed run containing either mismatch returns the established
  `ErrStepValidation` classification and exit code `1`.
- Reporter formatting and writer errors are fatal and preserve an external
  process exit code when one exists.
- Optional bat discovery or non-cancellation execution failure is suppressed
  after fallback coloring succeeds.
- Fatal Validator errors retain `ErrValidatorFatal` and their original cause.
- A selected debug dump never changes an already-selected execution error or
  exit code.
- Reporter never calls `os.Exit`, `log.Fatal`, or the application logger.

Every returned error must contain or wrap a stable static error before dynamic
detail. The required nonfatal formats are:

```text
type validation failed for: <compact selector JSON>
response does not match expected: <headerless diff>
```

## Required tests

Implementation must add mapped automated tests for every acceptance criterion.
At minimum, coverage must include:

- JSON tag names, zero-value inclusion, index inclusion, and cyclic Definition
  exclusion.
- `runner.JQPretty` exact `jq -C .` invocation, forced ANSI output, invalid
  input, startup failure, non-zero exit, cancellation, concurrency, and
  no-shell behavior.
- `runner.BatDiff` exact invocation, stdin preservation, failure codes,
  cancellation, concurrency, and no-shell behavior.
- Validator aggregation of zero, one, and several type selector failures in
  lexicographic order.
- `errors.Is` classification for type, expected, and fatal validator errors.
- Preservation of the exact compact type-selector array and headerless diff
  payloads.
- Reporter root and nested relative paths, exact request paths, Windows path
  separator normalization, and invalid provenance.
- Byte-exact success, type, diff, combined-failure, and debug layouts.
- TTY light-cyan failure headers, jq forced colors, successful bat output,
  missing and failed bat fallback colors, and plain non-TTY diffs.
- No output on pre-write jq failure, one serialized write attempt per block,
  and surfaced writer failures including short writes.
- Concurrent Reporter calls without byte interleaving under `go test -race`.
- One success line for several passing steps in a directory and no green line
  for failed, fatal, or debug-terminated directories.
- Type-before-diff ordering and two distinct blocks for a step failing both.
- Debug selection before execution, winner phase execution, skipped winner
  capture, stop-before-next-step checks, active sibling joining, and no later
  stage.
- Exactly one nondeterministic winner under a same-stage debug race.
- Deferred debug after prior success and failure output.
- Deferred debug after curl, jq, Git, output, and validation outcomes where the
  debug step was selected.
- No debug output when fatal termination occurs before a debug step is reached.
- Execution-error precedence when deferred debug also fails.
- Validation exit code `1` with a final successful debug rendering.
- Fatal diagnostics absent from Reporter output.

Integration tests must use deterministic fake executables or controlled PATH
fixtures. They must not depend on installed jq, bat, curl, or Git behavior
outside the runner-specific integration boundary.

## Verification

- `go test -race -short ./...` — runs all unit and short integration tests with
  race detection, including concurrent Reporter and debug selection coverage.
- `go test -coverprofile=coverage.out ./...` — records repository-wide
  statement coverage.
- `go tool cover -func=coverage.out` — confirms repository-wide coverage
  remains at least 90%.
- `golangci-lint run --timeout 10m` — validates error wrapping, naming,
  concurrency, and changed Go code.
- `go build ./...` — verifies model tags, runner additions, Reporter, Validator,
  and StepRunner integration compile.
- `git diff --check` — verifies whitespace integrity.

## Review focus

- Confirm Reporter is the only standard-output owner used by
  `StepRunner.Execute` and fatal errors remain outside Reporter.
- Verify type failures contain only the failed selector names in one compact,
  classifying JSON error and Reporter does not rerun validation.
- Verify expected validation still owns projection, canonicalization, and Git
  comparison while Reporter owns only presentation.
- Check the exact success, type, diff, and bug-icon output bytes, including
  leading spaces, paths, labels, blank lines, final newline, and ANSI behavior.
- Confirm JSON marshaling uses the final runtime element, includes Index and
  all zero-valued public fields, and excludes cyclic Definition provenance.
- Inspect the debug-stop synchronization carefully: one winner, no later
  scheduling, winner execution and validation, active sibling joining, and one
  final deferred dump.
- Verify the deferred formatter uses a viable reporting context after fatal
  execution cancellation and cannot replace the original execution error.
- Confirm the winning debug step skips capture and no later step can depend on
  its captures.
- Verify output blocks remain contiguous and race-free under concurrent
  directory completion.
- Confirm bat remains optional, is invoked without a shell only for TTY diffs,
  and falls back to the established in-process colors.

## Follow-ups

- Machine-readable `--json` NDJSON events remain deferred to the PRD's later
  output implementation.
- Deterministic selection among multiple concurrently reachable debug steps is
  intentionally not provided. Users must declare a single reachable debug
  step when winner identity matters.
