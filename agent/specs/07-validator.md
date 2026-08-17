# internal.runtime.Validator Service

## Status

- Service: `internal/runtime.Validator`
- Package: `internal/runtime`
- Package name: `runtime`
- Shared models: `internal/models`
- External tools: terminal `jq` and Git
- Status: implementation specification

## Base specification

### Purpose

`Validator` validates the actual response stored in a runtime `models.Step`
against `Step.Response.Expected` and `Step.Response.Types`.

Expected-value validation treats a top-level expected object as a partial
assertion, projects the actual response to that object's declared members, and
compares the resulting pretty JSON documents with Git. Type validation applies
each configured jq selector to the actual response and checks the selected JSON
values against the declared base type, `zero`, and `optional` rules.

Both methods return validation failures instead of printing them or stopping
the suite. They preserve fatal context, process, and external-tool errors in the
same return slice so the caller can distinguish those errors and stop the suite
when required by the PRD.

### Responsibilities

`Validator` owns:

- Validating response JSON whenever expected-value or type validation is used.
- Projecting an actual object to the members declared by an expected object.
- Preserving the difference between a missing expected member and a member
  explicitly containing `null`.
- Comparing pretty-formatted expected and actual JSON with `runner.GitDiff`.
- Applying every `Step.Response.Types` jq selector to `Step.Response.Body`.
- Validating every selected value against its base type and modifiers.
- Returning all independently discoverable type assertion failures for a step.
- Preserving external-tool exit codes and context cancellation in fatal errors.

### Non-responsibilities

`Validator` does not:

- Validate HTTP status declarations or the actual HTTP status.
- Decode YAML or validate the structure of a type declaration. The Decoder
  already guarantees one supported base type followed only by unique `zero`
  and `optional` modifiers.
- Substitute variables, validate or format substituted expected JSON, or
  capture response variables. Those operations belong to `VariableProcessor`.
- Execute the HTTP request or populate `Step.Response.Body`.
- Decide the suite's final exit code.
- Print failures, render diffs, call `os.Exit`, or cancel sibling work.
- Retain a context, step, selected value, temporary path, or error after a call.

### Public contract

```go
var ErrValidatorFatal = errors.New("fatal validator error")

type Validator struct{}

func (v *Validator) ValidateTypes(
    ctx context.Context,
    step *models.Step,
) []error

func (v *Validator) ValidateExpected(
    ctx context.Context,
    step *models.Step,
) []error
```

The zero value of `Validator` is ready for use. No constructor or mutable
service state is required.

Each method passes its received context to every jq process it starts.
`ValidateExpected` also passes that context to `runner.GitDiff`.

The method return type is `[]error`, not `[]errors`; `error` is Go's built-in
interface. A successful or inapplicable validation returns nil.

`ErrValidatorFatal` is the public classification sentinel for errors that must
stop the suite. Every fatal error returned by either method must match it
through `errors.Is`. A nonfatal assertion failure must not match it.

### Input contract

The caller supplies a non-nil runtime step after request execution:

- `Step.Response.Body` contains curl's response body text.
- `Step.Response.Expected`, when present, has already passed variable
  substitution, jq validation, recursive object ordering, and pretty
  formatting through `VariableProcessor.ParseResponse`.
- `Step.Response.Types` has already passed Decoder declaration validation.

Passing a nil step is a caller error. The methods must return a non-nil fatal
error and must not panic.

### Return and failure contract

The slice distinguishes two classes by the errors it contains:

- An expected-value difference or rejected type assertion is a nonfatal step
  validation failure. It contains no wrapped process error. The step runner
  records it, continues the other response validations, and allows the suite to
  continue.
- Context cancellation, command startup failure, invalid required response
  JSON, jq failure, Git failure other than diff status `1`, temporary-file
  failure, or other infrastructure failure is fatal. The returned error wraps
  `ErrValidatorFatal` and its cause. An external process error remains
  discoverable with `errors.As`, including its exact `ExitCode()`; context
  errors remain discoverable with `errors.Is`.

Git exit status `1` has its defined diff meaning and must be converted into a
nonfatal expected-value validation error rather than returned as an operational
process failure. Git exit statuses greater than `1` are fatal.

A method stops its own work at the first fatal error. `ValidateTypes` may
return assertion failures collected before that fatal error followed by the
fatal error. The caller must scan the complete returned slice for a fatal error
before treating its entries only as validation failures:

```go
if errors.Is(err, runtime.ErrValidatorFatal) {
    // Stop the suite and derive the required exit code from the wrapped cause.
}
```

Exact final human and JSON-event formatting is outside this service and remains
deferred by the PRD. Every assertion failure must nevertheless identify:

- Whether it is an expected-value or type failure.
- The failed jq selector for a type failure.
- The complete type declaration.
- The first rejected value as compact JSON, or that no value was selected.
- The Git diff for an expected-value difference.

## `ValidateExpected`

```go
func (v *Validator) ValidateExpected(
    ctx context.Context,
    step *models.Step,
) []error
```

### Applicability

An empty `Step.Response.Expected` means no expected-value assertion.
`ValidateExpected` returns nil without invoking jq or Git. JSON string `""` is
not an omitted expectation: after parsing it is the non-empty JSON text `"\"\""`
and must be compared normally.

### JSON validation and canonical form

When expected is present, `ValidateExpected` uses terminal jq to validate both
`Step.Response.Expected` and `Step.Response.Body` as exactly one JSON value.
It must not accept identical invalid strings through a text-equality shortcut.

Both comparison inputs use jq's recursively sorted, pretty-formatted output.
Object keys are sorted at every nesting depth. Array order and scalar JSON
values are preserved.

If jq rejects expected or actual JSON, the method returns one fatal error that
wraps jq's process error. The caller stops the suite and forwards jq's exact
non-zero exit code. In particular, invalid actual JSON is not an ordinary
comparison mismatch.

The old implementation's strict single-value filter is an appropriate starting
point:

```go
const singleJSONFilter =
    `if length == 1 then .[0] else error("invalid json") end`

// Equivalent direct invocation; no shell is used.
jq -S --slurp singleJSONFilter
```

The jq output, rather than Go map serialization, is the canonical comparison
document.

### Object projection

When the expected top-level JSON value is an object, it is a partial assertion:

- Only members declared by expected participate in comparison.
- Additional actual members are discarded at every projected object depth.
- Every expected member must exist in actual.
- A missing actual member is omitted from the projection so it differs from an
  expected member explicitly containing `null`.
- An expected empty object matches any actual object and does not match a
  non-object value.
- Values at corresponding non-object positions retain normal JSON equality;
  JSON type differences fail.

The projection descends through declared object members. Arrays are values, not
object-member sets: their complete contents and element order are compared.
Objects reached through an array are not independently projected. This avoids
introducing array-prefix subset behavior that the PRD does not define.

This jq filter is adapted from `../apihydra-old` and illustrates the required
missing-versus-null behavior:

```jq
def comparison_shape:
  if type == "object" then {}
  elif type == "array" then []
  else .
  end;

def project($expected):
  if ($expected | type) == "object" then
    if type != "object" then comparison_shape
    else
      . as $actual |
      reduce ($expected | keys_unsorted[]) as $key
        ({};
          if ($actual | has($key)) then
            . + {($key): ($actual[$key] | project($expected[$key]))}
          else .
          end)
    end
  else .
  end;

project($expected[0])
```

The expected document is supplied to jq as data, such as with `--slurpfile`,
and must never be interpolated into the jq program.

### Whole-value comparison

When expected is a top-level array, string, number, boolean, or `null`, the
entire canonical actual response is compared with expected. No projection or
prefix matching is applied.

### Git comparison

After producing the final expected and comparison-actual documents,
`ValidateExpected` calls:

```go
runner.GitDiff(ctx, expected, actual)
```

An empty returned diff means validation passes. A non-empty returned diff
becomes the single nonfatal expected-value validation error, and the error's
presentation text must be exactly that headerless diff. A non-nil runner error
is fatal under the `ErrValidatorFatal` contract.

The reported diff compares expected against projected actual. Members removed
by object projection must not appear in it. `ValidateExpected` does not own
temporary comparison files or invoke Git directly, and it must not color the
diff.

### Mutation

`ValidateExpected` is read-only. It must not replace or reformat
`Step.Response.Expected` or `Step.Response.Body`.

### Acceptance criteria

#### AC-ValidateExpected-1: Pass the PRD object-subset example

Given actual:

```json
{"id":1,"name":"A","active":true}
```

and expected:

```json
{"id":1}
```

when `ValidateExpected` runs, then it projects actual to `{"id":1}`, compares
the pretty canonical documents with Git, and returns nil.

#### AC-ValidateExpected-2: Ignore nested extra object members

Given expected `{"x1":{"y1":1,"y2":2},"x2":2}` and actual
`{"x1":{"y2":2,"y1":1,"y3":3},"x2":2,"x3":false}`, when validation runs,
then member order and formatting are ignored, `y3` and `x3` are projected out,
and validation passes.

#### AC-ValidateExpected-3: Return the projected Git diff

Given actual `{"id":2}` and expected `{"id":1}`, when Git reports a diff, then
the method returns one nonfatal expected-value error containing that diff.

Given the nested example from AC-ValidateExpected-2 with actual `y2` equal to
the JSON string `"2"`, the diff contains expected numeric `2` and actual string
`"2"`, and contains neither `y3` nor `x3`.

#### AC-ValidateExpected-4: Distinguish missing from null

Given expected `{"deleted_at":null}` and actual `{}`, when validation runs,
then the projected actual omits `deleted_at`, Git reports a difference, and the
method returns one nonfatal failure.

Given actual `{"deleted_at":null,"extra":true}`, the same expectation passes.

#### AC-ValidateExpected-5: Handle an empty expected object

Given expected `{}`, when actual is either `{}` or a non-empty object, then
validation passes. When actual is an array, scalar, or `null`, validation fails.

#### AC-ValidateExpected-6: Compare non-object expectations as whole values

Given a top-level array, string, number, boolean, or `null` expectation, when
validation runs, then it compares the entire actual JSON value. Extra array
elements, changed element order, and JSON type differences fail.

#### AC-ValidateExpected-7: Forward invalid expected JSON

Given invalid expected JSON, when jq validates it, then the method returns one
fatal jq error and preserves jq's exact exit code. It does not invoke Git.

#### AC-ValidateExpected-8: Forward invalid actual JSON

Given expected validation is configured and `Step.Response.Body` is empty,
non-JSON, or contains more than one JSON value, when jq validates the response,
then the method returns one fatal jq error, does not invoke Git, and does not
convert the error into a step assertion failure.

#### AC-ValidateExpected-9: Accept no expected assertion

Given an omitted expected value, when `ValidateExpected` runs, then it returns
nil without reading response JSON or invoking an external command.

#### AC-ValidateExpected-10: Preserve cancellation and Git failures

Given a canceled context or Git exit status other than `0` or `1`, when the
method returns, then it stops work, returns one fatal error, and preserves the
context identity or exact Git exit status.

## `ValidateTypes`

```go
func (v *Validator) ValidateTypes(
    ctx context.Context,
    step *models.Step,
) []error
```

### Applicability and selector order

An empty or nil `Step.Response.Types` map means no type assertions.
`ValidateTypes` returns nil without validating the response or invoking jq.

When at least one type assertion exists, `ValidateTypes` first runs the strict
single-value jq validation described by `ValidateExpected`. It supplies
`Step.Response.Body` to `jq -S --slurp` with `singleJSONFilter` and uses the
resulting single canonical JSON document as selector input. Empty input,
invalid JSON, and multiple top-level JSON values are fatal before any type
selector runs.

For deterministic behavior, selectors are evaluated in ascending bytewise
lexicographic order. The declaration array for each selector retains its
decoded order. The first item is the base type; later items set `zero` and
`optional` independently and may appear in either order.

### jq selection

For every selector, `ValidateTypes` invokes terminal jq directly with the
actual response on standard input and the equivalent arguments:

```text
jq -c -- <selector>
```

No shell is used. The complete selector is passed as one jq program argument.
The validator does not parse, sanitize, classify, or maintain an allowlist of
jq syntax.

The implementation must not use `jq -e` for this operation. A successful
selector that emits no values must remain distinguishable from a jq process
failure, while selected `false` and `null` values must remain ordinary output.

Output is consumed as a stream with `json.Decoder` and `UseNumber`, as in the
old validator:

```go
decoder := json.NewDecoder(output)
decoder.UseNumber()
for {
    var value any
    err := decoder.Decode(&value)
    if errors.Is(err, io.EOF) {
        break
    }
    if err != nil {
        // Fatal invalid jq output.
    }
    // Validate value without converting json.Number to float64.
}
```

Each jq output value is validated against the same declaration. A selector
passes only when all emitted values pass. On the first rejected value for one
selector, the method records one failure for that selector and may stop
consuming that selector's remaining output because later values cannot restore
the assertion. Validation then continues with the next selector.

When streaming is stopped at a rejected value, the implementation must
terminate and wait for that jq process without treating its deliberate stop as
an operational failure. The private sentinel pattern from the old validator is
an appropriate implementation example:

```go
var errStopSelection = errors.New("stop jq selection")

// The output consumer returns errStopSelection after recording the first
// rejected value. The command runner kills and waits for jq, and the validator
// converts only that sentinel into the nonfatal selector failure.
```

If a selector emits no values, it passes only with `optional`; otherwise one
failure identifies the empty selection. If jq fails to start, rejects the
selector, fails while evaluating it, or is canceled, the method stops and
returns a fatal error preserving the cause and exact process exit code.

### Declaration semantics

The supported base types are:

```text
string
number
integer
boolean
object
array
null
datetime
uuid
```

With no modifier, a selected member must be present, non-null, of the declared
base type, and not that type's zero value. `null` is the exception described
below.

The modifiers behave independently:

- `zero` permits the base type's zero value without changing its type rule.
- `optional` permits zero emitted values or an emitted `null`. A non-null value
  must still have the base type and must still be nonzero unless `zero` is also
  present.

For example:

```text
[integer]                 rejects absent, null, 0, and non-integers
[integer, zero]           accepts 0 and every other integer
[integer, optional]       accepts absent or null, rejects 0
[integer, zero, optional] accepts absent, null, 0, or another integer
```

`null` is special:

- `[null]` requires at least one emitted value and requires every emitted value
  to be JSON `null`.
- It does not require `zero`.
- `zero` is permitted with `null` but does not change its behavior.
- `optional` also permits no emitted values.

### Base types and zero values

| Base type | Accepted non-null JSON value | Zero value |
| --- | --- | --- |
| `string` | JSON string | `""` |
| `number` | Any JSON number | Any numeric representation equal to `0` |
| `integer` | A JSON number whose exact mathematical value is an integer | Any numeric representation equal to `0` |
| `boolean` | JSON boolean | `false` |
| `object` | JSON object | `{}` |
| `array` | JSON array | `[]` |
| `null` | JSON `null` | Special; accepted without `zero` |
| `datetime` | Accepted datetime string | Any accepted representation of `0001-01-01T00:00:00Z` |
| `uuid` | Canonical hyphenated UUID string | `00000000-0000-0000-0000-000000000000` |

Integer validation must use the exact `json.Number` token rather than a
`float64` conversion. Decimal and exponent spellings pass when their exact
mathematical value is integral; fractional values fail. The `math/big` based
`integerSign` approach from `../apihydra-old` is an appropriate implementation
example because it avoids range and precision loss.

An object is zero only when it has no members. An array is zero only when it
has no elements.

### Datetime validation

`datetime` accepts only strings matching one of these forms:

```text
YYYY-MM-DD
YYYY-MM-DDZ
YYYY-MM-DD±HH:MM
YYYY-MM-DDTHH:MM:SSZ
YYYY-MM-DDTHH:MM:SS±HH:MM
YYYY-MM-DDTHH:MM:SS.sssZ
YYYY-MM-DDTHH:MM:SS.sss±HH:MM
YYYY-MM-DDTHH:MM:SS.ssssssZ
YYYY-MM-DDTHH:MM:SS.ssssss±HH:MM
```

Calendar and clock components must form a real Gregorian date and time. Offset
hours range from `00` through `23` and minutes from `00` through `59`. A value
with a time component must include `Z` or an offset. Fractional seconds contain
exactly three or six digits.

Date-only values represent midnight. A date-only value without a zone is
interpreted as UTC for zero-value comparison. Zoned values are converted to
their UTC instant. The datetime zero value is any accepted spelling whose
instant equals `0001-01-01T00:00:00Z`.

The implementation may adapt the old validator's pattern-plus-calendar-check
approach. The pattern must be expanded for the PRD's date-only forms and
restricted fractional precision; matching the regular expression alone is not
enough to accept an impossible date.

### UUID validation

`uuid` accepts a JSON string matching this case-insensitive canonical shape:

```go
var uuidPattern = regexp.MustCompile(
    `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)
```

Any UUID version is accepted. Braceless non-hyphenated strings, malformed
groups, and non-string JSON values fail. The all-zero UUID requires `zero`.

### Error collection and mutation

`ValidateTypes` returns at most one validation failure per selector, in
selector evaluation order. It continues after a rejected selector so failures
from independent selectors can be reported for the same step.

The method is read-only. It must not alter `Step.Response.Types` or
`Step.Response.Body`.

### Acceptance criteria

#### AC-ValidateTypes-1: Pass the PRD type-only example

Given `".id": [number]` and actual `{"id":7}`, when `ValidateTypes` runs, then
the assertion passes regardless of the number's specific nonzero value.

#### AC-ValidateTypes-2: Validate multiple selected values

Given `".items[].id": [integer]`, when the selector emits several values, then
all emitted values must be nonzero integers. A non-integer or zero causes one
failure identifying the selector and first rejected value.

#### AC-ValidateTypes-3: Apply zero and optional together

Given `".change_id": [integer, zero, optional]`, when `change_id` is absent,
`null`, `0`, or a nonzero integer, then validation passes. When it is present
with a value of another type, validation fails.

#### AC-ValidateTypes-4: Distinguish empty selection and null

Given a selector that emits no values or emits `null`, when `optional` is
absent, then validation fails except that `[null]` accepts an emitted null.
When `optional` is present, no output and emitted null both pass.

#### AC-ValidateTypes-5: Enforce zero values for every base type

Given a declaration without `zero`, when the selector emits `""`, numeric `0`,
`false`, `{}`, `[]`, the zero datetime instant, or the all-zero UUID for its
corresponding base type, then validation fails. Adding `zero` makes the same
value pass without accepting a different JSON type.

Given `[null]`, an emitted JSON `null` passes without `zero`.

#### AC-ValidateTypes-6: Distinguish number and integer

Given `[number]`, JSON integer, decimal, and exponent numbers pass when nonzero.
Given `[integer]`, only numbers with an exact integral mathematical value pass;
strings, fractional values, booleans, and null fail.

#### AC-ValidateTypes-7: Validate objects and arrays by JSON type

Given `[object]` or `[array]`, a non-empty value of the matching JSON type
passes. An empty matching value requires `zero`. A value of the other container
type fails even when non-empty.

#### AC-ValidateTypes-8: Pass the PRD datetime examples

Given `[datetime, zero]`, each of these valid examples passes:

```text
2026-01-01
2026-01-01Z
2026-01-01+00:00
2026-01-01T00:00:00Z
2026-01-01T00:00:00.000+00:00
2026-01-01T00:00:00.000000Z
```

A time without a timezone, an impossible calendar value, a non-string, or a
fraction containing any number of digits other than three or six fails.

#### AC-ValidateTypes-9: Validate datetime offsets and zero instants

Given any valid `±HH:MM` offset, when the datetime is otherwise valid, then it
passes format validation. Given an accepted spelling representing the instant
`0001-01-01T00:00:00Z`, when `zero` is absent it fails and when `zero` is
present it passes.

#### AC-ValidateTypes-10: Validate canonical UUIDs

Given canonical lower-, upper-, or mixed-case hyphenated UUID strings of any
version, when the UUID is nonzero, then `[uuid]` passes. The all-zero UUID
requires `zero`. Malformed or non-string values fail.

#### AC-ValidateTypes-11: Collect independent selector failures

Given multiple selectors whose selected values fail, when `ValidateTypes`
runs, then it returns one failure per failed selector in lexicographic selector
order and does not stop after the first assertion failure.

#### AC-ValidateTypes-12: Forward jq failures

Given invalid jq syntax, a jq evaluation failure, invalid required response
JSON, command startup failure, or context cancellation, when selector execution
fails, then the method stops, returns a fatal error preserving the cause and
exact external exit code when present, and does not classify it as an assertion
mismatch.

#### AC-ValidateTypes-13: Use jq directly and safely

Given a selector containing spaces, quotes, or shell metacharacters, when it is
evaluated, then the complete selector is passed to jq as one argument after
`--`, no shell interprets it, and jq alone determines its syntax and result.

#### AC-ValidateTypes-14: Accept no type assertions

Given an empty or nil type map, when `ValidateTypes` runs, then it returns nil
without reading response JSON or invoking jq.

## State and concurrency

`Validator` has no mutable package or instance state. All selectors, decoders,
jq processes, buffers, and errors are local to one call.

Different step runners may call the same `Validator` concurrently. Their jq
processes, standard streams, and returned slices must remain isolated and
race-free.

### Acceptance criteria

#### AC-Concurrency-1: Validate steps independently

Given concurrent calls for different steps, when their validations complete,
then each result contains only its own selectors, values, diff, and errors; one
call's cancellation does not corrupt another call.
