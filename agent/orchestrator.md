# APIHydra Orchestrator and Services

## Current implementation status

There is no production `internal/orchestrator` package yet. The full design is
represented by the compileable `internal/orchestrator_mock` package.

The only partially implemented production services are:

- `internal/suite.Loader`, which discovers YAML files.
- `internal/suite.Parser`, which currently returns an almost-empty `Suite` and
  does not parse YAML yet.
- The CLI, which currently only resolves and prints the working directory.

This document therefore describes the intended architecture encoded by the
mock and product requirements, not a completed application.

## What the orchestrator is

The orchestrator is the application coordinator. It does not discover files,
parse YAML, execute curl, or validate responses itself. Instead, it invokes
specialized services and passes each result to the next service.

```text
RunOptions
   |
   v
Loader.Files
   | []string
   v
Parser.Parse
   | *Suite
   v
Selector.Select
   | filtered *Suite
   v
Defaults.Resolve        mutates Suite
   v
Steps.Resolve           mutates Suite
   v
Stages.Plan
   | []Stage
   v
Tools.Check
   v
Run stages sequentially
   `-- run files concurrently
         `-- run steps in each file sequentially
               `-- Runner.Run
   v
Output.Emit events
   v
RunResult
```

The orchestrator constructs and invokes services, transfers shared models,
owns cancellation, applies stage barriers, and decides when results are
presented.

## Orchestrator methods

### `New`

```go
func New(services Services) *OrchestratorMock
```

Input:

- `Services`, containing all directly coordinated service implementations.

Output:

- A configured orchestrator pointer.

State:

- `services`: the service dependency bundle.
- `now`: a clock function, initially `time.Now`.

The clock field allows tests to replace real time with deterministic time.

### `Run`

```go
func (o *OrchestratorMock) Run(
    ctx context.Context,
    options RunOptions,
) RunResult
```

Inputs:

- `ctx`: cancellation and deadline propagation. Fatal execution errors cancel
  in-flight work.
- `options.WorkDir`: suite-root directory.
- `options.Filter`: optional name and label filters.
- `options.JSON`: intended output mode, although the mock does not currently
  use it.

Output:

```go
type RunResult struct {
    Suite    SuiteResult
    ExitCode int
    Err      error
}
```

`Suite` contains completed step results and the aggregate summary. `Err`
represents a fatal error. Ordinary validation mismatches live inside step
results instead of `Err`.

Execution order:

1. Emit `suite_started`.
2. Discover YAML files.
3. Parse them into a suite tree.
4. Filter step documents.
5. Resolve inherited defaults.
6. Resolve runtime steps.
7. Plan stages.
8. Sort stages by depth.
9. Check required tools.
10. Run stages sequentially.
11. Summarize passing and failing steps.
12. Emit the final summary.
13. Return the result and exit code.

State:

`Run` does not retain per-run results after returning. Files, suite tree,
stages, cancellation objects, and results are local to the call. Its injected
services, especially the variable store and output writer, may hold run state.

### `runStage`

```go
func (o *OrchestratorMock) runStage(
    parent context.Context,
    stage Stage,
) ([]StepResult, error)
```

Input:

- Parent context.
- One depth-based execution stage.

Output:

- All completed step results from that stage.
- The first fatal error, if one occurs.

State:

- Child cancellation context.
- Buffered completion channel.
- `sync.WaitGroup` tracking file workers.
- Accumulated results.
- First fatal error.

It starts one goroutine per runnable step file. If one file encounters a fatal
error, it cancels the stage context and waits for every worker to terminate.

### `runFile`

```go
func (o *OrchestratorMock) runFile(
    ctx context.Context,
    file *File,
) fileResult
```

Input:

- Cancellation context.
- A resolved file.

Output:

- Completed step results.
- An optional fatal error.

State:

- A local result slice.

Steps in the file run sequentially. After each successful execution, including
an execution containing validation failures, it emits `step_finished`.

A returned `error` stops the file. Validation failures do not return an error,
so later steps continue.

### Other orchestrator helpers

#### `stageFiles`

```go
func stageFiles(stage Stage) []*File
```

Extracts non-nil files containing runtime steps. It is stateless.

#### `finishSummary`

```go
func (o *OrchestratorMock) finishSummary(suite *SuiteResult)
```

Mutates the supplied result summary by counting passed and failed steps and
setting the end time.

#### `emit`

```go
func (o *OrchestratorMock) emit(ctx context.Context, event Event) error
```

Forwards an event to `Output.Emit`.

#### `failed`

```go
func (o *OrchestratorMock) failed(
    suite SuiteResult,
    err error,
    fallback int,
) RunResult
```

Creates a failed result. If the error implements `ExitCode() int`, that code
overrides the fallback exit code.

## Directly coordinated services

These are the services stored in the `Services` dependency bundle.

## 1. Loader service

```go
Files(ctx context.Context, workDir string) ([]string, error)
```

Purpose:

- Discover regular `.yaml` and `.yml` files recursively.

Input:

- Cancellation context.
- Suite-root path.

Output:

- A list of matching file paths.
- A discovery error if the directory is invalid or traversal fails.

State:

The target mock contract is conceptually stateless because `workDir` is a
method argument.

The current production API is slightly different:

```go
loader := NewLoader(workDir)
files, err := loader.Files()
```

In this implementation, `WorkDir` is constructor state stored on `Loader`.

## 2. Parser service

```go
Parse(
    ctx context.Context,
    workDir string,
    files []string,
) (*Suite, error)
```

Purpose:

- Read YAML files.
- Ignore unrelated YAML.
- Strictly validate APIHydra documents.
- Reject duplicate keys and unknown fields.
- Validate root and defaults placement.
- Create exact `SourceRef` locations.
- Build the directory-oriented `Suite`.

Input:

- Cancellation context.
- Selected suite root.
- Paths discovered by Loader.

Output:

- A parsed `*Suite`.
- A fatal configuration error.

State:

The target contract is conceptually stateless. The returned `Suite` is the new
domain state.

The current production `Parser` stores `WorkDir` and `Files` as constructor
state, but `Parse` is currently only a placeholder.

## 3. Selector service

```go
Select(
    ctx context.Context,
    suite *Suite,
    filter Filter,
) (*Suite, error)
```

Purpose:

- Select step documents by metadata.
- Match `Filter.Name` exactly against metadata name.
- Require all `Filter.Labels` to be present.
- Retain root and defaults documents needed for inheritance.

Output:

- The selected suite tree.
- A configuration error, such as when no steps match.

State:

The service owns no persistent state. Selection is represented by the returned
suite. The sketch does not specify whether the implementation must return a
copy or may mutate the supplied tree.

## 4. Defaults resolver

```go
Resolve(ctx context.Context, suite *Suite) error
```

Purpose:

- Find the nearest parent root or defaults document.
- Resolve directory inheritance.
- Overlay child values over parent values.
- Merge headers case-insensitively.
- Populate `RuntimeDefaults`.

Output:

- `nil` on success or an error.

State:

It explicitly mutates the supplied `Suite`. The resulting state is stored in:

```go
File.ParentDefaults
File.RuntimeDefaults
Directory.RuntimeDefaults
```

The resolver should not retain a separate copy after returning.

## 5. Step resolver

```go
Resolve(ctx context.Context, suite *Suite) error
```

Purpose:

- Overlay step request fields over directory defaults.
- Merge request headers.
- Choose default method, timeout, and retry values.
- Construct the final URL.
- Validate execution-required fields.
- Populate `RuntimeStep` values.

Output:

- `nil` on success or an error.

State:

It mutates the suite by populating each step file's `File.RuntimeSteps`. It
should not own persistent run state.

Variable substitution is generally not completed here because some variables
are created by earlier executing steps.

## 6. Stage planner

```go
Plan(ctx context.Context, suite *Suite) ([]Stage, error)
```

Purpose:

- Group directories by their depth beneath the suite root.

Input:

- Resolved suite tree.

Output:

- Stages containing directory pointers.
- An error if planning fails.

State:

The planner is stateless. Its result describes execution barriers:

```go
type Stage struct {
    Depth       int
    Directories []*Directory
}
```

All files in one stage may run concurrently. The next stage starts only after
the previous stage finishes.

## 7. Tool preflight

```go
Check(ctx context.Context, stages []Stage) error
```

Purpose:

- Inspect selected runtime steps.
- Derive required external executables.
- Check whether they are available before any request executes.

Rules:

- `curl` is always required.
- `jq` is required for bodies, capture, expected values, or types.
- `git` is required for expected-value comparison.

Output:

- `nil` if required tools exist.
- A fatal missing-tool error otherwise.

State:

No persistent application state. The service observes the current process
environment and `PATH`.

## 8. Step runner

```go
Run(ctx context.Context, step RuntimeStep) (StepResult, error)
```

Purpose:

Coordinate the complete lifecycle of one step. Its intended production
implementation composes the variable store, substitutor, request executor, and
response processor.

Typical operation:

1. Serialize and store `step.vars`.
2. Substitute variables into the request body.
3. Validate substituted JSON.
4. Execute the request.
5. Process and validate the response.
6. Store captured variables.
7. Construct timing and result information.

Output distinction:

- `StepResult.Failures`: ordinary assertion mismatches; nonfatal.
- `error`: configuration, substitution, tool, cancellation, or another fatal
  failure.

State:

The runner itself need not own state, but every runner in one suite run must
share the same variable store.

## 9. Output service

```go
Emit(ctx context.Context, event Event) error
```

Purpose:

- Render lifecycle events as human-readable terminal text or NDJSON.

Events can represent:

- `suite_started`.
- `step_finished`.
- `error`.
- `summary`.

Output:

- `nil` when the event is written successfully.
- An error if writing or encoding fails.

State:

Usually yes, but it is I/O state rather than business state. A production
implementation will probably retain a writer and output mode. Because step
events can arrive concurrently, it must serialize writes or otherwise be
concurrency-safe.

## Services composed behind StepRunner

These services are modeled in the package but are not fields of `Services`.
The orchestrator talks to them indirectly through `Runner`.

## 10. Variable store

```go
Set(
    ctx context.Context,
    key string,
    value string,
    source SourceRef,
) error

Get(
    ctx context.Context,
    key string,
    source SourceRef,
) (string, error)
```

This is the main stateful business service.

Its state is one run-wide, in-memory string-to-string map:

```text
map[string]string
```

Semantics:

- It is shared by all concurrent step runners.
- Values are preserved as JSON literal strings.
- Keys are write-once.
- `Set` must atomically reject duplicate keys without changing the original
  value.
- `Get` must distinguish an absent key from a present key containing `null`,
  `0`, `false`, or `""`.
- `SourceRef` identifies where to report an assignment or lookup error; it is
  not part of the key.

Because files run concurrently, the map needs synchronization, normally a
mutex. Its lifetime should be exactly one `Run`: a new suite run should receive
a fresh store.

## 11. Substitutor

```go
Substitute(
    ctx context.Context,
    value YAMLString,
    source SourceRef,
) (YAMLString, error)
```

Purpose:

- Expand variables in request bodies or expected-response JSON.

Rules:

- `$key` inserts the complete stored JSON literal.
- `${key}` removes surrounding JSON quotes when present and embeds the
  contents.
- `$$` produces a literal `$`.

Output:

- Substituted text.
- A fatal missing-variable or substitution error.

State:

It does not need to own mutable state, but it is state-dependent because it
reads the shared variable store. That dependency is implicit in the mock
signature and would need to be provided through the production constructor.

## 12. Request executor

```go
Execute(
    ctx context.Context,
    step RuntimeStep,
) (HTTPResponse, error)
```

Purpose:

- Prepare curl arguments.
- Validate a substituted body with `jq` when needed.
- Execute curl.
- Capture HTTP status, headers, and body.

Output:

```go
type HTTPResponse struct {
    StatusCode int
    Headers    map[string][]string
    Body       []byte
}
```

State:

Normally stateless. It performs external side effects and must honor context
cancellation. Tool exit codes should be preserved in returned errors.

## 13. Response processor

```go
Process(
    ctx context.Context,
    step RuntimeStep,
    response HTTPResponse,
) (
    captured map[string]string,
    failures []ValidationFailure,
    err error,
)
```

Purpose:

- Validate response JSON when required.
- Run `response.capture` jq expressions.
- Compare accepted HTTP statuses.
- Substitute variables into `response.expected`.
- Compare expected and actual JSON using Git diff.
- Evaluate type assertions.

Outputs:

- `captured`: successfully extracted key/value pairs.
- `failures`: nonfatal status, expected-value, and type mismatches.
- `error`: fatal invalid configuration or external-tool failure.

State:

It may depend on the shared variable store because same-step captures must
become visible to that step's `response.expected`. The mock does not settle
whether the processor writes captures itself or coordinates that write through
StepRunner. That production contract still needs clarification.

## Important state carriers

Services move data through distinct lifecycle states:

```text
Step
  `-- declarative YAML state

RuntimeStep
  `-- defaults resolved, concrete request configuration

StepResult
  `-- response, captures, failures, and timing
```

The suite tree becomes progressively enriched:

```text
Parser
  -> parsed File/Directory tree

Defaults.Resolve
  -> ParentDefaults and RuntimeDefaults added

Steps.Resolve
  -> RuntimeSteps added

Runner
  -> StepResults produced
```

`Value[T]` records whether a YAML value was present:

```go
type Value[T any] struct {
    Value T
    Set   bool
}
```

This matters because an omitted value and a value explicitly set to zero or an
empty value have different inheritance and validation semantics.

## Concurrency and state visibility

The required execution order is:

- Stages are sequential.
- Files within a stage are concurrent.
- Steps within one file are sequential.

Consequently:

- A step can safely consume variables created by an earlier step in the same
  file.
- A step can safely consume variables from any completed earlier stage.
- A step cannot safely depend on a variable produced by another concurrently
  executing file in the same stage.
- Duplicate concurrent writes must be detected atomically.
- A fatal error cancels concurrent work.
- A validation failure is stored in `StepResult.Failures` and does not cancel
  work.

## Current inconsistencies and gaps

There are several differences between the mock and the product requirements:

- The mock returns validation exit code `1`; the PRD requires `101`.
- The mock uses internal error code `4`; the PRD requires `3`.
- `RunOptions.JSON` is currently unused.
- `EventError` exists, but `Run` does not emit it on fatal errors.
- Fatal paths generally do not emit a final summary event.
- The variable store's ownership and construction are not represented in
  `Services`.
- Same-step capture visibility needs a clearer contract between
  `ResponseProcessor` and `StepRunner`.
- The production parser and orchestrator have not been implemented.
