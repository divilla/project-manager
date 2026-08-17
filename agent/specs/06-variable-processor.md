# internal.runtime.VariableProcessor Service

## Status

- Service: `internal/executor.VariableProcessor`
- Package: `internal/executor`
- Package name: `executor`
- Shared models: `internal/domain`
- Variable store: `internal/executor.KeyValueStore`
- Command runner: `pkg/runner`
- Status: implementation specification

## Base specification

### Purpose

`VariableProcessor` owns runtime variable handling for one step. It loads
literal `Step.Vars` values into the run-wide `KeyValueStore`, substitutes stored
values into `Step.Request.Body` and `Step.Response.Expected`, validates and
formats those JSON members, and captures response values through jq.

One processor receives the run-wide `KeyValueStore` as a constructor
dependency. Every processor participating in the same suite run must use the
same store.

### Responsibilities

`VariableProcessor` owns:

- Serializing step variable values as compact JSON.
- Writing literal and captured variables to `KeyValueStore`.
- Replacing `$key` and `${key}` references in request and expected JSON.
- Replacing `$$` with a literal `$`.
- Reporting missing variable references.
- Validating substituted request and expected values as JSON.
- Recursively ordering object members and pretty-formatting both values.
- Applying every `Step.Response.Capture` selector to `Step.Response.Body` with
  `runner.JQFilter`.
- Returning fatal errors with the exit code the application must forward.

### Non-responsibilities

`VariableProcessor` does not:

- Own the variable map or implement write-once synchronization.
- Validate variable-key syntax in decoded definitions.
- Execute HTTP requests or populate `Step.Response.Body`.
- Compare expected and actual responses.
- Validate response status or type declarations.
- Print errors, call `os.Exit`, or terminate goroutines itself.
- Create a package-global variable store.

### Public contract

```go
type VariableProcessor struct {
    keyValueStore *variable.KeyValueStore
}

func NewVariableProcessor(
    keyValueStore *variable.KeyValueStore,
) *VariableProcessor

func (p *VariableProcessor) Load(
    ctx context.Context,
    step *models.Step,
) (int, error)

func (p *VariableProcessor) ParseRequest(
    ctx context.Context,
    step *models.Step,
) (int, error)

func (p *VariableProcessor) ParseResponse(
    ctx context.Context,
    step *models.Step,
) (int, error)

func (p *VariableProcessor) Capture(
    ctx context.Context,
    step *models.Step,
) (int, error)
```

All four operations return `(0, nil)` on success. A non-nil error is fatal to
the suite run. The caller must stop execution and return the accompanying exit
code; the processor does not terminate the process directly.

### Dependency and lifetime

`NewVariableProcessor` retains the supplied `*variable.KeyValueStore` without
copying it. The dependency must be non-nil.

`VariableProcessor` has no mutable state besides the state owned by its shared
store. It must not retain a context, step, selector result, or error after a
method returns.

### Exit-code contract

- Successful operations return exit code `0` and nil.
- Invalid variable data, a missing variable, or a duplicate assignment returns
  exit code `2` and a non-nil runtime configuration error.
- A jq selector or JSON-processing failure returns the exact non-zero exit code
  supplied by `runner` and its error unchanged or wrapped with step-member
  context while preserving the underlying error.
- A command-start error with no process exit status returns the code supplied by
  `runner` and its non-nil `runner.CommandError`.
- Context cancellation returns a non-nil context error and must not be reported
  as a successful operation.

## `NewVariableProcessor`

```go
func NewVariableProcessor(
    keyValueStore *variable.KeyValueStore,
) *VariableProcessor
```

### Behavior

`NewVariableProcessor` returns a processor backed by the supplied run-wide
store. It does not allocate another store, read variables, or invoke jq.

### Acceptance criteria

#### AC-NewVariableProcessor-1: Retain the supplied store

Given a `KeyValueStore`, when `NewVariableProcessor` is called, then subsequent
`Load` and `Capture` writes are visible through that exact store.

## `Load`

```go
func (p *VariableProcessor) Load(
    ctx context.Context,
    step *models.Step,
) (int, error)
```

### Behavior

`Load` processes every key-value pair in `step.Vars`. It serializes each value
as compact JSON and calls `KeyValueStore.Set` with the key and serialized
string.

Examples include:

```text
YAML 1       -> 1
YAML true    -> true
YAML Test    -> "Test"
YAML [1, 2]  -> [1,2]
YAML null    -> null
```

An empty `Vars` map is a successful no-op. `Load` does not remove entries from
`step.Vars` or replace their decoded values.

If serialization fails, `Load` returns exit code `2` and an error identifying
the variable key. If `KeyValueStore.Set` rejects a duplicate, `Load` returns
exit code `2`, preserves the store error for `errors.Is`, and stops processing
the step. The store preserves the original value.

### Acceptance criteria

#### AC-Load-1: Store compact JSON values

Given JSON-compatible values in `Step.Vars`, when `Load` succeeds, then every
key is present in the shared store with its compact JSON representation.

#### AC-Load-2: Accept no variables

Given an empty or nil `Step.Vars`, when `Load` is called, then it returns `0`
and nil without changing the store.

#### AC-Load-3: Reject duplicate assignments

Given a key already exists in the shared store, when `Load` attempts to assign
that key, then it returns exit code `2` and the duplicate error, leaves the
original value unchanged, and performs no further assignments for the step.

#### AC-Load-4: Reject an unserializable value

Given a step variable that cannot be represented as JSON, when `Load` reaches
that value, then it returns exit code `2`, identifies the key, and does not
write that value.

## `ParseRequest` and `ParseResponse`

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

### Variable replacement

Both methods apply the PRD variable syntax to their selected member.
`ParseRequest` processes only `Step.Request.Body`; `ParseResponse` processes
only `Step.Response.Expected`.

- `$key` inserts the exact stored string as a complete JSON literal.
- `${key}` removes one leading and one trailing double quote when both are
  present on the stored string, then inserts the remaining text. A stored value
  without both outer quotes is inserted unchanged.
- `$$` inserts one literal `$` and does not begin a variable reference.
- Keys match `[A-Za-z_][A-Za-z0-9_:]*`.

Every lookup uses `KeyValueStore.Get`. A missing key returns exit code `2` and
an error identifying the key and the affected step member. Present values such
as `null`, `false`, `0`, or an empty JSON string must not be treated as missing.

An omitted selected member is left unchanged. Neither method alters the other
member, regardless of the value of `Step.Response.Body`.

### JSON validation, ordering, and formatting

After replacing all references in the selected member, each method must:

1. Use terminal jq to validate the complete value as JSON.
2. Recursively order every object's members alphabetically.
3. Render the result as pretty-formatted JSON.
4. Replace the corresponding `Step.Request.Body` or
   `Step.Response.Expected` value with the formatted result.

The jq filter used for ordering must recursively sort object entries while
preserving array order and scalar values. The methods may use
`runner.JQFilter(ctx, selector, input)` for that validation and ordering, then
render the returned JSON with two-space indentation. Compact jq output is an
intermediate value only; the value written back to the step is
pretty-formatted.

Arrays preserve element order. Object ordering applies at every nesting depth.
Top-level strings, numbers, booleans, arrays, and `null` remain valid and are
formatted without changing their JSON value.

If jq rejects a value, the method returns its exact exit code and error. It
must identify whether request body or response expected failed. An invalid
request body must never be sent by the request executor.

The step runner invokes request-body parsing before request execution. It
invokes expected-response parsing after `Capture`, so a same-step captured
value is available for substitution.

### Mutation

Each member is replaced only after its substitution, validation, ordering, and
formatting all succeed. A failed member retains its value from before that
member's parse began.

### Acceptance criteria

#### AC-Parse-1: Replace JSON-literal references

Given `change_id` stores `1` and a request body contains `$change_id`, when the
body is parsed, then `1` is inserted as a JSON number rather than a quoted
string.

#### AC-Parse-2: Replace embedded references

Given `date` stores `"2026-01-01"` and a JSON string contains `${date}`, when
the member is parsed, then the stored outer quotes are removed and the date is
inserted into the surrounding JSON string.

#### AC-Parse-3: Escape a dollar sign

Given `$$` in a body or expected value, when that member is parsed, then the
result contains one literal `$` and performs no variable lookup for it.

#### AC-Parse-4: Reject a missing variable

Given a body or expected value references an absent key, when its parse method
reaches the reference, then it returns exit code `2`, identifies the key and
member, and leaves that member unchanged.

#### AC-Parse-5: Validate, order, and format request body

Given a valid request body with unsorted nested object members, when parsing
succeeds, then `Step.Request.Body` contains recursively ordered,
pretty-formatted JSON before request execution.

#### AC-Parse-6: Validate, order, and format expected response

Given a valid expected value with unsorted nested object members, when parsing
succeeds after capture, then `Step.Response.Expected` contains recursively
ordered, pretty-formatted JSON.

#### AC-Parse-7: Forward invalid JSON failures

Given substitution produces invalid JSON, when jq rejects the affected member,
then the parse method returns jq's exact exit code and an error identifying
that member.

#### AC-Parse-8: Use same-step captures in expected

Given `response.capture` creates a variable referenced by
`response.expected`, when `Capture` succeeds before expected parsing, then
`ParseResponse` resolves that variable from the shared store.

#### AC-Parse-9: Parse only the request

Given a step with request and expected variable references, when
`ParseRequest` succeeds, then it updates only `Step.Request.Body` and leaves
`Step.Response.Expected` unchanged regardless of the current response body.

#### AC-Parse-10: Parse only the response

Given a step with request and expected variable references, when
`ParseResponse` succeeds, then it updates only `Step.Response.Expected` and
leaves `Step.Request.Body` unchanged regardless of whether the response body is
empty.

#### AC-Parse-11: Remove phase inference

Given any response-body value, when either parse method runs, then its selected
member is determined only by the method name and never by
`Step.Response.Body`.

## `Capture`

```go
func (p *VariableProcessor) Capture(
    ctx context.Context,
    step *models.Step,
) (int, error)
```

### Behavior

For every key-selector pair in `Step.Response.Capture`, `Capture` calls:

```go
runner.JQFilter(ctx, string(selector), step.Response.Body)
```

When `JQFilter` succeeds, `Capture` calls `KeyValueStore.Set` with the capture
key and jq output. The output is stored unchanged.

An empty `Capture` map is a successful no-op. `Capture` does not independently
validate whether jq output can later be embedded in another JSON document.

If `JQFilter` returns an error, `Capture` immediately returns its exit code and
an error preserving the runner error. The caller must stop the suite and
forward that exit code. The failed key is not written.

If `KeyValueStore.Set` rejects a duplicate capture key, `Capture` returns exit
code `2`, preserves the duplicate error, and the original stored value remains
unchanged.

### Acceptance criteria

#### AC-Capture-1: Capture a response value

Given a capture key and selector and a JSON value in `Step.Response.Body`, when
`JQFilter` succeeds, then its output is stored unchanged under the capture key
and `Capture` returns `0` and nil.

#### AC-Capture-2: Capture false and null

Given a selector produces `false` or `null`, when `JQFilter` accepts jq exit
code `1`, then `Capture` stores the returned literal output and succeeds.

#### AC-Capture-3: Capture multiple jq results

Given a selector emits an object, array, or multiple JSON results, when
`JQFilter` succeeds, then `Capture` stores its complete returned output without
rejecting or reformatting it.

#### AC-Capture-4: Forward jq failure

Given `JQFilter` returns a selector error and a non-zero exit code, when
`Capture` receives it, then it writes no value for that key, returns the same
exit code, preserves the runner error, and signals the caller to terminate the
suite.

#### AC-Capture-5: Reject a duplicate capture

Given a capture key already exists, when jq succeeds and the store rejects the
write, then `Capture` returns exit code `2`, preserves the original stored
value, and returns the duplicate error.

#### AC-Capture-6: Accept no captures

Given an empty or nil `Step.Response.Capture`, when `Capture` is called, then it
returns `0` and nil without invoking jq or changing the store.

## Concurrent behavior

Different step runners may call one or more processors concurrently. Processor
calls rely on `KeyValueStore` for atomic write-once storage and must not add
shared mutable state outside that store.

### Acceptance criteria

#### AC-Concurrency-1: Share the run-wide store safely

Given concurrent processors using the same `KeyValueStore`, when they load,
parse, or capture variables, then store access remains race-free and duplicate
writes retain the store's atomic first-write-wins behavior.

## Required tests

At minimum, tests must cover:

- Independent request and response parsing with no response-body phase
  inference.
- Same-step capture use by `ParseResponse`.
