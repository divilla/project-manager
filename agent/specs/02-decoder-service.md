# internal.definition.Decoder Service

## Status

- Service: `internal/definition.Decoder`
- Package: `internal/definition`
- Package name: `definition`
- Shared models: `internal/models`
- Error construction: `pkg/errs`
- Status: implementation specification

## Base specification

### Purpose

`Decoder` converts the files classified by `Loader` into complete,
kind-specific APIHydra definitions and validates configuration rules that can be
decided from the declared YAML alone.

It performs three explicit operations:

1. Strictly decode every classified root, defaults, and steps file.
2. Validate all decoded defaults definitions.
3. Validate all decoded steps definitions.

The service enriches the directory-oriented model produced by `Loader`:

```text
*models.Directory
  -> DecodeFiles
       -> Directory.DefaultsDefinition
       -> Directory.StepsDefinitions
       -> DefaultsDefinition.File
       -> StepsDefinition.File
  -> ValidateDefaultsDefinitions
       -> read-only defaults validation
  -> ValidateStepsDefinitions
       -> read-only steps validation
```

The root document and descendant defaults documents share the same decoded
model. A root file decodes to a `DefaultsDefinition` whose kind is `KindRoot`;
a descendant defaults file decodes to a `DefaultsDefinition` whose kind is
`KindDefaults`.

### Responsibilities

`Decoder` owns:

- Traversal of an already loaded and classified directory hierarchy for the
  purpose of full YAML decoding and static validation.
- Strict, kind-specific decoding of recognized APIHydra files.
- Rejection of multiple YAML documents in one recognized file.
- Rejection of duplicate YAML mapping keys at any nesting depth.
- Rejection of unknown YAML fields at any schema-defined level.
- Typed decoding of optional metadata, defaults, steps, requests, and response
  declarations.
- Connecting every decoded definition to its exact source `File`.
- Populating decoded definitions on their containing `Directory`.
- Validation of metadata rules decidable from the complete decoded suite.
- Validation of defaults values that are invalid whenever explicitly declared.
- Validation of steps values that are invalid whenever explicitly declared.
- Validation of variable-key syntax and JSON compatibility for literal step
  variables.
- Validation of response status declarations.
- Validation of response type-declaration structure.
- File- and YAML-path-aware configuration errors.
- Construction of YAML-source errors through `errs.Definition`.

### Non-responsibilities

`Decoder` does not:

- Discover directories or files.
- Read files from the filesystem for decoding; it consumes `File.Bytes` loaded
  by `Loader`. On an error path, `errs.Definition` may read the source file only
  to resolve the formatted line number.
- Recognize unrelated YAML or classify APIHydra document kinds.
- Revalidate root/defaults placement or file cardinality already owned by
  `Loader`.
- Resolve parent defaults or defaults inheritance.
- Merge or canonicalize header names.
- Filter step definitions by metadata.
- Populate `Directory.ResolvedDefaults` or `Directory.ResolvedSteps`; those
  belong to the next processing stage.
- Supply inherited or built-in request defaults.
- Compose or validate final request URLs.
- Resolve variables or enforce the run-wide write-once variable store.
- Validate request bodies or expected response values as JSON.
- Execute or statically validate `jq` expressions.
- Check for external tools or invoke `curl`, `jq`, or Git.
- Process captures, compare responses, or evaluate response type assertions.
- Plan or execute stages.
- Produce terminal or machine-readable output.

### Public contract

```go
type Decoder struct{}

func NewDecoder() *Decoder

func (d *Decoder) DecodeFiles(
    ctx context.Context,
    root *models.Directory,
) error

func (d *Decoder) ValidateDefaultsDefinitions(
    ctx context.Context,
    root *models.Directory,
) error

func (d *Decoder) ValidateStepsDefinitions(
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
```

`DecodeFiles` requires successful Loader classification. Both validation
methods require successful full decoding. The validation methods are otherwise
independent and may be called in either order.

The methods remain separate operations. `DecodeFiles` must not implicitly call
either public validation method, and neither validation method must invoke the
other.

### Service state

`Decoder` has no retained configuration or mutable run state.

It must not retain:

- A context.
- A directory tree.
- A file or file bytes.
- A decoded definition.
- YAML syntax nodes.
- Per-call validation state or errors.

Contexts, temporary YAML representations, decoded candidates, name indexes,
and traversal state are scoped to individual method calls.

### Cancellation

Every Decoder method accepts `context.Context` as its first argument. Each
method must check `ctx.Err()` before beginning work and at useful boundaries
throughout directory, file, definition, and step traversal.

Cancellation must prevent additional decoding or validation work after the
next useful boundary. A cancelled call returns the context error.

### Mutation policy

- `DecodeFiles` mutates only `Directory.DefaultsDefinition` and
  `Directory.StepsDefinitions`.
- The `File` field of each newly decoded definition is set before the definition
  is committed to its directory.
- `ValidateDefaultsDefinitions` is read-only.
- `ValidateStepsDefinitions` is read-only.
- No Decoder method mutates Loader-owned directory fields, file fields, or file
  bytes.
- `DecodeFiles` is transactional across the complete supplied tree. An error
  leaves all pre-call definition fields unchanged.
- Validation failure or cancellation never modifies the supplied tree or any
  decoded definition.

### Error construction

Every Decoder error attributable to a specific YAML definition must be built
with:

```go
errs.Definition(sourcePath, yamlPath, staticMessage, cause)
```

During `DecodeFiles`, before a definition exists, the Decoder uses the
classified source `File.Path` directly. It supplies the most specific YAML path
available, a stable non-empty static message for the failed Decoder rule, and
the underlying YAML or typed-decoding error when one exists. Pure validation
failures pass a nil cause.

Document-wide errors use `$`. Schema fields use exact paths such as
`$.metadata.name`, `$.spec.timeout`, or
`$.spec.steps[0].request.path`. YAML sequence positions in these paths remain
zero-based.

For a duplicate metadata name, the later definition in deterministic traversal
order is the primary error location at `$.metadata.name`; the static message
identifies the earlier definition's file. Duplicate-key, unknown-field, and
typed-decoding causes must remain attached through `errs.Definition`.

Errors without an attributable YAML location do not use `errs.Definition`.
These include nil roots, broken in-memory relationships without a usable source
definition, suite-level absence of all steps definitions, and context
cancellation.

#### AC-DecoderErrors-1: Build YAML-source errors with `errs.Definition`

Given a decoding or validation failure attributable to a YAML file and member,
when a Decoder method returns the error, then its filename, one-based line, YAML
path, static message, and optional attached cause follow the `errs.Definition`
contract.

#### AC-DecoderErrors-2: Preserve decoding causes

Given a duplicate-key, unknown-field, YAML-syntax, or typed-decoding cause, when
the Decoder returns its definition error, then `errors.Is` and `errors.As`
continue through the wrapper produced by `errs.Definition`.

#### AC-DecoderErrors-3: Keep non-definition errors outside definition format

Given cancellation or an input-tree failure without an attributable YAML
member, when the Decoder returns the error, then it does not fabricate arguments
for `errs.Definition`.

## Types specification

### `models.Directory`

The Decoder-visible directory contract is:

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
}
```

The Loader-owned fields are inputs. The Decoder-owned fields are outputs:

- `DefaultsFile` is the source of `DefaultsDefinition`. At the suite root its
  kind is `KindRoot`; in a descendant its kind is `KindDefaults`.
- `StepsFiles` is the ordered source list for `StepsDefinitions`.
- `DefaultsDefinition` is nil in a descendant with no defaults file. The suite
  root must have a non-nil value after successful decoding.
- `StepsDefinitions` contains one decoded definition for each `StepsFiles`
  entry and preserves the exact `StepsFiles` order.

Decoder traversal follows `Children` from the supplied root. It must not follow
`Parent` recursively because `Parent` and `Children` are bidirectional.

### `models.File`

The Decoder-visible file contract is:

```go
type File struct {
    Path      string
    Kind      DocumentKind
    Directory *Directory
    Bytes     []byte
}
```

`File.Bytes` is the sole input used to decode a definition. Decoder performs no
filesystem read for decoding; `errs.Definition` may read `File.Path` only after
a source-aware error must be formatted.
`File.Path` identifies the source in every file-specific decoding or validation
error.

The Decoder must not change any `File` field. A definition points to its source
through `Definition.File`; a `File` does not point back to a decoded definition.

### `models.DefaultsDefinition`

```go
type DefaultsDefinition struct {
    App      string       `yaml:"app"`
    Kind     DocumentKind `yaml:"kind"`
    Metadata Metadata     `yaml:"metadata"`
    Spec     Defaults     `yaml:"spec"`
    File     *File        `yaml:"-"`
}
```

One type represents both defaults-bearing document kinds:

```text
kind: root     -> DefaultsDefinition{Kind: KindRoot}
kind: defaults -> DefaultsDefinition{Kind: KindDefaults}
```

The decoded `App` must be exactly `apihydra`. The decoded `Kind` must exactly
match the source `File.Kind`. `File` must point to the exact Loader-created file
from which the definition was decoded.

### `models.StepsDefinition`

```go
type StepsDefinition struct {
    App      string       `yaml:"app"`
    Kind     DocumentKind `yaml:"kind"`
    Metadata Metadata     `yaml:"metadata"`
    Spec     struct {
        Steps []Step `yaml:"steps"`
    } `yaml:"spec"`
    File *File `yaml:"-"`
}
```

The decoded `App` must be exactly `apihydra`. The decoded `Kind` must be
`KindSteps` and must exactly match the source `File.Kind`. `File` must point to
the exact Loader-created source file.

The order of `Spec.Steps` is the declaration order in YAML and must be
preserved.

### `models.Metadata`

```go
type Metadata struct {
    Name   string   `yaml:"name"`
    Labels []string `yaml:"labels"`
}
```

`metadata`, `metadata.name`, and `metadata.labels` are optional. When present,
`metadata.labels` must decode as an array whose every item is a string.

Metadata does not alter unfiltered decoding or validation. Selection by name or
labels belongs to a later service.

### `models.Defaults`

```go
type Defaults struct {
    BaseURL  string            `yaml:"baseUrl"`
    BasePath string            `yaml:"basePath"`
    Headers  map[string]string `yaml:"headers"`
    Timeout  int               `yaml:"timeout"`
    Retries  int               `yaml:"retries"`
}
```

All defaults fields are optional declarations. Omission is distinct from an
explicit invalid zero or empty value when the PRD assigns different behavior to
those states.

The current Go zero values do not by themselves preserve YAML field presence.
Validation must therefore use the original syntax available through
`DefaultsDefinition.File.Bytes`, or another private presence-aware decoding
representation scoped to the validation call. Third-party YAML AST types must
not become shared model fields.

### `models.Step`

The Decoder decodes the complete declarative step schema represented by
`models.Step`, including:

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
response.capture
response.types
response.expected
```

Dynamic mapping keys beneath `vars`, `request.headers`, `response.capture`, and
`response.types` are data, not schema field names. Strict unknown-field checks
still apply to every enclosing schema-defined mapping.

As with defaults, validation must distinguish omitted optional fields from
explicit empty or zero values whenever the PRD distinguishes them. The
implementation may use private syntax information obtained from the source
`File.Bytes`; it must not expose a third-party YAML AST through shared models.

### `models.YAMLString`

`YAMLString` preserves the literal scalar text used for request bodies and
expected response values.

The Decoder must not parse, normalize, compact, format, substitute, or validate
that text as JSON. JSON validation occurs after variable substitution through
the service that owns the required `jq` invocation.

## Method specification

## `NewDecoder`

```go
func NewDecoder() *Decoder
```

### Input

None.

### Output

- A non-nil, stateless `*Decoder`.

### Behavior

`NewDecoder` performs no traversal, decoding, validation, filesystem access, or
external-tool access.

### Acceptance criteria

#### AC-NewDecoder-1: Create a stateless decoder

When `NewDecoder` is called, then it returns a non-nil Decoder with no retained
suite or per-call state.

## `DecodeFiles`

```go
func (d *Decoder) DecodeFiles(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: a directory hierarchy successfully loaded and classified by
  `Loader.DecodeBaseDefinitions`.

For every traversed directory, Loader classification must satisfy:

```text
DefaultsFile == nil
```

or:

```text
DefaultsFile.Directory == directory
DefaultsFile.Kind == KindRoot     // supplied root only
DefaultsFile.Kind == KindDefaults // descendants only
```

Every entry in `StepsFiles` must have `KindSteps` and must point back to the
containing directory.

### Output

- `nil` when every classified file has been strictly decoded and all decoded
  definitions have been committed.
- An error when the input tree is invalid, a classified file cannot be decoded,
  a strict document rule fails, a suite-wide metadata rule fails, or the
  context is cancelled.

Successful output is represented by mutation of exactly these fields:

```go
Directory.DefaultsDefinition
Directory.StepsDefinitions
```

### Traversal and source selection

The method traverses the supplied root and its descendants through
`Directory.Children`.

For each directory it decodes:

1. `Directory.DefaultsFile`, when non-nil.
2. Every `Directory.StepsFiles` entry in slice order.

It does not decode unclassified values directly from `Directory.Files` and does
not repeat Loader classification. Unrelated YAML therefore remains ignored.

### Strict document decoding

For each classified source file, `DecodeFiles` must:

1. Check the context.
2. Decode exactly one YAML document from `File.Bytes`.
3. Reject a second YAML document, including an empty document introduced by an
   additional document separator.
4. Reject duplicate mapping keys at any depth, even when the duplicate values
   are identical.
5. Decode with strict schema validation.
6. Reject every unknown or misspelled schema field at any depth.
7. Require schema-defined scalar, sequence, and mapping values to have types
   compatible with their model fields.
8. Require `spec` to be a mapping. When present, `metadata`, each step,
   `request`, and `response` must also be mappings; sequence and dynamic-map
   fields must use their declared YAML collection type.
9. Require decoded `app` and `kind` values to match the Loader classification.
10. Set the decoded definition's `File` field to the exact source file.

An error must identify `File.Path`. Duplicate-key and unknown-field errors must
also identify the duplicated key or unknown YAML field path. Other typed decode
errors must identify the most specific available YAML path.

Strict field checking does not apply to keys of schema-defined dynamic maps,
such as header names, variable names, capture names, or response type selectors.

### Kind-specific decoding

`KindRoot` and `KindDefaults` files decode into `DefaultsDefinition`.

```text
KindRoot
  -> root.DefaultsDefinition

KindDefaults
  -> descendant.DefaultsDefinition
```

A `KindSteps` file decodes into `StepsDefinition` and is appended to the
temporary definitions slice corresponding to its position in `StepsFiles`.

```text
directory.StepsFiles[i]
  -> directory.StepsDefinitions[i]
  -> directory.StepsDefinitions[i].File == directory.StepsFiles[i]
```

### Metadata validation

After all files have decoded successfully, the method must inspect every
explicitly declared `metadata.name` across all root, defaults, and steps
definitions. Each supplied name must be unique within the complete suite tree.
Name comparison is exact and case-sensitive.

A duplicate name is a configuration error that identifies the name and both
source file paths. An explicitly declared empty name participates like any other
declared value; an omitted name does not. Labels require no semantic validation
beyond strict decoding as an array of strings.

Metadata validation does not filter or reorder definitions.

### Mutation behavior

All definitions and the suite-wide metadata-name index are prepared in
temporary state. Only after every classified file and metadata name passes must
the method commit the complete decoded state.

On success:

- Every directory's `DefaultsDefinition` replaces its pre-call value with the
  definition decoded from `DefaultsFile`, or with nil when `DefaultsFile` is
  nil.
- Every directory's `StepsDefinitions` replaces its pre-call value with the
  definitions decoded from `StepsFiles` in matching order.
- Empty `StepsFiles` produces an empty `StepsDefinitions` result.
- Repeated successful calls replace decoded definition pointers rather than
  appending duplicate definitions.

The method must not modify:

```text
Directory.Stage
Directory.Path
Directory.Parent
Directory.Children
Directory.Files
Directory.DefaultsFile
Directory.StepsFiles

File.Path
File.Kind
File.Directory
File.Bytes
```

If decoding, metadata validation, input validation, or cancellation fails, all
pre-call `DefaultsDefinition` and `StepsDefinitions` fields remain unchanged.

### Acceptance criteria

#### AC-DecodeFiles-1: Decode the root as defaults

Given a classified root file with a valid root schema, when files are decoded,
then `root.DefaultsDefinition` contains its decoded metadata and defaults,
retains `KindRoot`, and points to the exact root `DefaultsFile`.

#### AC-DecodeFiles-2: Decode descendant defaults

Given a descendant with one classified defaults file, when files are decoded,
then that directory's `DefaultsDefinition` contains its decoded values, retains
`KindDefaults`, and points to the exact source file.

#### AC-DecodeFiles-3: Preserve a missing descendant default

Given a descendant whose `DefaultsFile` is nil, when files are decoded, then
its `DefaultsDefinition` is nil.

#### AC-DecodeFiles-4: Decode ordered steps definitions

Given a directory with multiple classified steps files, when files are decoded,
then it receives one `StepsDefinition` per source file, in exact `StepsFiles`
order, and each definition points to its corresponding source file.

#### AC-DecodeFiles-5: Preserve step declaration order

Given a steps document containing multiple steps, when it is decoded, then its
`Spec.Steps` order exactly matches YAML declaration order.

#### AC-DecodeFiles-6: Decode optional metadata

Given definitions with omitted metadata, omitted metadata fields, or valid name
and string-label values, when files are decoded, then all forms succeed and the
declared values are preserved.

#### AC-DecodeFiles-7: Reject non-string metadata labels

Given `metadata.labels` that is not an array of strings, when files are decoded,
then the method returns an error identifying the source file and
`metadata.labels` path and commits no decoded state.

#### AC-DecodeFiles-8: Require one YAML document

Given a classified APIHydra file containing a second YAML document, when files
are decoded, then the method returns an error identifying the file and commits
no decoded state.

#### AC-DecodeFiles-9: Reject duplicate mapping keys

Given a duplicate mapping key anywhere in a classified document, when files are
decoded, then the method returns an error identifying the file and duplicated
key and commits no decoded state.

#### AC-DecodeFiles-10: Reject unknown fields deeply

Given an unknown field at the document, metadata, spec, step, request, or
response level, when files are decoded, then the method returns an error
identifying the file and complete YAML field path and commits no decoded state.

#### AC-DecodeFiles-11: Reject incompatible YAML types

Given a known field whose YAML value cannot decode to the field's declared
model type, when files are decoded, then the method returns a path-aware error
and commits no decoded state.

#### AC-DecodeFiles-12: Require classification consistency

Given a classified file whose fully decoded `app` or `kind` does not match its
Loader classification, when files are decoded, then the method returns an
error identifying the file and commits no decoded state.

#### AC-DecodeFiles-13: Ignore unclassified files

Given unrelated YAML retained only in `Directory.Files`, when classified files
are decoded, then the unrelated file is not decoded and produces no definition.

#### AC-DecodeFiles-14: Reject duplicate metadata names

Given any two decoded root, defaults, or steps definitions with the same
explicitly declared `metadata.name`, including an explicitly empty name, when
files are decoded, then the method returns an error identifying the name and
both source files and commits no decoded state.

#### AC-DecodeFiles-15: Replace rather than append on repeat calls

Given a tree with previously decoded definitions, when decoding succeeds again,
then definition fields are replaced atomically and no steps definition is
duplicated.

#### AC-DecodeFiles-16: Reject a nil root

Given a nil root, when files are decoded, then the method returns an error and
does not panic.

#### AC-DecodeFiles-17: Reject an invalid classified tree

Given a classified file whose kind, containing directory pointer, or directory
role is inconsistent with Loader's contract, when files are decoded, then the
method returns an error and commits no decoded state.

#### AC-DecodeFiles-18: Keep decoding transactional

Given a decoding or metadata failure after earlier definitions have decoded,
when the method returns, then none of the newly decoded definitions are
committed anywhere in the tree.

#### AC-DecodeFiles-19: Honor cancellation transactionally

Given a cancelled context, when the method is called, then it returns
`context.Canceled` and commits no decoded state. Given cancellation during
traversal, it stops at the next cancellation check and preserves all pre-call
definition fields.

## `ValidateDefaultsDefinitions`

```go
func (d *Decoder) ValidateDefaultsDefinitions(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: a directory hierarchy successfully enriched by `DecodeFiles`.

The supplied root must contain a `DefaultsDefinition` backed by its classified
root file. Each descendant must contain either no defaults definition and no
defaults file, or one definition backed by its classified defaults file.

### Output

- `nil` when every root/defaults definition passes all static defaults
  validation.
- An error identifying the source file and YAML path when a definition is
  missing, inconsistent, invalid, or validation is cancelled.

The method is read-only and returns the first error in deterministic directory
traversal order.

### Validation rules

For the suite root and every descendant defaults definition, the method must
validate:

1. The definition and its source file remain mutually consistent with the
   containing directory and expected document kind.
2. An explicitly declared `spec.basePath` is not an empty string.
3. An explicitly declared `spec.timeout` is a positive integer.
4. An explicitly declared `spec.retries` is a non-negative integer. Zero is
   valid and means no retries beyond the initial attempt.

Omitted fields are valid. In particular:

- An omitted `baseUrl` may later be supplied by a nearer definition or step.
- An omitted `basePath` contributes no local override.
- An omitted `timeout` may be inherited or receive the built-in default of 10.
- An omitted `retries` may be inherited or receive the built-in default of 3.
- An omitted headers map is valid.

The method must distinguish an omitted numeric field from an explicitly
declared zero. It must not reject an omitted `timeout` merely because the Go
field has its zero value, and it must accept an explicitly declared
`retries: 0`.

The method does not validate:

- Whether a final resolved `baseUrl` exists.
- URL syntax or URL composition.
- Parent-default discovery or inheritance.
- Header merging, case-insensitive comparison, or canonicalization.
- Whether a defaults value is ultimately used by a selected step.

### Acceptance criteria

#### AC-ValidateDefaultsDefinitions-1: Accept an empty root spec

Given a decoded root definition with no declared defaults fields, when defaults
definitions are validated, then the method succeeds because resolved defaults
may be supplied later.

#### AC-ValidateDefaultsDefinitions-2: Accept a descendant without defaults

Given a decoded descendant with no defaults file or definition, when defaults
definitions are validated, then the method succeeds for that directory.

#### AC-ValidateDefaultsDefinitions-3: Accept valid declared defaults

Given root or descendant defaults declaring non-empty base path, string headers,
positive timeout, and non-negative retries, when defaults definitions are
validated, then the method succeeds.

#### AC-ValidateDefaultsDefinitions-4: Reject an empty declared base path

Given a defaults definition explicitly declaring `spec.basePath: ""`, when
defaults definitions are validated, then the method returns an error identifying
the file and `spec.basePath`.

#### AC-ValidateDefaultsDefinitions-5: Reject a non-positive declared timeout

Given a defaults definition explicitly declaring a zero or negative timeout,
when defaults definitions are validated, then the method returns an error
identifying the file and `spec.timeout`.

#### AC-ValidateDefaultsDefinitions-6: Accept zero retries

Given a defaults definition explicitly declaring `spec.retries: 0`, when
defaults definitions are validated, then the method succeeds for that field.

#### AC-ValidateDefaultsDefinitions-7: Reject negative retries

Given a defaults definition explicitly declaring negative retries, when
defaults definitions are validated, then the method returns an error identifying
the file and `spec.retries`.

#### AC-ValidateDefaultsDefinitions-8: Distinguish omission from explicit zero

Given omitted timeout and retries fields, when defaults definitions are
validated, then omission is accepted and is not confused with an invalid
explicit timeout of zero.

#### AC-ValidateDefaultsDefinitions-9: Reject missing decoded state

Given a Loader-classified defaults file without its corresponding decoded
definition, when defaults definitions are validated, then the method returns an
error and does not panic.

#### AC-ValidateDefaultsDefinitions-10: Remain read-only on failure

Given any invalid defaults definition after earlier valid definitions, when the
method returns an error, then no directory, file, definition, metadata, or spec
field has changed.

#### AC-ValidateDefaultsDefinitions-11: Reject a nil root

Given a nil root, when defaults definitions are validated, then the method
returns an error and does not panic.

#### AC-ValidateDefaultsDefinitions-12: Honor cancellation read-only

Given a cancelled context, when defaults definitions are validated, then the
method returns `context.Canceled` and changes no state. Given cancellation
during traversal, it stops at the next cancellation check and changes no state.

## `ValidateStepsDefinitions`

```go
func (d *Decoder) ValidateStepsDefinitions(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: a directory hierarchy successfully enriched by `DecodeFiles`.

Every Loader-classified steps file must have one matching decoded steps
definition in the same slice position, with an exact source-file pointer.

### Output

- `nil` when the complete unfiltered suite contains at least one declared step
  and every decoded steps definition passes static steps validation.
- An error identifying the source file and most specific YAML path when decoded
  state is missing, a declaration is invalid, the suite has zero steps, or the
  context is cancelled.

The method is read-only and returns the first error in deterministic directory,
steps-definition, and step declaration order.

### Suite-level validation

The complete unfiltered suite must contain at least one declared step across
all `StepsDefinitions`. Zero steps is an error whether there are no steps
documents or every steps document contains an empty `spec.steps` sequence.

Filter-specific zero-match behavior belongs to the later selection service and
is not evaluated here.

### Step request validation

For every declared step, the method must validate rules decidable before
defaults inheritance or variable substitution:

1. `request.path` is present and is not empty. No defaults definition can
   provide a request path.
2. When `request.method` is explicitly declared, it is not empty. Any non-empty
   method is accepted exactly as written; the Decoder must not restrict methods
   to a known list or normalize their spelling.
3. When `request.basePath` is explicitly declared, it is not empty.
4. When `request.query` is explicitly declared, it is not empty.
5. When `request.timeout` is explicitly declared, it is a positive integer.
6. When `request.retries` is explicitly declared, it is a non-negative integer;
   an explicit zero is valid.

The method must distinguish omission from explicit empty or zero values.

It does not require a declared `request.baseUrl`, because that value may be
inherited. It does not choose the default request method, timeout, or retry
count; those decisions belong to resolved-step resolution.

### Step variable validation

Every key declared in `step.vars` must match:

```text
[A-Za-z_][A-Za-z0-9_:]*
```

Every declared value must be representable as JSON. Accepted values include
JSON-compatible YAML scalars, arrays, objects, and `null`. Validation must reject
a value that cannot be serialized as JSON.

The method validates serializability but does not replace the original value
with serialized text and does not write anything to the run-wide variable
store.

Duplicate assignments across steps are not rejected statically because
filtering and execution order determine which assignments occur. The write-once
variable store owns that runtime rule.

### Response capture validation

Every variable key declared under `response.capture` must match:

```text
[A-Za-z_][A-Za-z0-9_:]*
```

Capture values must decode as strings. Their contents are `jq` expressions and
are not parsed or executed by the Decoder.

### Response status validation

When `response.status` is present, it must be:

- A non-empty array.
- Composed only of integers.
- Free of duplicate values.
- Limited to values from 100 through 599 inclusive.

When `response.status` is omitted, no HTTP status declaration is required.

### Response type-declaration validation

Each value under `response.types` is an ordered declaration array.

The first item must be exactly one supported base type:

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

Only these modifiers may follow the base type:

```text
zero
optional
```

Each modifier may occur at most once and the two modifiers may appear in either
order. A declaration must be rejected when it:

- Is empty.
- Starts with a modifier or unknown token.
- Contains an unsupported base type.
- Contains another base type after the first position.
- Contains an unknown modifier.
- Repeats `zero` or `optional`.

The special runtime meaning of `[null]`, zero-value rules, selector cardinality,
datetime formats, and UUID formats apply to actual response values and belong
to response processing. This method validates only declaration structure.

Response type map keys are full `jq` selector expressions. The Decoder does not
parse or run them.

### Deferred validation

`ValidateStepsDefinitions` must not perform checks whose result depends on
resolved defaults, variables, external tools, or an actual response. In
particular, it does not:

- Require `request.baseUrl` before defaults resolution.
- Compose or parse the final URL.
- Merge or canonicalize headers.
- Apply the default method, timeout, or retries.
- Substitute variables in `request.body` or `response.expected`.
- Validate literal or substituted body/expected text as JSON.
- Validate `jq` syntax for capture or type selectors.
- Enforce cross-step variable write-once behavior.
- Require a JSON response or evaluate status, expected-value, or type matches.

### Acceptance criteria

#### AC-ValidateStepsDefinitions-1: Accept valid static step declarations

Given one or more decoded steps with valid paths, optional request fields,
variable declarations, and response declarations, when steps definitions are
validated, then the method succeeds without changing them.

#### AC-ValidateStepsDefinitions-2: Reject an unfiltered suite with zero steps

Given no steps documents or only steps documents with empty `spec.steps`, when
steps definitions are validated, then the method returns an error and changes
no state.

#### AC-ValidateStepsDefinitions-3: Require a request path

Given a step with an omitted or empty `request.path`, when steps definitions are
validated, then the method returns an error identifying the source file and the
step's `request.path`.

#### AC-ValidateStepsDefinitions-4: Preserve arbitrary explicit methods

Given a step declaring any non-empty request method, including a custom method,
when steps definitions are validated, then the method accepts it unchanged.

#### AC-ValidateStepsDefinitions-5: Reject an explicitly empty method

Given a step explicitly declaring an empty request method, when steps
definitions are validated, then the method returns an error identifying the
source file and the step's `request.method`.

#### AC-ValidateStepsDefinitions-6: Allow an omitted method

Given a step omitting `request.method`, when steps definitions are validated,
then the method accepts the omission without choosing `GET` or `POST`.

#### AC-ValidateStepsDefinitions-7: Reject empty declared optional request text

Given a step explicitly declaring an empty `request.basePath` or
`request.query`, when steps definitions are validated, then the method returns
an error identifying the exact field path.

#### AC-ValidateStepsDefinitions-8: Validate declared timeout and retries

Given an explicitly declared non-positive timeout or negative retries, when
steps definitions are validated, then the method returns a field-specific
error. Given omitted values or explicit `retries: 0`, then those declarations
are accepted.

#### AC-ValidateStepsDefinitions-9: Accept valid variable keys

Given variable keys beginning with a letter or underscore and continuing with
letters, digits, underscores, or colons, when steps definitions are validated,
then those keys are accepted.

#### AC-ValidateStepsDefinitions-10: Reject invalid variable keys

Given an invalid key under `vars` or `response.capture`, when steps definitions
are validated, then the method returns an error identifying the source file,
step, and key.

#### AC-ValidateStepsDefinitions-11: Validate variable JSON compatibility

Given JSON-compatible YAML scalar, array, object, or null values under `vars`,
when steps definitions are validated, then they are accepted. Given a value that
cannot be represented as JSON, then the method returns a configuration error
identifying its source.

#### AC-ValidateStepsDefinitions-12: Accept an omitted response

Given a step with no `response` section, when steps definitions are validated,
then the omission is accepted.

#### AC-ValidateStepsDefinitions-13: Accept an omitted status declaration

Given a response with no `status`, when steps definitions are validated, then
the omission is accepted and no default accepted status is introduced.

#### AC-ValidateStepsDefinitions-14: Validate response statuses

Given a present status array, when steps definitions are validated, then a
non-empty unique set of integers from 100 through 599 is accepted, while an
empty array, duplicate, or out-of-range value returns a field-specific error.

#### AC-ValidateStepsDefinitions-15: Accept valid response type declarations

Given each supported base type with no modifier, `zero`, `optional`, or both
modifiers in either order, when steps definitions are validated, then the
declarations are accepted.

#### AC-ValidateStepsDefinitions-16: Reject malformed response type declarations

Given an empty declaration, unknown token, duplicate modifier, missing leading
base type, multiple base types, or base type outside the first position, when
steps definitions are validated, then the method returns an error identifying
the source file, step, and type selector.

#### AC-ValidateStepsDefinitions-17: Defer JSON and jq validation

Given a request body, expected response, capture expression, or type selector
whose runtime validity has not yet been established, when steps definitions are
validated, then the method does not invoke external tools or reject it on the
basis of JSON or jq syntax alone.

#### AC-ValidateStepsDefinitions-18: Reject missing decoded state

Given a Loader-classified steps file without a corresponding decoded definition,
or with an incorrect definition-to-file association, when steps definitions are
validated, then the method returns an error and does not panic.

#### AC-ValidateStepsDefinitions-19: Remain read-only on failure

Given an invalid step after earlier valid steps, when the method returns an
error, then no directory, file, definition, metadata, spec, or step field has
changed.

#### AC-ValidateStepsDefinitions-20: Reject a nil root

Given a nil root, when steps definitions are validated, then the method returns
an error and does not panic.

#### AC-ValidateStepsDefinitions-21: Honor cancellation read-only

Given a cancelled context, when steps definitions are validated, then the
method returns `context.Canceled` and changes no state. Given cancellation
during traversal, it stops at the next cancellation check and changes no state.
