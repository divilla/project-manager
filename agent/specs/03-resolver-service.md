# internal.definition.Resolver Service

## Status

- Service: `internal/definition.Resolver`
- Package: `internal/definition`
- Package name: `definition`
- Shared models: `internal/models`
- Error construction: `pkg/errs`
- Status: implementation specification

## Base specification

### Purpose

`Resolver` is the processing stage immediately after `Decoder`. It converts the
decoded declarations in every directory into effective resolved values for the
later preparation stage.

It performs two explicit operations:

```text
*models.Directory
  -> ResolveDefaults
       -> Directory.ResolvedDefaults
  -> ResolveSteps
       -> Directory.ResolvedSteps
```

`ResolveDefaults` builds inherited resolved defaults for every directory.
`ResolveSteps` then resolves every locally declared step against its
directory's resolved defaults. Defaults resolution processes parents before
children so a child can use its parent's already resolved defaults.

The service deliberately uses the shared declarative value types for its
outputs:

```go
Directory.ResolvedDefaults models.Defaults
Directory.ResolvedSteps    [][]models.Step
```

Resolved values must be independent copies. Apart from assigning the
non-YAML `Step.Definition` and `Step.Index` provenance fields, resolving a suite
must not mutate or alias values owned by `DefaultsDefinition` or
`StepsDefinition`.

The lifecycle boundary is:

```text
decoded Defaults and Steps
  -> Resolver
       -> ResolvedDefaults and ResolvedSteps
  -> preparation with variable values
       -> RuntimeDefaults and RuntimeSteps
```

`ResolvedSteps` still contain their declarative `body` and `expected` values.
The preparation phase applies available variable values and parses those fields
when it creates runtime values. Resolver must not perform that preparation.

### Name

`Resolver` is preferred over `Runtimer` because the stage resolves inheritance
and overlays into resolved values; it does not run requests. The shorter name
also leaves `Runner` available for the later execution service.

### Responsibilities

`Resolver` owns:

- Traversing a fully decoded and validated directory hierarchy.
- Resolving each directory's defaults from its parent and its optional local
  defaults definition.
- Merging headers case-insensitively and emitting canonical HTTP header names.
- Preserving local steps grouped by their decoded steps definition and in their
  decoded order.
- Resolving every local step against its directory's resolved defaults.
- Attaching each declared and resolved step to its exact source
  `StepsDefinition` with a zero-based index within that definition.
- Applying built-in request method, timeout, and retry defaults.
- Ensuring each resolved step has the values required by later request
  processing.
- Populating `Directory.ResolvedDefaults` and `Directory.ResolvedSteps` for every
  traversed directory.
- Reporting source-file- and YAML-path-aware resolution errors.
- Construction of YAML-source errors through `errs.Definition`.

### Non-responsibilities

`Resolver` does not:

- Discover, read, classify, or decode files.
- Repeat static declaration validation owned by `Decoder`.
- Filter steps definitions by metadata.
- Compose a final URL or execute URL-specific validation.
- Substitute variables or mutate the run-wide variable store.
- Parse prepared request bodies or expected response values.
- Produce `RuntimeDefaults` or `RuntimeSteps`; those belong to preparation.
- Serialize step variables as JSON.
- Validate request bodies or expected response values with `jq`.
- Check external-tool availability.
- Invoke `curl`, `jq`, Git, or any other process.
- Process response captures, comparisons, or type assertions.
- Regroup directories into stages or execute a stage.
- Produce terminal or machine-readable output.

On an error path, `errs.Definition` may read the source YAML file solely to
resolve the formatted line number. Resolver itself performs no other filesystem
access.

### Public contract

```go
type Resolver struct{}

func NewResolver() *Resolver

func (r *Resolver) ResolveDefaults(
    ctx context.Context,
    root *models.Directory,
) error

func (r *Resolver) ResolveSteps(
    ctx context.Context,
    root *models.Directory,
) error
```

### Required call sequence

The intended workflow is:

```text
Loader.LoadDirectoryStructure
  -> Loader.LoadDirectoryFiles
  -> Loader.DecodeBaseDefinitions
  -> Decoder.DecodeFiles
  -> Decoder.ValidateDefaultsDefinitions
  -> Decoder.ValidateStepsDefinitions
  -> Resolver.ResolveDefaults
  -> Resolver.ResolveSteps
```

Both methods require the complete Decoder sequence to have succeeded.
`ResolveSteps` additionally requires `ResolveDefaults` to have
succeeded for the same decoded tree. The methods remain separate operations:
neither invokes the other or any Decoder method implicitly.

### Service state

`Resolver` has no retained configuration or mutable run state. A context,
directory tree, temporary presence information, candidate resolved values, and
errors are scoped to one method call.

### Cancellation

Each Resolver method must check `ctx.Err()` before work begins and at useful
boundaries throughout its directory, definition, and step traversal. A
cancelled call returns the context error and commits none of that method's
candidate values.

### Mutation policy

- `ResolveDefaults` mutates only `Directory.ResolvedDefaults`.
- `ResolveSteps` reads `Directory.ResolvedDefaults` and mutates
  `Directory.ResolvedSteps` plus only the `Definition` and `Index` provenance
  fields of entries in `StepsDefinition.Spec.Steps`.
- All other directory, file, definition, and declared value fields are
  read-only.
- Each method is independently transactional across the complete supplied tree.
  An error or cancellation leaves the fields owned by that method at their
  pre-call values.
- A successful repeated call replaces the field owned by that method throughout
  the tree; it must not append to or merge with results from a previous call.
- Every output map and nested mutable value is copied deeply enough that later
  resolved-output mutation cannot modify decoded declarations, parent resolved
  values, or a sibling resolved step.
- `Step.Definition` is the sole intentional alias into decoded state. It is a
  read-only provenance pointer to the exact containing `StepsDefinition`.

### Error construction

Every Resolver error attributable to a defaults or steps definition must be
built with:

```go
errs.Definition(definition.File.Path, yamlPath, staticMessage, cause)
```

The Resolver supplies the exact source file, the most specific YAML path, and a
stable non-empty static message for the failed resolution rule. Resolution
rules normally pass a nil cause; an underlying source-inspection or conversion
error is attached when one exists.

Defaults paths use forms such as `$.spec.baseUrl` and `$.spec.headers`. Step
paths use the source definition and the zero-based YAML position derived from
the resolved step's zero-based index:

```text
$.spec.steps[<Step.Index>].request.baseUrl
```

Errors without an attributable YAML definition do not use `errs.Definition`.
These include nil roots, broken directory relationships without a usable source
definition, and context cancellation.

#### AC-ResolverErrors-1: Build YAML-source errors with `errs.Definition`

Given a defaults- or step-resolution failure attributable to a YAML member,
when a Resolver method returns the error, then its filename, one-based line,
YAML path, static message, and optional attached cause follow the
`errs.Definition` contract.

#### AC-ResolverErrors-2: Use step indexes as YAML indexes

Given a resolved step with zero-based `Step.Index`, when Resolver constructs its
definition error, then the YAML path uses `Step.Index` directly as the
zero-based `spec.steps` sequence index.

#### AC-ResolverErrors-3: Keep non-definition errors outside definition format

Given cancellation or an input-tree failure without an attributable YAML
member, when Resolver returns the error, then it does not fabricate arguments
for `errs.Definition`.

## Types specification

### `models.Directory`

The Resolver-visible contract is:

```go
type Directory struct {
    Stage              int
    Path               string
    Parent             *Directory
    Children           []*Directory
    Files              []*File
    DefaultsFile       *File
    StepsFiles         []*File
    DefaultsDefinition *DefaultsDefinition
    StepsDefinitions   []*StepsDefinition
    ResolvedDefaults    Defaults
    ResolvedSteps       [][]Step
}
```

`ResolveDefaults` reads `Parent`, `Children`, and
`DefaultsDefinition`, and writes only `ResolvedDefaults`.
`ResolveSteps` reads `Children`, `StepsDefinitions`, and
`ResolvedDefaults`, and writes `ResolvedSteps` plus each source step's
`Definition` and `Index` provenance fields.

Traversal follows `Children`; it must not recursively follow `Parent`.
Resolution must verify that every child's `Parent` points to its containing
directory before using inherited values.

### `models.Defaults`

`Directory.ResolvedDefaults` uses the same type as a decoded defaults spec:

```go
type Defaults struct {
    BaseURL  string
    BasePath string
    Headers  map[string]string
    Timeout  int
    Retries  int
}
```

Despite the shared type, a resolved value is a resolved copy and never the
decoded declaration itself.

The zero values have these resolved meanings:

- An empty `BaseURL` means no ancestor or local defaults definition supplied
  one. It is allowed on a directory that has no local steps.
- An empty `BasePath` means no base path is configured.
- A nil or empty `Headers` map means no default headers are configured.
- Zero `Timeout` and `Retries` mean no defaults definition supplied those
  values. Built-in values are applied while resolving each resolved step, not
  stored as directory defaults.

This preserves the distinction between inherited suite configuration and
request-level built-ins.

### `models.Step`

Each outer `Directory.ResolvedSteps` entry corresponds to one
`Directory.StepsDefinitions` entry. Each inner entry is a resolved copy of one
decoded step with inherited and built-in request values populated. Fields
unrelated to defaults resolution, including `Vars` and all `Response` fields,
are preserved unchanged in value and copied without mutable aliases.

In particular, `Request.Body` and `Response.Expected` remain `YAMLString`
values and may still contain variable references. Their variable-aware parsing
belongs to preparation, after this service returns.

`ResolvedSteps` preserves this deterministic two-dimensional structure:

1. Outer slice indexes correspond exactly to `Directory.StepsDefinitions`
   indexes.
2. Inner slice indexes correspond exactly to
   `StepsDefinition.Spec.Steps` declaration indexes.

A directory with no steps definitions receives an empty outer `ResolvedSteps`
slice. A steps definition with no declared steps contributes an empty inner
slice so later definitions retain their positional correspondence.

For a step at zero-based position `i` in `StepsDefinition.Spec.Steps`, Resolver
must set the source step and its resolved copy to:

```go
definition.Spec.Steps[i].Definition = definition
definition.Spec.Steps[i].Index = i

resolved.Definition = definition.Spec.Steps[i].Definition
resolved.Index = definition.Spec.Steps[i].Index
```

`Definition` points to the exact containing `*models.StepsDefinition`, and
`Index` is zero-based for user-facing identity and errors. The pointer is an
intentional read-only provenance link; Resolver must not clone the definition.
No other field on the decoded source step may change.

### Presence-sensitive declarations

The shared models intentionally use value fields, so Go zero values alone do
not distinguish omission from an explicit declaration. Resolution must retain
the Decoder's required semantics, especially:

- An explicit descendant `spec.retries: 0` overrides a positive inherited
  retry count, while an omitted `spec.retries` leaves it inherited.
- An explicit step `request.retries: 0` overrides a positive inherited retry
  count.
- An omitted `request.retries` inherits a configured retry count or receives
  the built-in value `3`.
- Method selection distinguishes an omitted body from a present body when
  deciding between `GET` and `POST`.

`Resolver` may derive this limited field-presence information from each
definition's `File.Bytes`. It must not expose YAML syntax types through shared
models, replace the Decoder's strict schema validation, or mutate file bytes.

## `NewResolver`

```go
func NewResolver() *Resolver
```

### Behavior

`NewResolver` returns a non-nil, stateless Resolver. It performs no traversal,
resolution, filesystem access, decoding, or validation.

### Acceptance criteria

#### AC-NewResolver-1: Create a stateless resolver

When `NewResolver` is called, then it returns a non-nil Resolver with no
retained suite or per-call state.

## `ResolveDefaults`

```go
func (r *Resolver) ResolveDefaults(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: a hierarchy successfully decoded and validated by `Decoder`.

The root must have no parent and must contain a root `DefaultsDefinition`.
Every descendant may contain no defaults definition or one defaults definition.

### Output

- `nil` after `ResolvedDefaults` has been populated for every directory.
- An error when the decoded tree is inconsistent, required presence
  information cannot be recovered, or the context is cancelled.

The method returns the first error in deterministic parent-before-child,
`Children` order.

### Directory traversal

Defaults resolution uses depth-first, parent-before-child traversal in the
exact order of each `Children` slice. A directory's complete candidate
`ResolvedDefaults` is built before its children are visited.

The `Stage` field does not control inheritance and is not changed. A child
inherits from its actual parent even if the supplied stage values are absent or
incorrect; stage validation and execution belong elsewhere.

### Resolution

For the supplied root:

1. Begin with an empty `models.Defaults` value.
2. Overlay the root `DefaultsDefinition.Spec`.
3. Store the result as the root candidate `ResolvedDefaults`.

For every descendant:

1. Deep-copy its parent's candidate `ResolvedDefaults`.
2. If the directory has a `DefaultsDefinition`, overlay every explicitly
   declared local field.
3. Store the result as the child candidate `ResolvedDefaults`.

An explicitly declared local scalar replaces the inherited scalar. An omitted
scalar leaves the inherited value unchanged. A descendant without a defaults
definition receives an independent copy equal in value to its parent's resolved
defaults.

### Header resolution

Default headers are resolved as follows:

1. Header names are compared case-insensitively.
2. A child header replaces an inherited header with the same case-insensitive
   name.
3. Unrelated inherited headers remain present.
4. Every output key is canonicalized with Go's standard HTTP header casing.
5. The output contains at most one entry for a case-insensitive header name.

Resolution must not mutate any input header map. If one declaration itself
contains two differently cased forms of the same header name, the later entry
cannot be determined reliably from a Go map; Decoder must reject that ambiguity
before Resolver is called.

### Source-aware errors

An inherited-default error identifies the local `DefaultsDefinition.File.Path`
and relevant YAML path when available. Errors must not depend on pointer
addresses or unordered map iteration.

### Acceptance criteria

#### AC-ResolveDefaults-1: Resolve root defaults

Given a validated root definition, when `ResolveDefaults` is called,
then the root receives an independent `ResolvedDefaults` copy containing every
explicitly declared root default.

#### AC-ResolveDefaults-2: Inherit through directories without defaults

Given descendants with no local defaults definition, when the method succeeds,
then each receives resolved defaults equal in value to its parent's values
without sharing mutable maps.

#### AC-ResolveDefaults-3: Overlay descendant defaults

Given parent defaults and a child definition that declares only some fields,
when the method succeeds, then the child receives its declared values and
inherits every omitted value from its parent.

#### AC-ResolveDefaults-4: Merge and canonicalize headers

Given a parent `content-type` header and child `Content-Type` and `x-trace`
headers, when the child defaults are resolved, then its headers contain the
child `Content-Type`, canonical `X-Trace`, and no duplicate content-type key.

#### AC-ResolveDefaults-5: Populate every directory

Given directories with and without local defaults definitions, when the method
succeeds, then every traversed directory has candidate `ResolvedDefaults`.

#### AC-ResolveDefaults-6: Avoid mutable aliases

Given successfully resolved defaults, when a child's resolved header is
changed, then no decoded definition, parent resolved defaults, or sibling
directory changes.

#### AC-ResolveDefaults-7: Replace prior defaults

Given previously populated resolved defaults, when the method succeeds again,
then every directory reflects only the current decoded defaults definitions.

#### AC-ResolveDefaults-8: Reject inconsistent input transactionally

Given a missing root definition, broken parent link, or defaults definition
whose source file does not belong to its directory, when the method is called,
then it returns an error and leaves all `ResolvedDefaults` unchanged.

#### AC-ResolveDefaults-9: Honor cancellation transactionally

Given cancellation before or during traversal, when the method observes the
cancelled context, then it returns the context error, performs no further work,
and leaves all `ResolvedDefaults` unchanged.

## `ResolveSteps`

```go
func (r *Resolver) ResolveSteps(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: the same decoded hierarchy previously passed successfully to
  `ResolveDefaults`.

Every directory's `ResolvedDefaults` must represent the current decoded defaults
definitions. Every steps definition must remain connected to its containing
directory through its source file. Because `models.Defaults` has no resolved
state marker, the caller owns the method ordering; zero-valued resolved defaults
do not prove that defaults resolution was skipped.

### Output

- `nil` after `ResolvedSteps` has been populated for every directory.
- An error when the decoded tree is inconsistent, a required resolved value
  cannot be resolved, presence information cannot be recovered, or the context
  is cancelled.

An error or cancellation leaves every `ResolvedDefaults` untouched, every
`ResolvedSteps` at its pre-call value, and all source-step `Definition` and
`Index` fields at their pre-call values.

### Directory traversal

Step resolution traverses directories depth-first in the exact order of each
`Children` slice. Within a directory it resolves definitions and steps in their
stored order. Parent-first traversal is deterministic but is not an inheritance
mechanism here; the method reads each directory's already resolved defaults.

The `Stage` field does not control traversal and is not changed.

### Resolution

For each directory, Resolver allocates one candidate outer slice entry per
`StepsDefinition`. For each declared step at zero-based position `i` within that
definition, Resolver assigns the source step's `Definition` to the exact
containing definition and `Index` to `i`. It then deep-copies the annotated
declaration and resolves its request fields as follows:

- An explicitly declared step value wins.
- Otherwise `baseUrl`, `basePath`, `headers`, `timeout`, and `retries` are taken
  from the directory's resolved defaults when configured.
- Step headers replace resolved-default headers with the same case-insensitive
  name; unrelated default headers remain, and all output names use Go's
  canonical HTTP header casing.
- An omitted method becomes `POST` when `request.body` is present and `GET`
  otherwise.
- An omitted timeout becomes `10` when neither the step nor directory defaults
  supply one.
- An omitted retry count becomes `3` when neither the step nor directory
  defaults supply one.
- `path`, `query`, `body`, `vars`, and response declarations have no defaults
  source and preserve their declared values.

After resolution, every resolved step must have:

- A `Definition` pointer to its exact containing `StepsDefinition`.
- An `Index` equal to its zero-based position within
  `StepsDefinition.Spec.Steps`.
- A non-empty request base URL.
- A non-empty request method.
- A non-empty request path.
- A positive timeout.
- A non-negative retry count.

Static invalid declarations should already have been rejected by Decoder. The
method reports only inconsistencies in its input contract and values that
remain missing or invalid after inheritance and built-ins.

Resolver does not join `baseUrl`, `basePath`, `path`, and `query`. Those
components remain separate on the resolved `Step` for the preparation stage.

### Source-aware errors

A resolved-step error identifies the originating `StepsDefinition.File.Path`,
the step's zero-based `Index` within that definition, and the relevant YAML path.
Errors must not depend on pointer addresses or unordered map iteration.

### Acceptance criteria

#### AC-ResolveSteps-1: Populate every directory

Given directories with and without steps definitions, when the method succeeds,
then each `ResolvedSteps` has one outer entry per local steps definition, each
inner entry contains that definition's resolved steps, and directories without
steps definitions have an empty outer slice.

#### AC-ResolveSteps-2: Preserve step order

Given multiple steps definitions in one directory, when they are resolved,
then the outer `ResolvedSteps` indexes preserve definition order and each inner
slice preserves its definition's declaration order, including an empty inner
slice for an empty definition.

#### AC-ResolveSteps-3: Attach definition and zero-based index

Given a step at zero-based slice position `i` in a `StepsDefinition`, when it is
resolved, then both the source step and resolved copy have `Definition` pointing
to that exact definition and `Index` equal to `i`. Every other decoded
source-step field remains unchanged.

#### AC-ResolveSteps-4: Overlay step request values

Given directory resolved defaults and a step that explicitly overrides some
request fields, when the method succeeds, then the resolved step uses every
explicit step value and inherits each omitted defaultable field.

#### AC-ResolveSteps-5: Merge step headers

Given resolved default headers and step headers with case-insensitive overlap,
when the step is resolved, then the step value wins, unrelated defaults remain,
and every output header name is canonical.

#### AC-ResolveSteps-6: Apply method defaults

Given an omitted request method, when the step has a body then the resolved
method is `POST`, and when it has no body the resolved method is `GET`. Given an
explicit method, its spelling is preserved exactly.

#### AC-ResolveSteps-7: Apply timeout and retry defaults

Given no step or directory timeout and retry declarations, when the step is
resolved, then its resolved timeout is `10` and retries is `3`. Inherited values
override those built-ins, and explicit step values override inherited values.

#### AC-ResolveSteps-8: Preserve an explicit zero retry count

Given an inherited positive retry count and a step explicitly declaring
`request.retries: 0`, when the step is resolved, then its resolved retry count is
zero rather than the inherited or built-in value.

#### AC-ResolveSteps-9: Require a resolved base URL

Given a declared step for which neither the step nor its directory defaults
supply `baseUrl`, when the method is called, then it returns a source-aware
error, leaves all `ResolvedSteps` and source-step provenance fields unchanged,
and preserves `ResolvedDefaults`.

#### AC-ResolveSteps-10: Avoid mutable aliases

Given successfully resolved steps, when a resolved header, variable, capture,
type declaration, or other nested mutable value is changed, then no decoded
definition, resolved defaults, or sibling resolved step changes.

#### AC-ResolveSteps-11: Replace prior steps

Given previously populated resolved steps, when the method succeeds again, then
every directory reflects only the current steps definitions and contains no
stale or duplicated step.

#### AC-ResolveSteps-12: Reject inconsistent input transactionally

Given a broken child link or steps definition whose source file does not belong
to its directory, when the method is called, then it returns an error, leaves
all `ResolvedSteps` and source-step provenance fields unchanged, and preserves
`ResolvedDefaults`.

#### AC-ResolveSteps-13: Honor cancellation transactionally

Given cancellation before or during traversal, when the method observes the
cancelled context, then it returns the context error, performs no further work,
leaves all `ResolvedSteps` and source-step provenance fields unchanged, and
preserves `ResolvedDefaults`.

#### AC-Resolver-1: Remain filesystem- and process-free

Given valid in-memory inputs, when either Resolver method succeeds, then it
performs no filesystem read and invokes no external process.
