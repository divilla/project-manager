# `internal/reporter` Reporter

## Status and authority

- Package: `internal/reporter`
- Shared models: `internal/domain`
- Error builder: `pkg/errs`
- Binding reference: `skeleton/internal/reporter/reporter.go`
- Status: skeleton-aligned specification

This specification must not extend the Reporter beyond the binding skeleton
and `agent/prds/app.md`.

## Purpose

`Reporter` is the sole owner of human-readable standard output. The CLI
normally constructs it with `os.Stdout`; tests may supply a buffer, a failing
writer, or another `io.Writer`.

The CLI and execution services call Reporter methods instead of retaining or
writing to a standard-output writer themselves. Fatal diagnostics written to
standard error remain owned by the CLI.

## Public contract

```go
var ReporterError = errors.New("reporter error")

type Reporter struct {
    output io.Writer
}

func NewReporter(output io.Writer) *Reporter

func (r *Reporter) WorkingDirectory(workDir string) error

func (r *Reporter) Success(
    ctx context.Context,
    directory *domain.Directory,
) error

func (r *Reporter) FailureTypes(
    ctx context.Context,
    step *domain.Step,
    failure error,
) error

func (r *Reporter) FailureExpected(
    ctx context.Context,
    step *domain.Step,
    failure error,
) error

func (r *Reporter) Debug(
    ctx context.Context,
    step *domain.Step,
) error
```

The names `FailureExpected` and `internal/reporter` are authoritative. Legacy
names such as `FailureDiff`, `internal.runtime.Reporter`, and
`internal.output.Reporter` are not part of the product API.

## Construction and injection

`NewReporter` retains the supplied writer and performs no output. The
application composition root supplies `os.Stdout`:

```go
reporter.NewReporter(os.Stdout)
```

Tests inject another writer through the same constructor. Components that
produce standard output receive the constructed `*Reporter`; they do not
receive `os.Stdout` or a separate `io.Writer`.

## `WorkingDirectory`

`WorkingDirectory` writes exactly:

```text
Working Directory: <workDir>

```

The method returns nil after a successful write. A nil Reporter, nil writer,
or write failure returns a built error matching `ReporterError` with internal
exit code `103`. A writer failure remains discoverable through
`errors.Is`/`errors.As`.

## Execution reporting entry points

`Success`, `FailureTypes`, `FailureExpected`, and `Debug` establish the
presentation boundary used by `execution.StepRunner`:

- `Success` receives a completed directory.
- `FailureTypes` receives one step and one nonfatal type-validation failure.
- `FailureExpected` receives one step and one nonfatal expected-response
  failure.
- `Debug` receives one runtime step selected for debug presentation.

Their exact text, colors, aggregation, terminal detection, formatting tools,
and scheduling rules are intentionally not specified by the current skeleton.
Adding any of those commitments requires changing the skeleton first.

## Error ownership

`ReporterError` is the Reporter package's static classification. Contextual
errors are built through `pkg/errs`; Reporter must not use `fmt.Errorf` or
another local wrapping mechanism.

Reporter methods return errors to their caller. They never print their own
errors, choose a process exit code independently of a built error, or write a
fatal diagnostic.

## Concurrency and state

The current skeleton retains only the injected writer. It does not establish a
public buffering, mutex, terminal-detection, or prior-output contract. Because
StepRunner may invoke reporting from same-stage directory goroutines, a future
implementation must make writer access safe before it emits execution output;
the precise mechanism is private.

## Non-responsibilities

Reporter does not:

- discover, decode, validate, resolve, prepare, schedule, or execute steps;
- invoke external tools;
- define validation semantics or error classifications owned by execution;
- choose suite or step control flow;
- write fatal diagnostics to standard error;
- define a JSON event stream, summary model, or CLI flag.

## Acceptance criteria

1. `NewReporter` retains an injected writer and writes nothing.
2. The CLI constructs Reporter with `os.Stdout`; tests can inject another
   writer.
3. `WorkingDirectory` produces the exact two-line boundary and preserves
   writer failures in an error matching `ReporterError` with code `103`.
4. StepRunner receives `*reporter.Reporter`, not `io.Writer`.
5. No production component outside `internal/reporter` writes standard output
   directly.
6. The four execution reporting methods compile with the exact skeleton names
   and parameter types.
7. No legacy Git diff, `jq` pretty-printing, `bat`, ANSI-color, debug-stop, or
   per-directory output behavior is treated as a current contract.
