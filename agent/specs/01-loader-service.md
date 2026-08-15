# internal.definition.Loader Service

## Status

- Service: `internal/definition.Loader`
- Package: `internal/definition`
- Package name: `definition`
- Shared models: `internal/models`
- Error construction: `pkg/errs`
- Status: implementation specification

## Base specification

### Purpose

`Loader` loads the filesystem representation of an APIHydra suite before
kind-specific YAML parsing begins.

It performs three explicit operations:

1. Load the directory hierarchy rooted at the configured working directory.
2. Load YAML files and their bytes into that hierarchy.
3. Decode the base definition of every loaded file, validate document placement,
   and classify recognized APIHydra definitions.

The service produces and enriches a directory-oriented model:

```text
WorkDir
  -> LoadDirectoryStructure
       -> *models.Directory
  -> LoadDirectoryFiles
       -> Directory.Files and File.Bytes
  -> DecodeBaseDefinitions
       -> File.Kind
       -> Directory.DefaultsFile
       -> Directory.StepsFiles
```

### Responsibilities

`Loader` owns:

- Filesystem directory traversal.
- Directory parent, child, and stage relationships.
- Discovery of regular `.yaml` and `.yml` files.
- Reading the exact bytes of discovered YAML files.
- Shallow decoding of `app`, `kind`, and `spec` into `BaseDefinition`.
- Recognition of files declaring `app: apihydra`.
- Base validation of recognized APIHydra definitions.
- Classification of root, defaults, and steps definitions.
- Root/defaults placement validation that can be decided from base definitions.
- Construction of YAML-source errors through `errs.Definition`.

### Non-responsibilities

`Loader` does not:

- Fully decode root, defaults, or steps specifications.
- Validate kind-specific fields inside `spec`.
- Decode metadata.
- Resolve defaults inheritance.
- Resolve defaults or steps.
- Filter steps by metadata.
- Plan execution stages beyond recording directory depth in `Directory.Stage`.
- Execute requests or external tools.
- Produce terminal or machine-readable output.

### Public contract

```go
type Loader struct {
    WorkDir string
}

func NewLoader(workDir string) *Loader

func (l *Loader) LoadDirectoryStructure(
    ctx context.Context,
) (*models.Directory, error)

func (l *Loader) LoadDirectoryFiles(
    ctx context.Context,
    root *models.Directory,
) error

func (l *Loader) DecodeBaseDefinitions(
    ctx context.Context,
    root *models.Directory,
) error
```

The existing flat `Files() ([]string, error)` operation is not part of this
target contract. File discovery is represented through `Directory.Files`.

### Service state

`Loader` retains only stable constructor configuration:

```go
WorkDir string
```

It must not retain:

- A context.
- A directory tree.
- Loaded files or file bytes.
- Decoded base definitions.
- Per-call errors or traversal state.

Contexts and all mutable traversal state are scoped to individual method calls.

### Cancellation

Every Loader method that may traverse directories, inspect files, read bytes, or
iterate over a loaded tree accepts `context.Context` as its first argument.

Go filesystem calls may not be interruptible while one syscall is in progress.
The Loader must nevertheless check `ctx.Err()` before beginning work and at
useful boundaries throughout traversal so cancellation prevents additional
filesystem or decoding work.

### Mutation policy

- `LoadDirectoryStructure` creates and returns a new directory tree.
- `LoadDirectoryFiles` mutates only file-related fields in the supplied tree.
- `DecodeBaseDefinitions` mutates only classification fields in the supplied
  tree.
- No Loader method mutates `Loader.WorkDir`.
- A method returning an error must not expose a partially applied mutation from
  that method. Work must be collected in temporary state and committed only
  after that method has completed successfully.

### Error construction

Every Loader error attributable to a specific YAML definition must be built
with:

```go
errs.Definition(file.Path, yamlPath, staticMessage, cause)
```

The Loader supplies:

- The exact `File.Path` as the filename.
- The most specific YAML path available.
- A stable, non-empty static message describing the Loader-owned rule.
- The underlying decoding error as `cause` when one exists, or nil for a pure
  classification or placement rule.

Document-wide errors use `$`. Field errors use paths such as `$.app`, `$.kind`,
or `$.spec`. Placement and cardinality errors use the offending definition's
`$.kind` path. When multiple files conflict, deterministic traversal chooses
the primary offending file and the static message identifies the other
conflicting paths.

Errors without a YAML definition location do not use `errs.Definition`. These
include invalid working directories, directory traversal and file-reading
failures before definition decoding, nil or inconsistent in-memory trees,
missing-root errors with no candidate definition, and context cancellation.

#### AC-LoaderErrors-1: Build YAML-source errors with `errs.Definition`

Given a Loader failure attributable to a YAML file and member, when the Loader
returns the error, then its filename, one-based line, YAML path, static message,
and optional attached cause follow the `errs.Definition` contract.

#### AC-LoaderErrors-2: Keep non-definition errors outside definition format

Given a filesystem, context, or structural failure without an attributable YAML
member, when the Loader returns the error, then it does not fabricate arguments
for `errs.Definition`.

## Types specification

### `models.DocumentKind`

```go
type DocumentKind string

const (
    KindRoot     DocumentKind = "root"
    KindDefaults DocumentKind = "defaults"
    KindSteps    DocumentKind = "steps"
)
```

The zero value means that a loaded YAML file has not been recognized as an
APIHydra definition.

### `models.Directory`

The Loader-visible directory contract is:

```go
type Directory struct {
    Stage        int
    Path         string
    Parent       *Directory
    Children     []*Directory
    Files        []*File
    DefaultsFile *File
    StepsFiles   []*File
}
```

Field meanings:

- `Stage` is the directory depth relative to the configured root. The root is
  stage `0`; each child is its parent's stage plus one.
- `Path` is the cleaned filesystem path of the directory.
- `Parent` is `nil` only for the returned root directory.
- `Children` contains the directory's immediate child directories.
- `Files` contains every loaded regular `.yaml` or `.yml` file located directly
  in the directory, including unrelated YAML files.
- `DefaultsFile` is populated by `DecodeBaseDefinitions`. At the root it points
  to the unique root definition. In a descendant it is `nil` or points to the
  directory's defaults definition.
- `StepsFiles` contains zero or more recognized steps definitions located
  directly in the directory.

`Parent` and `Children` form bidirectional references. Directory traversal must
therefore avoid recursively following both directions without cycle control.

### `models.File`

The Loader-visible file contract is:

```go
type File struct {
    Path      string
    Kind      DocumentKind
    Directory *Directory
    Bytes     []byte
}
```

Field meanings:

- `Path` is the cleaned filesystem path of the file.
- `Kind` is initially empty. `DecodeBaseDefinitions` sets it only for recognized
  APIHydra definitions.
- `Directory` points to the file's immediate containing directory.
- `Bytes` contains the exact bytes read from the file.

`Directory.Files` and `File.Directory` form bidirectional references.

Execution stage is owned by `Directory.Stage`. A file obtains its stage through
`File.Directory.Stage`; a separate file stage is not part of the Loader
contract.

### `models.BaseDefinition`

```go
type BaseDefinition struct {
    App  string     `yaml:"app"`
    Kind string     `yaml:"kind"`
    Spec YAMLString `yaml:"spec"`
}
```

`BaseDefinition` is a temporary shallow-decoding carrier. It is not retained on
`File` or `Directory`.

- `App` is used to distinguish APIHydra definitions from unrelated YAML.
- `Kind` remains a string during shallow decoding so unsupported values can be
  diagnosed before conversion to `DocumentKind`.
- `Spec` carries the raw YAML representation of the `spec` value. It is not a
  kind-specific decoded specification at this phase.

A `spec` is base-valid when it exists and its raw representation is non-empty
after surrounding whitespace is ignored. Kind-specific semantic validation is
deferred to later decoding services. In particular, an empty mapping such as
`spec: {}` has a non-empty raw representation and passes base validation.

The YAML implementation may use a private raw-message or YAML-node type while
decoding. Third-party YAML AST types must not become shared model fields.

## Method specification

## `NewLoader`

```go
func NewLoader(workDir string) *Loader
```

### Input

- `workDir`: the filesystem directory selected as the APIHydra suite root.

### Output

- A non-nil `*Loader` retaining `workDir` as configuration.

### Behavior

`NewLoader` stores its argument. It performs no filesystem access and does not
validate whether the path exists or is a directory.

### Acceptance criteria

#### AC-NewLoader-1: Retain the configured working directory

Given any working-directory string, when `NewLoader` is called, then it returns
a non-nil Loader whose `WorkDir` equals the supplied string.

#### AC-NewLoader-2: Perform no filesystem validation

Given a path that does not exist, when `NewLoader` is called, then construction
succeeds without filesystem access; path validation occurs when directory
loading is requested.

## `LoadDirectoryStructure`

```go
func (l *Loader) LoadDirectoryStructure(
    ctx context.Context,
) (*models.Directory, error)
```

### Input

- `ctx`: cancellation and deadline signal.
- `l.WorkDir`: the configured root path.

### Output

- The root `*models.Directory` of a complete directory hierarchy.
- An error when the root is invalid, traversal fails, or the context is
  cancelled.

### Behavior

`LoadDirectoryStructure` recursively loads physical directories rooted at
`Loader.WorkDir`.

It must:

1. Check the context before filesystem work.
2. Require `WorkDir` to exist and be a directory.
3. Clean paths using the platform filesystem path rules.
4. Create one `Directory` value for the root with `Stage == 0` and
   `Parent == nil`.
5. Create one `Directory` value for every physical descendant directory.
6. Connect each child to its immediate parent in both directions.
7. Set each child's stage to `parent.Stage + 1`.
8. Sort each directory's `Children` deterministically by path.
9. Return the root directory.

It must not create `File` values, read file bytes, or decode YAML. On success,
all file and classification fields are empty:

```go
directory.Files == nil or empty
directory.DefaultsFile == nil
directory.StepsFiles == nil or empty
```

Directory symlinks are not followed as physical descendant directories.

### Acceptance criteria

#### AC-LoadDirectoryStructure-1: Load the configured root

Given an existing empty `WorkDir`, when `LoadDirectoryStructure` is called,
then it returns a root directory whose path is the cleaned `WorkDir`, stage is
zero, parent is nil, and children and files are empty.

#### AC-LoadDirectoryStructure-2: Load nested directories recursively

Given physical directories at multiple depths beneath `WorkDir`, when the
method is called, then every descendant directory appears exactly once under
its immediate parent.

#### AC-LoadDirectoryStructure-3: Populate bidirectional relationships

Given a loaded child directory, when the returned tree is inspected, then the
parent contains the child in `Parent.Children` and the child's `Parent` points
to that exact parent object.

#### AC-LoadDirectoryStructure-4: Assign stages by relative depth

Given root, child, and grandchild directories, when the tree is loaded, then
their stages are `0`, `1`, and `2` respectively.

#### AC-LoadDirectoryStructure-5: Return deterministic child order

Given the same filesystem hierarchy across repeated calls, when results are
inspected, then every `Children` slice is ordered lexicographically by cleaned
path.

#### AC-LoadDirectoryStructure-6: Do not load files

Given directories containing files, when only `LoadDirectoryStructure` is
called, then no `File` values or bytes are added to the returned tree.

#### AC-LoadDirectoryStructure-7: Reject a missing root

Given a `WorkDir` that does not exist, when the method is called, then it
returns a nil root and a non-nil error.

#### AC-LoadDirectoryStructure-8: Reject a non-directory root

Given a `WorkDir` that identifies a regular file, when the method is called,
then it returns a nil root and a non-nil error.

#### AC-LoadDirectoryStructure-9: Propagate traversal failure atomically

Given a filesystem error while loading descendant directories, when the method
returns, then it returns a nil root and an error rather than a partial tree.

#### AC-LoadDirectoryStructure-10: Honor cancellation

Given a cancelled context, when the method is called, then it returns no root,
returns `context.Canceled`, and performs no additional traversal. Given
cancellation during traversal, it stops at the next cancellation check and
does not return a partial tree.

#### AC-LoadDirectoryStructure-11: Do not follow directory symlinks

Given a symlink to a directory beneath `WorkDir`, when the structure is loaded,
then the symlink target is not added as a physical child directory and cannot
cause recursive traversal outside the configured root.

## `LoadDirectoryFiles`

```go
func (l *Loader) LoadDirectoryFiles(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: the directory hierarchy returned by `LoadDirectoryStructure`.

### Output

- `nil` on success.
- An error when the input is invalid, directory inspection or file reading
  fails, or the context is cancelled.

The successful output is represented by mutation of the supplied tree.

### Behavior

For every directory in the supplied hierarchy, `LoadDirectoryFiles` must:

1. Check the context before inspecting the directory and before reading each
   file.
2. Inspect only files located directly at `Directory.Path`; recursion follows
   the supplied `Directory.Children` hierarchy.
3. Select regular files whose extension is exactly `.yaml` or `.yml`.
4. Ignore directories, symlinks, non-regular entries, and other extensions.
5. Read each selected file's exact bytes.
6. Create one `File` per selected filesystem entry.
7. Set `File.Path` to the cleaned file path.
8. Set `File.Directory` to the exact containing directory object.
9. Set `File.Bytes` to the exact bytes read from disk.
10. Leave `File.Kind` empty.
11. Sort each directory's `Files` deterministically by path.
12. Commit all `Directory.Files` slices only after the complete operation
    succeeds.

It must not populate `Directory.DefaultsFile` or `Directory.StepsFiles` and must
not decode YAML.

### Acceptance criteria

#### AC-LoadDirectoryFiles-1: Load YAML and YML files

Given regular `.yaml` and `.yml` files directly in a loaded directory, when
`LoadDirectoryFiles` is called, then one `File` value is added for each matching
filesystem entry.

#### AC-LoadDirectoryFiles-2: Load files in every directory

Given matching files in root and descendant directories, when the method is
called, then every matching file is placed in the `Files` slice of its immediate
containing directory.

#### AC-LoadDirectoryFiles-3: Connect files to their directory

Given a loaded file, when the result is inspected, then `File.Directory` points
to the exact directory whose `Files` slice contains that file.

#### AC-LoadDirectoryFiles-4: Preserve exact file bytes

Given a matching file with arbitrary byte content, when it is loaded, then
`File.Bytes` is byte-for-byte equal to the content read from that path.

#### AC-LoadDirectoryFiles-5: Leave files unclassified

Given a loaded YAML file containing a valid APIHydra definition, when only
`LoadDirectoryFiles` has completed, then `File.Kind` remains empty and the
directory's `DefaultsFile` and `StepsFiles` remain empty.

#### AC-LoadDirectoryFiles-6: Ignore non-YAML entries

Given non-YAML files, directories, symlinks, and non-regular entries, when the
method is called, then none of those entries is represented in
`Directory.Files`.

#### AC-LoadDirectoryFiles-7: Return deterministic file order

Given the same files across repeated calls on equivalent fresh directory trees,
when results are inspected, then each `Files` slice is ordered lexicographically
by cleaned path.

#### AC-LoadDirectoryFiles-8: Reject a nil root

Given a nil root, when the method is called, then it returns an error and
performs no filesystem work.

#### AC-LoadDirectoryFiles-9: Report directory inspection failure atomically

Given an error while inspecting a directory represented in the tree, when the
method returns, then it reports the error and no `Directory.Files` slice in the
tree has been modified.

#### AC-LoadDirectoryFiles-10: Report file reading failure atomically

Given an error while reading a selected YAML file, when the method returns,
then it reports the file path and underlying error and no `Directory.Files`
slice in the tree has been modified.

#### AC-LoadDirectoryFiles-11: Honor cancellation atomically

Given a cancelled context, when the method is called, then it returns
`context.Canceled` and does not modify the tree. Given cancellation during
inspection or reading, it stops at the next cancellation check and commits no
partial file state.

## `DecodeBaseDefinitions`

```go
func (l *Loader) DecodeBaseDefinitions(
    ctx context.Context,
    root *models.Directory,
) error
```

### Input

- `ctx`: cancellation and deadline signal.
- `root`: a directory hierarchy whose files and bytes were populated by
  `LoadDirectoryFiles`.

### Output

- `nil` when every file has been shallowly decoded, all recognized APIHydra
  definitions pass base validation, and all classifications have been committed.
- An error when decoding, base validation, placement validation, or cancellation
  fails.

The successful output is represented by mutation of exactly these fields:

```go
Directory.DefaultsFile
Directory.StepsFiles
File.Kind
```

### Base decoding

For every value in every `Directory.Files` slice, the method must:

1. Check the context.
2. Decode `File.Bytes` into a temporary `BaseDefinition`.
3. If `App` is not exactly `apihydra`, treat the file as unrelated YAML, leave
   `File.Kind` empty, and do not classify it.
4. If `App` is exactly `apihydra`, require `Kind` to be exactly `root`,
   `defaults`, or `steps`.
5. If `App` is exactly `apihydra`, require the raw `Spec` representation to be
   non-empty after surrounding whitespace is ignored.
6. Convert a valid base kind to `DocumentKind` in temporary classification
   state.

A YAML decoding failure must identify the corresponding `File.Path`.

### Placement and cardinality validation

The directory supplied as `root` is the suite-root directory.

The method must enforce:

1. The root directory contains exactly one recognized root definition.
2. The root directory contains no recognized defaults definition.
3. No descendant directory contains a root definition.
4. A descendant directory contains zero or one recognized defaults definition.
5. Every directory, including the root, contains zero or more recognized steps
   definitions.

On successful classification:

```go
root.DefaultsFile.Kind == models.KindRoot
root.DefaultsFile.Directory == root
```

For every descendant directory:

```go
directory.DefaultsFile == nil
```

or:

```go
directory.DefaultsFile.Kind == models.KindDefaults
directory.DefaultsFile.Directory == directory
```

For every classified steps file:

```go
stepFile.Kind == models.KindSteps
stepFile.Directory == directory
```

### Mutation behavior

Base definitions are decoded and validated into temporary state first. Only
after every file and the complete tree pass validation may the method commit:

- `Kind` for each recognized `File`.
- The root definition to `root.DefaultsFile`.
- Each descendant defaults definition to its directory's `DefaultsFile`.
- Every steps definition to its directory's `StepsFiles`.

Unrelated YAML files remain in `Directory.Files`, retain an empty `File.Kind`,
and appear in neither classification field.

The method must not modify:

```go
Directory.Stage
Directory.Path
Directory.Parent
Directory.Children
Directory.Files

File.Path
File.Directory
File.Bytes
```

`StepsFiles` must follow the deterministic path ordering established by
`Directory.Files`.

### Acceptance criteria

#### AC-DecodeBaseDefinitions-1: Ignore unrelated YAML

Given a loaded YAML file whose `app` field is missing or is not exactly
`apihydra`, when base definitions are decoded, then the method succeeds for
that file, leaves its `Kind` empty, leaves it in `Directory.Files`, and does not
add it to `DefaultsFile` or `StepsFiles`.

#### AC-DecodeBaseDefinitions-2: Require one root definition

Given a root directory containing exactly one valid `app: apihydra`,
`kind: root` definition and no defaults definition, when base definitions are
decoded, then the method succeeds and `root.DefaultsFile` points to that exact
file.

#### AC-DecodeBaseDefinitions-3: Reject a missing root definition

Given a root directory containing no recognized root definition, when the
method is called, then it returns an error and commits no classification state.

#### AC-DecodeBaseDefinitions-4: Reject multiple root definitions

Given two or more recognized root definitions in the root directory, when the
method is called, then it returns an error identifying the conflicting files
and commits no classification state.

#### AC-DecodeBaseDefinitions-5: Reject defaults in the root directory

Given a recognized defaults definition directly in the root directory, when
the method is called, then it returns an error identifying that file and
commits no classification state, whether or not a root definition is also
present.

#### AC-DecodeBaseDefinitions-6: Reject nested roots

Given a recognized root definition in any descendant directory, when the
method is called, then it returns an error identifying that file and commits no
classification state.

#### AC-DecodeBaseDefinitions-7: Accept zero descendant defaults files

Given a descendant directory with no defaults definition, when otherwise valid
base definitions are decoded, then the method succeeds and that directory's
`DefaultsFile` remains nil.

#### AC-DecodeBaseDefinitions-8: Classify one descendant defaults file

Given a descendant directory with one valid defaults definition, when base
definitions are decoded, then the method sets the file's kind to
`KindDefaults` and sets that directory's `DefaultsFile` to the exact file.

#### AC-DecodeBaseDefinitions-9: Reject multiple descendant defaults files

Given a descendant directory with two or more recognized defaults definitions,
when the method is called, then it returns an error identifying the conflicting
files and commits no classification state.

#### AC-DecodeBaseDefinitions-10: Accept zero steps files

Given any directory containing no recognized steps definitions, when otherwise
valid base definitions are decoded, then the method succeeds and that
directory's `StepsFiles` remains empty.

#### AC-DecodeBaseDefinitions-11: Classify multiple steps files

Given any directory containing multiple valid steps definitions, when base
definitions are decoded, then every file has `KindSteps`, every file appears
exactly once in that directory's `StepsFiles`, and the order matches
`Directory.Files` path order.

#### AC-DecodeBaseDefinitions-12: Preserve root kind while assigning defaults role

Given the valid root definition, when classification succeeds, then its
`File.Kind` is `KindRoot` even though the file is referenced through
`root.DefaultsFile`.

#### AC-DecodeBaseDefinitions-13: Reject an unsupported APIHydra kind

Given a file declaring `app: apihydra` and an unsupported or empty `kind`, when
the method is called, then it returns an error identifying the file and kind
and commits no classification state.

#### AC-DecodeBaseDefinitions-14: Require a base spec

Given a file declaring `app: apihydra` with a valid kind but a missing or empty
raw `spec`, when the method is called, then it returns an error identifying the
file and commits no classification state.

#### AC-DecodeBaseDefinitions-15: Defer kind-specific spec validation

Given a recognized APIHydra definition whose `spec` has a non-empty raw YAML
representation, when base definitions are decoded, then base classification
does not reject it merely because its kind-specific fields have not yet been
decoded or validated.

#### AC-DecodeBaseDefinitions-16: Report malformed YAML

Given bytes that cannot be decoded into a base YAML definition, when the method
is called, then it returns an error identifying `File.Path` and commits no
classification state.

#### AC-DecodeBaseDefinitions-17: Mutate only classification fields

Given a valid loaded directory tree, when base decoding succeeds, then
`Directory.DefaultsFile`, `Directory.StepsFiles`, and recognized `File.Kind`
values contain the classifications while all paths, stages, relationships,
file lists, and bytes remain unchanged.

#### AC-DecodeBaseDefinitions-18: Keep validation transactional

Given any base decoding or placement failure after earlier files have been
successfully examined, when the method returns, then no file kind or directory
classification from the attempted call has been committed.

#### AC-DecodeBaseDefinitions-19: Honor cancellation transactionally

Given a cancelled context, when the method is called, then it returns
`context.Canceled` and commits no classification state. Given cancellation
during traversal, it stops at the next cancellation check and commits no
partial classification state.
