# internal.runtime.StepRunner Service

## Status

- Service: `internal/runtime.StepRunner`
- Package: `internal/runtime`
- Package name: `runtime`
- Shared models: `internal/models`
- Variable processing: `internal/runtime.VariableProcessor`
- Response validation: `internal/runtime.Validator`
- Command runner: `pkg/runner`
- Status: implementation specification

## Base specification

### Purpose

`StepRunner` prepares and executes the runtime copies of all resolved steps in
one directory tree. Preparation copies every `Directory.ResolvedSteps` matrix
to the corresponding `Directory.RuntimeSteps` matrix. Execution runs stages
sequentially, directories in one stage concurrently, and every runtime step
within one directory sequentially through variable loading, request parsing,
curl, response capture, response parsing, type validation, and expected-value
validation.

`StepRunner` owns the complete lifecycle of a runtime `models.Step`. Every
operation receives the address of the actual step stored in `RuntimeSteps` so
mutations made by one phase are visible to every later phase and remain on the
runtime model after execution.

Before `StepRunner` is implemented, this change also splits the phase-dependent
`VariableProcessor.Parse` operation into explicit request and response methods,
adds `runner.GitDiff` as the shared Git wrapper used by expected-value
validation, and makes every external-process runner accept the shared
cancellable context required for parallel fatal shutdown.

### Responsibilities

`StepRunner` owns:

- Copying resolved steps into independently mutable runtime-step matrices.
- Grouping directories by `Directory.Stage`.
- Executing stages sequentially in ascending numeric order with a barrier
  between stages.
- Starting exactly one goroutine for every directory in the active stage.
- Coordinating fatal cancellation and waiting for every started directory
  goroutine before returning.
- Executing outer step groups and their steps sequentially in stored order.
- Passing the exact `*models.Step` runtime element to every runtime service.
- Calling the runtime phases in the order defined by this specification.
- Building the curl URL from the resolved request components.
- Writing successful curl output to `Step.Response.Body` before response work.
- Collecting every nonfatal type and expected-value validation failure for a
  step.
- Rendering the established regular integration-test status and failure output.
- Continuing later steps after validation failures.
- Canceling immediately on configuration, variable, external-tool, output, or
  other fatal errors and returning after active goroutines join.
- Reporting exit code `101` after the complete tree executes when at least one
  step had a validation failure.

### Non-responsibilities

`StepRunner` does not:

- Discover, decode, select, or resolve suite files.
- Apply defaults or alter `Directory.ResolvedSteps`.
- Reimplement variable substitution, jq selection, type rules, expected-value
  projection, or Git comparison.
- Validate HTTP response status. The current `runner.Curl` contract returns the
  response body but not an HTTP status value.
- Execute files or steps within one directory concurrently.
- Define an execution order between different directories in the same stage.
- Reorder `Directory.Children`, either dimension of a step matrix, variables,
  response types, or captures.
- Print fatal errors, call `os.Exit`, or choose a process exit code other than
  the completed-run validation code `101`.
- Add ANSI color or verbose/debug output.
- Mutate a declarative step through its `Definition` provenance pointer.

## Required prerequisite contract changes

The prerequisite changes in this section are part of this specification and
must be completed before the `StepRunner` implementation.

### Split `VariableProcessor.Parse`

Remove the phase-dependent method:

```go
func (p *VariableProcessor) Parse(
    ctx context.Context,
    step *models.Step,
) (int, error)
```

Replace it with:

```go
func (p *VariableProcessor) ParseRequest(
    ctx context.Context,
    step *models.Step,
) (int, error)

func (p *VariableProcessor) ParseResponse(
    ctx context.Context,
    step *models.Step,
) (int, error)
```

`ParseRequest` performs substitution, JSON validation, recursive object-key
ordering, and pretty formatting only for `Step.Request.Body`. An omitted body
is a successful no-op. It must not inspect `Step.Response.Body` to infer the
runtime phase and must not alter `Step.Response.Expected`.

`ParseResponse` performs the same operations only for
`Step.Response.Expected`. An omitted expectation is a successful no-op. It
must not alter `Step.Request.Body`. The step runner calls it after `Capture`, so
the expectation may reference a variable captured by the same step.

Both methods retain the exit-code, error-wrapping, context, variable syntax,
jq, JSON formatting, mutation, and nil-input contracts already specified for
`VariableProcessor.Parse` in `06-variable-processor.md`. The old `Parse` method
must not remain as a compatibility wrapper because its phase inference through
`Step.Response.Body` is no longer part of the service contract.

#### Acceptance criteria

##### AC-ParseSplit-1: Parse only the request

Given a step with request and expected variable references, when
`ParseRequest` succeeds, then it updates only `Step.Request.Body` and leaves
`Step.Response.Expected` unchanged regardless of the current response body.

##### AC-ParseSplit-2: Parse only the response

Given a step with request and expected variable references, when
`ParseResponse` succeeds, then it updates only `Step.Response.Expected` and
leaves `Step.Request.Body` unchanged regardless of whether the response body is
empty.

##### AC-ParseSplit-3: Make same-step captures available

Given `Step.Response.Expected` references a variable declared by
`Step.Response.Capture`, when `Capture` succeeds before `ParseResponse`, then
`ParseResponse` reads the captured value from the shared store and formats the
resulting expectation.

##### AC-ParseSplit-4: Remove phase inference

Given any response-body value, when either split method runs, then its selected
member is determined only by the method name and never by
`Step.Response.Body`.

### Add `runner.GitDiff`

Add this contract to `pkg/runner`:

```go
var GitDiffError = errors.New("git diff error")

func GitDiff(
    ctx context.Context,
    expected string,
    actual string,
) (string, int, error)
```

The argument order is `expected, actual`. This is the conventional assertion
order and matches Git operand order: removed `-` lines describe expected
content and added `+` lines describe actual content.

`GitDiff` writes the two supplied strings unchanged to private files named
`expected` and `actual` in a newly created temporary directory. File
permissions must be no broader than `0600`. It starts Git directly, without a
shell and under the supplied context, using the equivalent arguments:

```text
git diff --no-index -U0 <expected-file> <actual-file>
```

The temporary directory is removed before the function returns. Temporary
paths and Git's file-header lines are process mechanics and must not appear in
the returned comparison text.

The return contract is:

- Git exit `0`: return `"", 0, nil`.
- Git exit `1`: remove the first four Git header lines and one trailing line
  ending from the diff, then return the remaining headerless diff, `0`, and
  nil. Exit `1` is Git's defined difference result, not an operational error.
- Git exit greater than `1`: return an empty diff, Git's exact exit code, and a
  non-nil error matching `GitDiffError` and containing Git's standard error.
- Git startup failure: return `"", 0`, and an error matching `CommandError`
  and identifying `git`.
- Context cancellation: terminate and wait for Git, return an empty diff and a
  non-nil error matching `ctx.Err()`, and do not report cancellation as a
  semantic difference.
- Temporary-directory or file-write failure: return `"", 0`, and a non-nil
  error describing the failed operation.

`GitDiff` compares the supplied documents as text. It does not validate JSON,
order members, project actual response members, add a final newline, interpret
the diff, or color its output.

`Validator.ValidateExpected` retains ownership of JSON validation,
canonicalization, and actual-response projection. After producing the final
expected and comparison-actual documents, it must call
`runner.GitDiff(ctx, expected, actual)` instead of starting Git itself. An
empty returned diff means validation passes. A non-empty returned diff becomes
the single nonfatal expected-value validation error, and the error's
presentation text must be exactly that headerless diff. A non-nil runner error
remains fatal under the existing `ErrValidatorFatal` contract.

This `runner.GitDiff` contract supersedes only the direct Git invocation and
temporary comparison-file ownership in `07-validator.md`; all projection and
validation behavior there remains unchanged.

#### Acceptance criteria

##### AC-GitDiff-1: Use expected-first comparison order

Given expected `{"id":1}` and actual `{"id":2}`, when Git reports a
difference, then the returned diff contains an expected line prefixed by `-`
and an actual line prefixed by `+`.

##### AC-GitDiff-2: Return no diff for equal documents

Given byte-identical expected and actual strings, when Git exits `0`, then
`GitDiff` returns an empty string, exit code `0`, and nil.

##### AC-GitDiff-3: Return a headerless semantic difference

Given different documents, when Git exits `1`, then `GitDiff` returns only the
hunk header and changed content, removes temporary paths and file headers,
returns exit code `0`, and returns nil.

##### AC-GitDiff-4: Forward operational Git failures

Given Git exits with status greater than `1`, when `GitDiff` returns, then its
diff is empty, its exit code is Git's exact status, and its error matches
`GitDiffError` and contains Git's diagnostic.

##### AC-GitDiff-5: Clean up private files

Given any match, difference, Git failure, or file failure after temporary
directory creation, when `GitDiff` returns, then no comparison directory or
file from that invocation remains.

##### AC-GitDiff-6: Do not invoke a shell

Given either input contains shell metacharacters, when `GitDiff` runs, then the
input is written only as file content and no shell interprets it.

##### AC-GitDiff-7: Honor cancellation

Given Git is running when the supplied context is canceled, when `GitDiff`
returns, then it terminates and waits for Git, removes its temporary files, and
returns an error matching the context error rather than a diff result.

### Make every runner operation context-aware

Parallel fatal shutdown requires every external process to use the one
cancellable execution context. Update the existing runner contracts to:

```go
func Curl(
    ctx context.Context,
    method string,
    url string,
    headers map[string]string,
    timeout int,
    retries int,
    query string,
    body string,
) (string, int, error)

func JQFilter(
    ctx context.Context,
    selector string,
    input string,
) (string, int, error)
```

Together with `GitDiff`, all three wrappers must start their command with
`exec.CommandContext`, stop and wait for the process when the context is
canceled, and return an error matching `ctx.Err()`. Their existing argument,
output, semantic-exit-status, operational-exit-status, no-shell, and
concurrency contracts remain unchanged.

`VariableProcessor.ParseRequest`, `ParseResponse`, and `Capture` must pass
their received context to `runner.JQFilter`. `StepRunner` passes its shared run
context to `runner.Curl`, and `Validator` passes its received context to
`runner.GitDiff` and every jq process it starts. These signature changes
supersede the context-free declarations in `05-runner-pkg.md`.

#### Acceptance criteria

##### AC-RunnerContext-1: Cancel curl

Given curl is running in one directory when a sibling directory reports a
fatal error, when the shared context is canceled, then curl is terminated and
waited for and returns the context error.

##### AC-RunnerContext-2: Cancel jq

Given jq is running during parsing, capture, or validation when a sibling
directory reports a fatal error, when the shared context is canceled, then jq
is terminated and waited for and returns the context error.

##### AC-RunnerContext-3: Preserve semantic statuses

Given no cancellation, when jq returns its accepted `false` or `null` status or
Git returns difference status `1`, then the wrappers retain their existing
nonfatal semantic behavior.

## Public contract

```go
var ErrStepValidation = errors.New("step validation failed")

type StepRunner struct {
    variableProcessor *VariableProcessor
    validator         *Validator
    output            io.Writer
}

func NewStepRunner(
    variableProcessor *VariableProcessor,
    validator *Validator,
    output io.Writer,
) *StepRunner

func (s *StepRunner) Prepare(
    root *models.Directory,
) error

func (s *StepRunner) Execute(
    root *models.Directory,
) error
```

The user-facing method contract intentionally has no context parameter.
`Execute` derives one cancellable run context from `context.Background()` and
passes it to every `VariableProcessor`, `Validator`, and runner call. All
directory goroutines share that context. The first fatal error cancels it.
`Execute` must not create unrelated contexts for individual directories or
phases and must not retain the context after returning.

`ErrStepValidation` classifies a completed tree containing at least one
nonfatal validation mismatch. The returned error must match
`ErrStepValidation` through `errors.Is` and expose exit code `101` through:

```go
interface {
    ExitCode() int
}
```

Fatal dependency and external-tool errors are selected immediately, trigger
cancellation, and are returned after all active-stage directory goroutines
join. When a dependency reports an exit code separately from its error,
`StepRunner` must return a wrapping error that preserves the cause for
`errors.Is` and `errors.As` and exposes that exact code through the same
`ExitCode` interface.

### Dependency and lifetime

`NewStepRunner` retains the exact supplied `VariableProcessor`, `Validator`,
and output writer. All dependencies must be non-nil. The output writer is
normally standard output but is injected so unit and integration tests can
verify byte-exact output without replacing process-global state.

One `StepRunner` and its `VariableProcessor` are scoped to one suite run. This
ensures every step in that run uses the same run-wide variable store. A new
suite run must receive a new variable store, processor, and step runner.

`StepRunner` retains no root, directory, step, response, error, or validation
result after a method returns. It is not safe for callers to invoke `Prepare`
or `Execute` concurrently on the same tree or runner. During one `Execute`, the
service creates and joins the directory goroutines required by the PRD.

The shared `VariableProcessor` is safe for concurrent calls because its
run-wide `KeyValueStore` performs atomic synchronized `Set` and `Get`
operations. `Validator` is stateless and concurrency-safe. `StepRunner`
serializes all access to its output writer; it does not require the injected
writer itself to be concurrency-safe.

## `NewStepRunner`

```go
func NewStepRunner(
    variableProcessor *VariableProcessor,
    validator *Validator,
    output io.Writer,
) *StepRunner
```

### Behavior

`NewStepRunner` returns a runner backed by the supplied dependencies. It does
not allocate another variable store, prepare a directory, execute a command,
or write output.

### Acceptance criteria

#### AC-NewStepRunner-1: Retain dependencies

Given a processor, validator, and writer, when `NewStepRunner` is called, then
subsequent preparation and execution use those exact values.

#### AC-NewStepRunner-2: Perform no work

When `NewStepRunner` is called, then it does not traverse a directory, mutate a
step, invoke an external command, or write output.

## `Prepare`

```go
func (s *StepRunner) Prepare(
    root *models.Directory,
) error
```

### Input

`root` is the suite root after `Resolver.ResolveSteps` has completed
successfully. Every directory's `ResolvedSteps` uses the deterministic matrix
shape defined by `03-resolver-service.md`:

```text
ResolvedSteps[steps-definition index][step declaration index]
```

A nil root, a nil child, a repeated directory pointer, a child whose `Parent`
does not identify its containing directory, or a resolved step without valid
`Definition` provenance is invalid input. `Prepare` returns a non-nil error and
must not panic.

### Traversal and copy behavior

`Prepare` traverses the root first, then each directory's `Children` depth-first
in exact stored order. It does not follow `Parent` as a traversal edge.

For every directory, it creates a candidate runtime matrix with the same outer
length, inner lengths, empty groups, step order, and values as
`ResolvedSteps`. It deep-copies every mutable map and slice reachable through a
step, including variables, headers, response statuses, type declarations, and
captures. `Definition`, its `File`, and directory provenance remain the exact
read-only source pointers; the definition tree itself is not cloned.

The outer slice, every inner slice, every step value, and every mutable nested
value must be independently mutable from `ResolvedSteps` and all sibling
runtime steps. In particular, later assignments to request or expected bodies
and `Response.Body` affect only the addressed runtime step.

Preparation is transactional across the complete directory tree. It builds all
candidate matrices before assigning any `RuntimeSteps` field. If traversal,
input validation, or copying fails, every directory keeps its pre-call
`RuntimeSteps`. On success, every directory's existing `RuntimeSteps` is
replaced, including replacement with an empty outer slice.

`Prepare` does not parse variables or JSON, clear the variable store, invoke an
external process, or write output.

### Acceptance criteria

#### AC-Prepare-1: Populate every directory

Given a valid resolved directory tree, when `Prepare` succeeds, then every
directory has a `RuntimeSteps` matrix equal in value and shape to its
`ResolvedSteps` matrix.

#### AC-Prepare-2: Preserve the two-dimensional order

Given multiple resolved step definitions containing multiple steps, when
`Prepare` succeeds, then outer indexes retain definition order, inner indexes
retain declaration order, and empty inner slices retain their positions.

#### AC-Prepare-3: Avoid mutable aliases

Given a prepared tree, when any runtime step, header, variable, response status,
type declaration, or capture is changed, then no resolved step, decoded
definition, or sibling runtime step changes.

#### AC-Prepare-4: Preserve provenance pointers

Given a resolved step with a `Definition` pointer and `Index`, when it is
prepared, then its runtime copy retains that exact pointer and index without
mutating the definition.

#### AC-Prepare-5: Replace previous runtime steps

Given previously populated `RuntimeSteps`, when `Prepare` succeeds again, then
every directory reflects only its current `ResolvedSteps` and contains no stale
or duplicated steps.

#### AC-Prepare-6: Fail transactionally

Given an invalid node or relationship anywhere in the tree, when `Prepare`
fails, then every directory keeps its complete pre-call `RuntimeSteps` value.

#### AC-Prepare-7: Perform no runtime work

Given a valid tree, when `Prepare` runs, then it does not load variables, parse
JSON, invoke curl, jq, or Git, validate a response, or write status output.

## `Execute`

```go
func (s *StepRunner) Execute(
    root *models.Directory,
) error
```

### Input and stage scheduling

`root` is the same hierarchy previously prepared by `Prepare`. Because an empty
runtime matrix is valid, `Execute` does not infer whether preparation occurred
from slice length; the caller owns the `Prepare` then `Execute` ordering.

`Execute` validates the same tree-link invariants as `Prepare`. In addition:

- The root stage must be `0`.
- Every directory stage must be non-negative.
- Every child stage must equal its parent stage plus one.

It first traverses `Children` only to validate the tree, determine the maximum
stage, and group directory pointers by stage. This collection traversal starts
at root and follows each `Children` slice in stored order, but collection order
does not impose an execution order on directories in the same stage.

`Execute` then processes every stage number from `0` through the maximum in
ascending numeric order:

1. Identify every directory whose `Directory.Stage` equals the active stage.
2. Start exactly one goroutine for each identified directory, including a
   directory with no runtime steps.
3. Release the stage's directory goroutines only after all of them have been
   created, so a fast fatal failure cannot prevent another directory goroutine
   required for that active stage from being started.
4. Run all directory goroutines in that stage concurrently.
5. Wait for every started directory goroutine to finish.
6. Start the next stage only when the completed stage had no fatal error.

This wait is a stage barrier. No step in stage `N+1` may begin until every
directory goroutine in stage `N` has returned. Directory goroutines in the same
stage have no defined relative start, request, completion, variable-write, or
output order.

Within each directory goroutine, runtime groups and steps execute strictly
sequentially in this order:

```text
RuntimeSteps[0][0]
RuntimeSteps[0][1]
...
RuntimeSteps[0][n]
RuntimeSteps[1][0]
...
RuntimeSteps[m][n]
```

An empty outer matrix or empty inner group performs no request work. Execution
never consults `ResolvedSteps`. One outer group corresponds to one resolved
steps file; the order already matches ascending cleaned `StepsFiles` paths.
Every step in one group finishes before the next step begins, and every group
finishes before the next group begins. `StepRunner` must not create a separate
goroutine for a group, file, step, phase, or external command.

The scheduling shape is:

```text
stage 0
  -> one directory goroutine for the suite root
       -> RuntimeSteps[0]: all steps sequentially
       -> RuntimeSteps[1]: all steps sequentially
  -> wait

stage 1
  -> one goroutine for child directory A --+
  -> one goroutine for child directory B --+ run concurrently
  -> wait for both

stage 2
  -> one goroutine for every stage-2 directory
  -> wait for all
```

### Variable visibility under parallel execution

The scheduling order defines which prior variable writes a step may depend on:

- A step may use variables written by an earlier step in the same runtime
  group.
- A step may use variables written by any completed earlier group in the same
  directory.
- Every step may use variables written by any completed earlier stage.
- A step must not rely on a variable written by another directory in the same
  stage because those directory goroutines have no defined relative order.

All directory goroutines share the same `VariableProcessor` and run-wide
`KeyValueStore`. Concurrent writes of the same variable key are resolved by the
store's atomic write-once rule: exactly one write may succeed, the original
value must never be overwritten, and every rejected duplicate is a fatal
runtime configuration error that triggers shared cancellation.

### Pointer contract

Every phase must receive the address of the exact element stored in the runtime
matrix. The required iteration form is equivalent to:

```go
for groupIndex := range directory.RuntimeSteps {
    for stepIndex := range directory.RuntimeSteps[groupIndex] {
        step := &directory.RuntimeSteps[groupIndex][stepIndex]
        // Execute every phase with step.
    }
}
```

The implementation must not take the address of a ranged step value:

```go
for _, group := range directory.RuntimeSteps {
    for _, step := range group {
        // Invalid: &step points to a copied Step value.
    }
}
```

The exact pointer is passed to `Load`, `ParseRequest`, `Capture`,
`ParseResponse`, `ValidateTypes`, and `ValidateExpected`. This guarantees that
formatted bodies, curl output, and every later mutation remain visible on the
runtime tree.

### Per-step execution order

For each runtime step, `Execute` performs exactly this order:

```text
VariableProcessor.Load
  -> VariableProcessor.ParseRequest
  -> runner.Curl
  -> Step.Response.Body = curlOutput
  -> VariableProcessor.Capture
  -> VariableProcessor.ParseResponse
  -> Validator.ValidateTypes
  -> Validator.ValidateExpected
  -> render the step result
```

This order supersedes the response-validation order stated in
`07-validator.md`. Type validation runs before expected-value validation. A
nonfatal type mismatch does not skip expected-value validation. A fatal type
error stops the current step before expected-value validation.

#### Load variables

Call `VariableProcessor.Load(ctx, step)`. A non-nil error is fatal. No later
phase for the current step may run. The directory goroutine publishes the
fatal error and cancels the shared context.

#### Parse the request

Call `VariableProcessor.ParseRequest(ctx, step)`. The formatted value written
to `Step.Request.Body` is the exact body supplied to curl. A non-nil error is
fatal and curl must not run.

#### Execute curl

Construct the URL by direct concatenation in this order:

```text
Step.Request.BaseURL + Step.Request.BasePath + Step.Request.Path
```

Do not trim, insert, remove, normalize, or encode slashes. Pass the request
query separately because `runner.Curl` owns appending `?` when it is non-empty.
Invoke:

```go
curlOutput, exitCode, err := runner.Curl(
    ctx,
    step.Request.Method,
    step.Request.BaseURL+step.Request.BasePath+step.Request.Path,
    step.Request.Headers,
    step.Request.Timeout,
    step.Request.Retries,
    step.Request.Query,
    string(step.Request.Body),
)
```

A non-nil curl error is fatal and its exact returned exit code must remain
available from the error returned by `Execute`. `Step.Response.Body` retains
its pre-call value when curl fails.

When curl succeeds, assign its output unchanged before any response phase:

```go
step.Response.Body = curlOutput
```

An empty successful output is assigned as an empty string. `StepRunner` does
not parse, trim, or otherwise interpret curl output.

#### Capture response variables

Call `VariableProcessor.Capture(ctx, step)` after the response body assignment.
Capture occurs before any validation so successfully selected values enter the
run-wide store even when type or expected-value validation later fails. A
capture error is fatal.

#### Parse the response expectation

Call `VariableProcessor.ParseResponse(ctx, step)` after capture. This ordering
allows the current expectation to use a variable captured by the same step. A
parse error is fatal and no validator runs.

#### Validate response types

Call `Validator.ValidateTypes(ctx, step)` and inspect the complete returned
slice. Every nonfatal element is one failed `Step.Response.Types` selector.
Retain every such element in returned order; do not stop at the first failed
selector.

If any element matches `ErrValidatorFatal`, the current step is incomplete and
the directory goroutine publishes that fatal error immediately. `Execute`
returns it after the active stage joins. Nonfatal failures collected before it
must not be rendered as a completed result.

#### Validate the expected response

When type validation produced no fatal error, call
`Validator.ValidateExpected(ctx, step)` even if one or more type selectors
failed. Inspect the complete returned slice for `ErrValidatorFatal` before
treating any element as a mismatch.

A nonfatal expected-value error contains exactly the headerless diff produced
by `runner.GitDiff(ctx, expected, actual)`. Retain it after every type failure
so both validation dimensions are rendered for the step.

### Validation failures and return behavior

Type and expected-value mismatches fail the current step but are not fatal to
suite execution. After rendering the failed result, that directory goroutine
continues with the next step in its current inner slice and then its later
groups. Other directory goroutines in the same stage continue independently,
and later stages run after the stage barrier.

`Execute` records through synchronized state whether at least one completed
step failed validation. After every runtime step has executed:

- Return nil when every step passed.
- Return an error matching `ErrStepValidation` with exit code `101` when one or
  more steps failed type or expected-value validation.

The validation error is returned only after the complete tree runs. It must not
cause a failed step to be rendered twice and need not repeat validation details
already written to output.

Variable errors, curl errors, fatal validator errors, writer failures, invalid
tree input, and internal failures are fatal. A fatal error takes precedence
over validation failures accumulated by earlier completed steps.

### Fatal cancellation and joining

Before starting stage `0`, `Execute` creates one cancellable context shared by
every directory goroutine and external command in the run. Each directory
goroutine checks that context before starting every step and before starting
each external-process phase.

The first non-cancellation fatal error is recorded exactly once and becomes the
original fatal error. Recording it immediately cancels the shared context. On
cancellation:

- No directory goroutine starts another step or external command.
- Every in-flight curl, jq, and Git process receives cancellation, terminates,
  and is waited for.
- All directory goroutines already created for the active stage return.
- `Execute` waits for every one of those goroutines.
- No later stage is started.
- The original fatal error and its exit code are returned.

Cancellation errors produced by sibling goroutines during this shutdown must
not replace the original fatal error. Fatal errors that race after the first
fatal error are retained only as shutdown outcomes and do not replace the
first error or its exit code. A caller cannot cancel `Execute` directly because
its accepted public signature has no context parameter; cancellation here is
the internal mechanism for coordinated fatal shutdown.

A directory goroutine that completed and rendered a step before cancellation
does not roll it back. A step interrupted before all validations and output
finish is incomplete and receives no status line. After every active-stage
goroutine has joined, `Execute` returns the original fatal error.

### Regular output contract

`Execute` writes one status line for every completed step immediately after all
of that step's validations finish. It constructs the complete status-and-detail
block in directory-local memory, then holds an internal output mutex for one
writer call so blocks from concurrent directories cannot interleave. The path
components are:

- Directory: `/` for the supplied root. For a descendant, use its cleaned path
  relative to `root.Path`, convert path separators to `/`, and prefix one `/`.
- Request path: `Step.Request.Path` exactly. Do not include base URL, base path,
  or query.

A passing step writes exactly:

```text
 ✅ <directory> <request-path>
```

A failed step writes exactly:

```text
 ❌ <directory> <request-path>
<type failure 1>
<type failure 2>
...
<headerless expected/actual Git diff>

```

There is exactly one ASCII space before the icon, after the icon, and between
directory and request path. Every status and detail occupies its own line.

For type failures, write every nonfatal error returned by `ValidateTypes`, one
line per failed selector and in the exact returned order. Do not collapse the
slice to its first element. Each line is the error's `Error()` text and
therefore identifies the failed selector, complete declaration, and first
rejected value or empty selection as required by `07-validator.md`.

When expected validation fails, write the complete headerless Git diff after
all type-failure lines. Preserve all embedded diff line endings and write one
empty line after the complete failure details. A type-only failure also ends
with one empty line. A passing status line has no following empty line.

When both validation dimensions fail, render one failed status line, all type
failure lines, and then the expected diff. Do not render separate statuses for
the same step.

Output is plain UTF-8 text with no ANSI escapes. A writer error is fatal,
returned with its cause preserved, and triggers the same cancellation and join
behavior as an external-tool failure.

Within one directory, result blocks retain exact runtime-step order. Between
directories in the same stage, result blocks appear in actual completion order
and no stable relative order is promised. Stage barriers guarantee that every
completed output block from stage `N` precedes every output block from stage
`N+1`.

The implementation must copy this old regular integration golden fixture
byte-for-byte from
`../apihydra-old/testdata/expected-output/normal.txt` to
`testdata/expected-output/normal.txt`:

```text
 ✅ / /root
 ✅ / /repeat
 ✅ /nested /nested
```

The destination includes exactly one trailing newline and no leading blank
line. The StepRunner integration test must execute an equivalent prepared tree
and compare output byte-for-byte with that copied fixture.

An expected-value mismatch uses the old regular diff layout. For example:

```text
 ❌ / /users
@@ -2 +2 @@
-  "id": 1
+  "id": 2

```

### Acceptance criteria

#### AC-Execute-1: Execute the exact runtime elements

Given runtime steps whose request and response members are changed by runtime
phases, when `Execute` succeeds, then every change remains on the corresponding
`Directory.RuntimeSteps` element and no resolved step changes.

#### AC-Execute-2: Preserve sequential order within a directory

Given multiple groups and steps in one directory, when its goroutine runs, then
it executes `[0][0]` before `[0][1]`, finishes one inner group before starting
the next, and never overlaps two curl invocations from that directory.

#### AC-Execute-3: Run phases in the accepted order

Given one step, when execution completes, then observed calls are exactly
`Load`, `ParseRequest`, `Curl`, response-body assignment, `Capture`,
`ParseResponse`, `ValidateTypes`, and `ValidateExpected` in that order.

#### AC-Execute-4: Supply resolved curl arguments

Given a resolved request, when curl runs, then it receives method, concatenated
base URL/base path/path, headers, timeout, retries, query, and the parsed request
body without additional transformation.

#### AC-Execute-5: Assign curl output before response work

Given curl returns a body, when `Capture`, `ParseResponse`, and both validators
run, then each receives the same step pointer and observes that exact body in
`Step.Response.Body`.

#### AC-Execute-6: Capture before parsing expected

Given expected JSON references a variable captured from the current response,
when the step executes, then capture stores the value before `ParseResponse`
looks it up and expected validation receives the substituted expectation.

#### AC-Execute-7: List every failed type selector

Given three independent entries in `Step.Response.Types` fail validation, when
the step result is rendered, then its single failed status is followed by all
three validator error lines in returned order and one trailing empty line.

#### AC-Execute-8: Show the expected/actual Git diff

Given `ValidateExpected` receives different valid documents, when the step
result is rendered, then it contains the headerless output from
`runner.GitDiff(ctx, expected, actual)` with expected `-` lines, actual `+`
lines, and one trailing empty line.

#### AC-Execute-9: Report both validation dimensions

Given type and expected-value validation both fail, when the step completes,
then one failed status is followed by every type failure and the Git diff, and
later steps still execute.

#### AC-Execute-10: Continue after validation failures

Given one step fails validation and later runtime steps exist, when `Execute`
runs, then later steps in that directory, other active-stage directories, and
all later stages execute normally, and the method returns `ErrStepValidation`
with exit code `101` only after every stage completes.

#### AC-Execute-11: Stop on variable or curl failure

Given `Load`, `ParseRequest`, curl, `Capture`, or `ParseResponse` returns an
error, when `Execute` observes it, then the current directory starts no later
phase or step, shared cancellation begins, and the returned original error
preserves the cause and exact non-zero exit code when supplied.

#### AC-Execute-12: Stop on fatal validation failure

Given a validator slice contains `ErrValidatorFatal`, when `Execute` scans the
slice, then it does not render the current step as completed, does not run a
later validator or step in that directory, cancels and joins the active stage,
and returns the fatal error rather than `ErrStepValidation`.

#### AC-Execute-13: Render byte-stable passing output

Given the equivalent of the old behavior-parity runtime tree, when all steps
pass, then output is byte-identical to the copied `normal.txt` fixture and
`Execute` returns nil.

#### AC-Execute-14: Preserve prior completed output on fatal error

Given earlier steps completed before a later fatal error, when `Execute`
returns, then earlier output remains written exactly once and no status line is
written for the incomplete step.

#### AC-Execute-15: Accept empty runtime matrices

Given a valid tree with no runtime steps or with empty inner groups, when
`Execute` runs, then it invokes no runtime dependency, writes no output, and
returns nil.

#### AC-Execute-16: Stop on output failure

Given the output writer fails while rendering a completed result, when
`Execute` observes the failure, then it returns an error preserving the writer
cause, cancels the shared context, and joins every active directory goroutine.

#### AC-Execute-17: Execute stages through barriers

Given directories at stages `0`, `1`, and `2`, when `Execute` runs, then no
stage-1 step starts before every stage-0 directory goroutine finishes and no
stage-2 step starts before every stage-1 directory goroutine finishes.

#### AC-Execute-18: Start exactly one goroutine per active-stage directory

Given several directories in one stage, including a directory with no runtime
steps, when that stage starts, then `Execute` creates exactly one goroutine for
each directory, creates no file or step goroutine, and joins every directory
goroutine before leaving the stage.

#### AC-Execute-19: Run same-stage directories concurrently

Given two directories in the same stage whose first curl calls block, when the
stage runs, then both curl calls can be in flight at the same time while steps
within each individual directory remain sequential.

#### AC-Execute-20: Enforce parallel variable visibility

Given variable dependencies from earlier steps, earlier groups in one
directory, or an earlier stage, when dependent steps run, then those values are
available. Given a dependency on another directory in the same stage, no order
or availability is guaranteed. Concurrent duplicate writes preserve exactly
one original value and trigger fatal cancellation.

#### AC-Execute-21: Cancel and join on the original fatal error

Given one directory reports a fatal tool error while sibling directories have
curl, jq, or Git processes in flight, when fatal shutdown runs, then no new work
starts, every in-flight process is canceled and waited for, every active-stage
directory goroutine is joined, no later stage starts, and `Execute` returns the
original tool error and exit code rather than a sibling cancellation error.

#### AC-Execute-22: Serialize concurrent result blocks

Given multiple same-stage directories complete steps concurrently, when they
write regular output, then each complete status-and-detail block is contiguous,
blocks appear in actual completion order, and no two blocks interleave at byte
boundaries.

## Required tests

Implementation must add mapped unit and integration coverage for every
acceptance criterion in this specification. At minimum, tests must cover:

- Independent request and response parsing with no response-body phase
  inference.
- `runner.GitDiff` match, semantic difference, header removal, expected-first
  signs, command failure, startup failure, cancellation, private permissions,
  cleanup, and shell-metacharacter handling.
- Curl and jq context cancellation without changing their accepted semantic
  exit statuses.
- Transactional deep-copy preparation across multiple directories and matrix
  shapes.
- Actual-element pointer mutation rather than mutation of ranged copies.
- Exact phase call order and curl argument mapping.
- Ascending stage scheduling, complete barriers, exactly one goroutine per
  active-stage directory, and no file or step goroutines.
- Concurrent same-stage directories with sequential work inside each one.
- Guaranteed and intentionally undefined variable-visibility relationships,
  including atomic duplicate assignment under a race.
- Same-step capture use by `ParseResponse`.
- Multiple type failures plus an expected diff on one step.
- Continuation after validation mismatches and coordinated cancel-and-join
  behavior after every fatal phase.
- Preservation of the original fatal error when sibling processes return
  cancellation errors.
- Atomic output blocks and completion-order output from parallel directories.
- Exit code `101` only after all remaining steps complete.
- Byte-exact regular output using the copied old `normal.txt` fixture.
- Empty trees, empty groups, repeated preparation, writer failures, nil input,
  broken parent links, and repeated directory pointers.

Tests must not rely on map iteration order, real network services, or mutation
of `ResolvedSteps`. External-process integration tests may require installed
`curl`, `jq`, and Git; unit tests must isolate sequencing and output so phase
failures can be exercised deterministically.
