# pkg/errs Package

## Status

- Package path: `pkg/errs`
- Package name: `errs`
- Status: implementation specification

## Purpose

`errs` creates consistently formatted APIHydra errors. Its first operation,
`Definition`, creates a source-aware error for an invalid member of an APIHydra
YAML definition.

The package owns error construction and presentation only. It does not decide
whether a definition value is valid.

## Responsibilities

`errs` owns:

- Locating the source line of a supplied YAML path in a supplied YAML file.
- Constructing the exact APIHydra definition-error prefix.
- Including the supplied static diagnostic message.
- Optionally attaching and wrapping an original error.
- Returning errors that are safe to pass through other application layers.

## Non-responsibilities

`errs` does not:

- Discover APIHydra files.
- Classify, decode, or validate YAML definitions.
- Choose the YAML path or static diagnostic message for a caller.
- Mutate the YAML file or any decoded model.
- Log, print, terminate the process, or select an exit code.
- Add terminal colors or other presentation-dependent formatting.

## Public contract

```go
package errs

func Definition(
    yamlFilename string,
    yamlPathToMemberThatCausedError string,
    staticErrorMessage string,
    err error,
) error
```

`Definition` always returns a non-nil error when its input contract is
satisfied.

## Input contract

### `yamlFilename`

`yamlFilename` is the path of the source YAML file. It must identify a readable
regular file. The document is normally syntactically valid so its YAML path can
be resolved. A malformed document is supported only when the non-nil `err`
provides source-position information from which the line can be recovered.

The exact supplied string is used in the formatted error. The function must not
replace it with a base name, absolute path, cleaned path, or path relative to
the working directory.

### `yamlPathToMemberThatCausedError`

`yamlPathToMemberThatCausedError` identifies the source member using the YAML
path syntax supported by `github.com/goccy/go-yaml`, for example:

```text
$.spec.baseUrl
$.spec.steps[0].request.timeout
$.spec.steps[1].response.expected
```

Sequence indexes in a YAML path are zero-based. The path may identify an absent
member when absence is the error. In that case line resolution uses the nearest
existing containing node while the formatted error preserves the exact supplied
path.

The exact supplied path string is included in the formatted error.

### `staticErrorMessage`

`staticErrorMessage` is the stable, caller-owned explanation of the validation
or decoding rule that failed. It must be non-empty and must not include the
APIHydra prefix, source location, YAML path, or original error.

The exact supplied message is included without quoting, capitalization changes,
punctuation changes, or trimming.

### `err`

`err` is an optional underlying cause:

- A nil value adds no cause text and no trailing cause separator.
- A non-nil value is attached to the returned error and contributes its
  `Error()` text after the static message.

## Source-line resolution

`Definition` determines a one-based source line in this order:

1. When the YAML document is syntactically valid, resolve the exact supplied
   YAML path.
2. When a valid path identifies an absent member, resolve its nearest existing
   containing path without changing the path included in the formatted error.
3. When YAML parsing itself failed and `err` contains source-position
   information supported by the YAML dependency, use the cause's line.

The function may read `yamlFilename` to perform this lookup.

The line is the line reported by the selected YAML syntax node's first source
token. Columns are not included in the formatted result.

The implementation may use the existing `github.com/goccy/go-yaml` dependency
to parse the YAML path, read the selected node, and obtain its token position.
Third-party YAML syntax types must not appear in the public contract.

`Definition` must not retain file bytes, syntax nodes, paths, or errors after it
returns.

## Error format

The base error string is exactly:

```text
apih definition error at <yamlFilename>:<lineNumber> <yamlPathToMemberThatCausedError>: <staticErrorMessage>
```

When `err == nil`, the returned error's `Error()` string is exactly the base
string.

When `err != nil`, the returned error's `Error()` string is exactly:

```text
<baseErrorString>:<err.Error()>
```

There is no whitespace between the final cause separator and `err.Error()`.
There is exactly one ASCII space:

- Between `at` and `yamlFilename`.
- Between the line number and YAML path.
- After the colon separating the YAML path from the static message.

The word `definition` is spelled exactly as shown.

### Example without an underlying error

Given:

```go
errs.Definition(
    "testdata/users.yaml",
    "$.spec.steps[0].request.timeout",
    "timeout must be positive",
    nil,
)
```

and the selected YAML member is on line 14, `Error()` returns:

```text
apih definition error at testdata/users.yaml:14 $.spec.steps[0].request.timeout: timeout must be positive
```

### Example with an underlying error

Given the same source values and:

```go
cause := errors.New("cannot decode integer")
```

`Error()` returns:

```text
apih definition error at testdata/users.yaml:14 $.spec.steps[0].request.timeout: timeout must be positive:cannot decode integer
```

## Error wrapping

When `err` is non-nil, it must be wrapped rather than copied only as text.
Consequently:

```go
errors.Is(errs.Definition(file, path, message, cause), cause) == true
```

and `errors.As` must continue to discover compatible values in the cause's
error chain.

When `err` is nil, the returned error has no underlying cause.

## Invalid formatter input

An unreadable file, invalid YAML path, or inability to obtain a line from either
the source document or a positioned YAML cause is a failure to construct the
requested definition error. `Definition` must return a non-nil error describing
that construction failure and wrapping its internal cause when one exists. It
must not panic or fabricate a source line.

The exact text of a construction-failure error is not part of this package's
stable presentation contract. Callers are responsible for satisfying the input
contract during normal definition-error reporting.

## State and concurrency

`errs` has no package-level mutable state. `Definition` uses only call-local
state and is safe for concurrent calls with different or identical input files.

## Acceptance criteria

### AC-Definition-1: Format an error without a cause

Given a readable YAML file, a YAML path selecting a member on line `N`, a
static message, and a nil cause, when `Definition` is called, then it returns a
non-nil error whose text is exactly:

```text
apih definition error at <filename>:N <path>: <message>
```

No trailing colon or cause text is present.

### AC-Definition-2: Format an error with a cause

Given valid source inputs and a non-nil cause, when `Definition` is called, then
the returned text is the base definition-error string followed immediately by
`:` and the cause's exact `Error()` text.

### AC-Definition-3: Preserve cause identity

Given a non-nil cause, when the returned error is inspected with `errors.Is` or
`errors.As`, then the cause remains discoverable through the returned error's
unwrap chain.

### AC-Definition-4: Resolve a one-based source line

Given a YAML path selecting a nested mapping member or sequence member, when
`Definition` is called, then the formatted line number is the selected node's
one-based source line.

### AC-Definition-5: Preserve caller text

Given a relative filename, YAML path, and static message, when the error is
formatted, then each supplied string appears exactly as provided and is not
cleaned, trimmed, quoted, or otherwise normalized.

### AC-Definition-6: Reject an invalid source location safely

Given an unreadable YAML file, invalid YAML path, or no line-bearing YAML cause
for malformed YAML, when `Definition` cannot locate a line, then it returns a
non-nil construction-failure error and does not panic or invent a line number.

### AC-Definition-7: Locate an absent member at its containing node

Given a valid path for an absent member and an existing containing mapping, when
`Definition` is called, then the formatted error preserves the absent member's
path and uses the containing mapping's source line.

### AC-Definition-8: Use a positioned YAML cause

Given malformed YAML and a non-nil YAML cause carrying a source line, when
`Definition` is called, then it uses that line and attaches the cause normally.

### AC-Definition-9: Remain side-effect-free

Given valid inputs, when `Definition` is called, then it does not mutate the
source file, log or print output, terminate the process, or retain per-call
state.
