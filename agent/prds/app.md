# APIHydra Product Requirements Document

## Authority and status

- Product: APIHydra
- CLI command: `apih`
- Status: skeleton-aligned draft
- Binding reference: `skeleton/`

This PRD describes only architecture, APIs, data, and behavior represented by
the binding skeleton. It must not introduce another package, service, command,
model, field, method signature, error owner, exit code, or execution path unless
the skeleton is explicitly changed first.

## Product

APIHydra is a Go CLI that discovers YAML definitions in a directory tree,
decodes and validates those definitions, resolves inherited request defaults,
prepares steps, executes HTTP requests through external command wrappers, and
validates responses.

The product operates on one `domain.Suite`. A suite has a selected working
directory and a root `domain.Directory`; directories retain their hierarchy,
files, decoded definitions, resolved values, and runtime steps throughout one
run.

## Package boundaries

The repository uses these skeleton packages:

| Package | Responsibility |
| --- | --- |
| `cmd/cli` | Resolve the working directory, compose services, print terminal errors, and return the selected exit code. |
| `internal/domain` | Own all shared suite, directory, file, definition, defaults, and step models. |
| `internal/definition` | Load, classify, decode, validate, and resolve definitions. |
| `internal/execution` | Own variables, validation, step preparation, stage scheduling, and step execution. |
| `pkg/errs` | Build every contextual error and attach exit-code metadata. |
| `pkg/runner` | Own every external command invocation. |

Services communicate through `internal/domain` values. Packages must not add
parallel model hierarchies or duplicate another package's responsibility.

## CLI contract

`cmd/cli` starts from `os.Getwd()`. When a first positional argument exists, it
joins that value to the current directory and requires the result to identify a
directory. Invalid paths match the CLI-owned `InvalidPathError`.

After resolving the working directory, the CLI writes:

```text
Working Directory: <path>

```

It creates `domain.Suite{WorkDir: workDir}` and invokes the definition pipeline
in this exact order:

1. `Loader.LoadDirectoryStructure`
2. `Loader.LoadDirectoryFiles`
3. `Loader.DecodeBaseDefinitions`
4. `Decoder.DecodeFiles`
5. `Decoder.ValidateDefaultsDefinitions`
6. `Decoder.ValidateStepsDefinitions`
7. `Resolver.ResolveDefaults`
8. `Resolver.ResolveSteps`

The current CLI skeleton ends after `ResolveSteps`. Execution services exist as
an API but are not yet composed into `cmd/cli`.

## Domain contract

### Documents

`domain.DocumentKind` has exactly these values:

```go
KindRoot     = "root"
KindDefaults = "defaults"
KindSteps    = "steps"
```

`BaseDefinition` contains `app`, `kind`, and a raw `spec`. A classified
defaults or root document decodes to `DefaultsDefinition`; a steps document
decodes to `StepsDefinition`. Definitions may carry `metadata.name` and
`metadata.labels`.

### Suite tree

```go
type Suite struct {
    WorkDir string
    Root    *Directory
}
```

Each `Directory` retains:

- `Stage`, `Path`, `Parent`, and `Children`;
- discovered `Files`, one `DefaultsFile`, and ordered `StepsFiles`;
- one `DefaultsDefinition` and ordered `StepsDefinitions`;
- `ResolvedDefaults`, `ResolvedSteps`, and `RuntimeSteps`.

Each `File` retains its `Stage`, `Path`, `Kind`, exact `Bytes`, and owning
`Directory` pointer.

### Defaults and steps

`Defaults` exposes only:

```text
baseUrl, basePath, headers, timeout, retries
```

`Step` exposes only:

```text
vars
request.method
request.baseUrl
request.basePath
request.path
request.headers
request.timeout
request.retries
request.query
request.body
response.status
response.body
response.expected
response.types
response.capture
debug
```

Variable values, request bodies, response expectations, and capture selectors
use `domain.YAMLString`. A step retains its source `Definition` pointer and
zero-based `Index`. `DirectoryStage`, `DirectoryPath`, and `FilePath` derive
source information through that provenance chain.

## Definition pipeline

### Loader

`NewLoader() *Loader` creates a loader.

- `LoadDirectoryStructure(ctx, suite)` builds the directory tree rooted at
  `suite.WorkDir`. The root directory path is `/`; descendant paths are
  relative to `Suite.WorkDir`.
- `LoadDirectoryFiles(ctx, suite)` populates each `Directory.Files` with `.yaml`
  and `.yml` files.
- `DecodeBaseDefinitions(ctx, suite)` decodes base headers, assigns `File.Kind`,
  and classifies defaults/root and steps files.

Discovered and classified file slices must be deterministic. `StepsFiles` are
ordered alphabetically by cleaned file path.

### Decoder

`NewDecoder() *Decoder` creates a decoder.

- `DecodeFiles(ctx, suite)` decodes `DefaultsFile` into
  `DefaultsDefinition` and `StepsFiles` into `StepsDefinitions` without
  mutating unrelated fields.
- `ValidateDefaultsDefinitions(ctx, suite)` validates every decoded defaults or
  root definition.
- `ValidateStepsDefinitions(ctx, suite)` validates every decoded steps
  definition.

### Resolver

`NewResolver() *Resolver` creates a resolver.

- `ResolveDefaults(ctx, suite)` traverses the tree and populates each
  `Directory.ResolvedDefaults` by merging its local defaults with inherited
  parent defaults.
- `ResolveSteps(ctx, suite)` populates `Directory.ResolvedSteps` by resolving
  local steps against the directory defaults.
- `ValidateStepsDefinitions(ctx, suite)` exposes the resolver-owned validation
  boundary present in the skeleton.

Resolution preserves alphabetical file order and step declaration order in the
two-dimensional `ResolvedSteps` matrix.

## Execution contract

### Key-value store

`NewKeyValueStore()` returns an empty, concurrency-safe, string-to-string
store. `Set` is write-once: a duplicate key returns execution-owned
`KeyExistError` and never overwrites the original value. `Get` returns the
stored value or execution-owned `NotFoundError`. The store uses an `RWMutex` so
same-stage directory goroutines may access it safely.

### Variable processor

`NewVariableProcessor() *VariableProcessor` creates a processor. Its exact
operations are:

```go
Load(ctx, step)                  (int, error)
ParseRequestBody(ctx, step)      (int, error)
ParseResponseExpected(ctx, step) (int, error)
Capture(ctx, step)               (int, error)
```

Success returns exit code `0` and nil. Failures return the applicable APIHydra
code or exact external-tool code with a non-nil error.

### Validator

`Validator` exposes:

```go
ValidateTypes(ctx, step)    []error
ValidateExpected(ctx, step) error
```

`ValidateTypes` may return more than one failed validation. A non-nil
`ValidateExpected` result represents an expected-response failure unless it is
a fatal built error. `ValidationError` is the execution package's static
classification for one or more nonfatal failures reported by either method.

### StepRunner

```go
func NewStepRunner(
    variableProcessor *VariableProcessor,
    validator *Validator,
    output io.Writer,
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

The constructor retains the supplied processor, validator, and writer.

`Prepare` traverses from `Suite.Root`. For each resolved step it runs
`VariableProcessor.Load` and `VariableProcessor.ParseRequestBody`.

`Execute` validates the directory tree before scheduling:

- the suite and root must be non-nil;
- root stage must be `0`;
- children must be non-nil and unique;
- each child's `Parent` must identify its containing directory;
- each child stage must equal its parent stage plus one;
- cycles and repeated pointers are invalid.

Invalid input matches execution-owned `ErrInvalidDirectoryTree` and returns
exit code `102` without panicking.

Execution order is exact:

1. Start at stage `0` and process higher stage numbers in ascending order.
2. Start one goroutine for every directory whose `Directory.Stage` equals the
   active stage.
3. Run those directory goroutines in parallel.
4. Wait for every active-stage directory goroutine before starting the next
   stage.
5. Within one directory, process files one at a time in the alphabetical order
   represented by `StepsFiles` and the outer step matrix.
6. Execute every step in one file sequentially in declaration order before
   beginning the next file.
7. Never create file or step goroutines.

For each step, execution uses only functions exposed by `pkg/runner` for
external work. The skeleton phase order is `runner.Curl`,
`ParseResponseExpected`, `ValidateTypes`, `ValidateExpected`, then `Capture`.

Nonfatal failures from `ValidateTypes` or `ValidateExpected` are written through
the injected output writer. Execution continues through all remaining steps,
files, directories, and stages. When at least one such failure occurred,
`Execute` returns exit code `101` and an error matching `ValidationError` after
the complete suite finishes.

The first fatal directory error cancels the shared execution context. Every
goroutine already started for that stage is joined, no later stage starts, and
the first error and its exit code take precedence over sibling cancellation
errors.

## External command boundary

Every external tool invocation belongs to `pkg/runner`.

The skeleton currently exposes:

```go
runner.Curl(method, url, headers, timeout, retries, query, body) (string, int, error)
runner.JQFilter(selector, input)                              (string, int, error)
```

`pkg/runner` owns the static `CommandError`, `CurlError`, and
`JQSelectorError` classifications. Any future external command must receive a
dedicated function in `pkg/runner` before another package may use it.

No production package outside `pkg/runner` may import `os/exec`, construct a
command, start a process, or invoke an external executable directly. Other
packages call runner functions and preserve their returned output, exact exit
code, and error.

## Error architecture

### Static errors

Every stable static error is declared with `errors.New` in the package where
the failure originates. Examples include:

- `cmd/cli.InvalidPathError`;
- `execution.NotFoundError`, `execution.KeyExistError`,
  `execution.ErrInvalidDirectoryTree`, `execution.ExecutionCanceledError`, and
  `execution.ValidationError`;
- `runner.CommandError`, `runner.CurlError`, and `runner.JQSelectorError`.

Static errors carry classification only. They do not build contextual messages
or own exit-code wrappers.

### Built errors

Every contextual error is built in `pkg/errs`. Production packages outside
`pkg/errs` must not call `fmt.Errorf` or otherwise compose, wrap, or decorate an
error. They pass their package-owned static error, original cause, exit code,
and contextual values to an `errs` builder.

`pkg/errs` owns:

```go
Build(code, errStatic, errOriginal, details...) error
WithExitCode(code, err) error
Code(err, fallback) int
DefaultsDefinitionError(...)
StepDefinitionError(...)
StepExecutionError(...)
```

A built error preserves both the originating static error and original cause
for `errors.Is`/`errors.As`, exposes its code through `ExitCoder`, and formats
context in one place. Definition errors default to configuration exit code
`102`. Step execution errors preserve a coded original tool error and otherwise
default to internal exit code `103`.

## Exit codes

APIHydra reserves:

| Code | Meaning |
| --- | --- |
| `0` | Success. |
| `101` | Execution completed with at least one type or expected validation failure. |
| `102` | Invocation, definition, tree, variable, or other configuration failure. |
| `103` | Internal failure or missing external dependency. |

External command failures return the exact non-zero code supplied by the
runner function. Consequently, low non-zero codes are never created by
APIHydra itself. An external tool may coincidentally return `101`, `102`, or
`103`, so a reserved number alone does not prove the error's origin; the built
error classification remains authoritative.

No error may be returned with exit code `0`.

## Output boundary

The CLI owns fatal terminal diagnostics. `StepRunner` owns only the injected
`io.Writer` used for validation output. The skeleton defines no Reporter,
machine-readable event stream, ANSI-color contract, or final summary model.

## Not specified by the skeleton

The following are not product commitments until explicitly added to the
skeleton:

- name or label filtering flags;
- external-tool preflight APIs;
- Git-based comparison functions;
- direct HTTP-status capture or validation behavior;
- a Reporter service or JSON event output;
- exact variable interpolation syntax;
- detailed response-type tokens and modifiers;
- URL normalization beyond values passed to `runner.Curl`;
- additional document fields, commands, services, or exit codes.

Fields already present in the domain model, including `Response.Status` and
`Debug`, remain available data but gain no behavior beyond an explicit skeleton
service contract.

## Acceptance criteria

1. Production code compiles against the exact skeleton packages, models, and
   method signatures without adapters that create a competing API.
2. The CLI runs the definition pipeline in the documented order and returns
   only `0`, a reserved APIHydra code, or an exactly forwarded runner code.
3. Directory stages run in ascending order with a complete barrier; directories
   in one stage run in parallel.
4. Files in one directory run alphabetically and never overlap; every file's
   steps run sequentially in declaration order.
5. Invalid directory graphs return a built error matching
   `ErrInvalidDirectoryTree` with code `102` and never panic.
6. Any type or expected validation failure allows remaining work to finish and
   produces final code `101`.
7. The first fatal same-stage error retains its cause and code, cancels sibling
   work, joins the stage, and prevents later stages.
8. Static errors are declared only by their originating package; all contextual
   errors are built only by `pkg/errs`.
9. Only `pkg/runner` may invoke external commands; no direct command execution
   exists elsewhere.
10. `go test ./...`, `go test -race ./...`, and `git diff --check` pass.
