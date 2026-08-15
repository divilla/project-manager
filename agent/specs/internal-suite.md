# internal.suite

## Instructions

- Preserve the exported signatures in `internal/suite` unless a later Change
  explicitly replaces them.
- Keep YAML discovery and suite-tree parsing separate from defaults resolution.
- Keep requirements and acceptance criteria grouped under the type they govern.

## Loader type

- Path: `internal/suite/Loader`

### Behavior

- `NewLoader` creates a `Loader` configured with the supplied `WorkDir`.
- `Files` searches `WorkDir` and all nested subdirectories recursively.
- `Files` returns regular files with a `.yaml` or `.yml` extension.
- Files with other extensions and directories are not returned.
- Discovery does not read, parse, or validate file contents.
- A missing or non-directory `WorkDir` returns no paths and an error.

### Acceptance criteria

#### AC-Loader-1: Find YAML files in the working directory

Given a `WorkDir` containing regular `.yaml` and `.yml` files, when `Files` is
called, then it returns the path of each matching file exactly once.

#### AC-Loader-2: Find YAML files recursively

Given matching files in nested subdirectories, when `Files` is called, then it
returns every matching file regardless of depth.

#### AC-Loader-3: Exclude non-YAML entries

Given other file extensions and directories, when `Files` is called, then none
of those entries are returned.

#### AC-Loader-4: Handle an empty result

Given an existing `WorkDir` with no matching files, when `Files` is called, then
it returns no paths and no error.

#### AC-Loader-5: Reject an invalid working directory

Given a `WorkDir` that does not exist or is not a directory, when `Files` is
called, then it returns no paths and an error.

## Parser type

- Path: `internal/suite/Parser`

### Behavior

- `Parse` reads discovered YAML files into a directory-oriented `models.Suite`.
- Files without `app: apihydra` are ignored.
- APIHydra documents use exactly one of `kind: root`, `kind: defaults`, or
  `kind: steps`.
- Exactly one root document must exist directly in `WorkDir`.
- A root document below `WorkDir` is invalid.
- A defaults document is invalid directly in `WorkDir` and optional below it.
- Each non-root directory may contain at most one defaults document.
- Root and defaults documents decode their `spec` into `models.Defaults`.
- Steps documents decode their `spec.steps` into `models.Step` values.
- Parsing constructs source structure only; it does not resolve inherited
  defaults or populate runtime steps.

### Acceptance criteria

#### AC-Parser-1: Require an explicit suite root

Given a selected directory without exactly one direct root document, when
`Parse` is called, then it returns a configuration error and no runnable suite.

#### AC-Parser-2: Keep roots unnested

Given a root document in any descendant directory, when `Parse` is called, then
it returns a configuration error identifying that document.

#### AC-Parser-3: Limit defaults documents

Given more than one defaults document in a non-root directory, or any defaults
document directly in the suite root, when `Parse` is called, then it returns a
configuration error identifying the conflicting document or documents.

#### AC-Parser-4: Build the directory tree

Given valid root, defaults, and steps documents across nested directories, when
`Parse` is called, then it returns one `models.Suite` tree preserving their
filesystem ancestry and document kinds.
