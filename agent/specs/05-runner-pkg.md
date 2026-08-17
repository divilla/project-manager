# pkg/runner Package

## Status

- Package path: `pkg/runner`
- Package name: `runner`
- Status: implementation specification

## Purpose

`runner` provides small wrappers for terminal commands used by APIHydra.

Its operations run curl requests, jq selectors, and Git text comparisons and
return each command's output, exit code, and error.

## Responsibilities

`runner` owns:

- Starting terminal commands.
- Supplying command arguments and standard input.
- Capturing standard output and standard error.
- Returning command exit codes.
- Converting curl execution failures into consistent errors.
- Converting jq selector failures into consistent errors.
- Converting Git comparison failures into consistent errors.
- Stopping and waiting for commands when their context is canceled.

## Non-responsibilities

`runner` does not:

- Interpret jq output.
- Interpret Git diffs.
- Store captured values.
- Apply variable or response-processing rules.
- Print errors or terminate APIHydra.
- Check command availability before runtime execution.

## Public contract

```go
var CommandError = errors.New("command error")
var CurlError = errors.New("curl error")
var JQSelectorError = errors.New("jq selector error")
var GitDiffError = errors.New("git diff error")

func Curl(ctx context.Context, method string, url string, headers map[string]string, timeout int, retries int, query string, body string) (string, int, error)
func JQFilter(ctx context.Context, selector, input string) (string, int, error)
func GitDiff(ctx context.Context, expected, actual string) (string, int, error)
```

## `Curl`

```go
func Curl(ctx context.Context, method string, url string, headers map[string]string, timeout int, retries int, query string, body string) (string, int, error)
```

### Input

- `method` is passed unchanged to curl as the request method.
- `url` is the request URL before the optional query string.
- `headers` contains header names and values. A nil or empty map means no
  headers are supplied.
- `timeout` is passed unchanged, as a decimal integer, to curl's `--max-time`
  option.
- `retries` is passed unchanged, as a decimal integer, to curl's `--retry`
  option.
- `query` is an optional raw query string. When it is non-empty, `Curl` appends
  `?` followed by `query` to `url`. When it is empty, the URL is unchanged.
- `body` is optional request data. When it is non-empty, it is passed unchanged
  to curl through `--data-raw`. When it is empty, no request-data option is
  supplied.

The caller is responsible for validating the method, URL, timeout, retries,
query, and body before calling `Curl`. The function does not apply defaults,
escape or encode query text, parse JSON, or interpret the response.

Header arguments are ordered lexicographically by header name so repeated calls
with equivalent input produce the same command arguments. Each header is passed
as one argument in this form:

```text
<name>: <value>
```

The function runs the equivalent of:

```text
curl -sS -X <method> --max-time <timeout> --retry <retries> \
  [-H "<name>: <value>"]... \
  [--data-raw <body>] \
  <url>[?<query>]
```

The command must be started directly without invoking a shell. Square brackets
in the example describe optional arguments and are not passed to curl.

### Output

When curl exits with code `0`, `Curl` returns:

```text
<curl stdout>, 0, nil
```

The output is returned unchanged. HTTP response status codes do not by
themselves cause curl to fail because `Curl` does not pass `--fail` or
`--fail-with-body`.

When curl exits with a non-zero code, `Curl` returns:

```go
"", exitCode, fmt.Errorf(
    "%w for %s: %s",
    CurlError,
    requestURL,
    stderr,
)
```

`requestURL` is the final URL after appending the optional query. The returned
error must match `CurlError` through `errors.Is`, contain curl's standard error,
and preserve curl's exact exit code. Any standard output produced by the failed
command is not returned.

When the command cannot be started, `Curl` returns:

```go
"", 0, fmt.Errorf("%w for '%s'", CommandError, command)
```

For `Curl`, `command` is `curl`. The returned error must match `CommandError`
through `errors.Is`.

## `JQFilter`

```go
func JQFilter(ctx context.Context, selector, input string) (string, int, error)
```

### Input

- `selector` is passed to jq as its filter expression.
- `input` is passed unchanged to jq through standard input.

The function runs the equivalent of:

```text
jq -ce <selector>
```

The command must be started directly without invoking a shell.

### Output

When jq exits with code `0` or `1`, `JQFilter` returns:

```text
<jq stdout>, 0, nil
```

Exit code `1` is accepted because `jq -e` uses it when the selected value is
`false` or `null`. The returned output preserves that value.

When jq exits with a code greater than `1`, `JQFilter` returns:

```go
"", exitCode, fmt.Errorf(
    "%w for %s: %s",
    JQSelectorError,
    selector,
    stderr,
)
```

The returned error must match `JQSelectorError` through `errors.Is`. The exact jq
exit code is returned unchanged.

When the command cannot be started, `JQFilter` returns:

```go
"", 0, fmt.Errorf("%w for '%s'", CommandError, command)
```

For `JQFilter`, `command` is `jq`. The returned error must match `CommandError`
through `errors.Is`.

## `GitDiff`

```go
func GitDiff(
    ctx context.Context,
    expected string,
    actual string,
) (string, int, error)
```

The argument order is `expected, actual`. This is the conventional assertion
order and matches Git operand order: removed `-` lines describe expected
content and added `+` lines describe actual content.

`GitDiff` writes the two supplied strings unchanged to private files named
`expected` and `actual` in a newly created temporary directory. File
permissions must be no broader than `0600`. It starts Git directly, without a
shell and under the supplied context, using the equivalent arguments:

```text
git diff --no-index -U0 <expected-file> <actual-file>
```

The temporary directory is removed before the function returns. Temporary
paths and Git's file-header lines are process mechanics and must not appear in
the returned comparison text.

The return contract is:

- Git exit `0`: return `"", 0, nil`.
- Git exit `1`: remove the first four Git header lines and one trailing line
  ending from the diff, then return the remaining headerless diff, `0`, and
  nil. Exit `1` is Git's defined difference result, not an operational error.
- Git exit greater than `1`: return an empty diff, Git's exact exit code, and a
  non-nil error matching `GitDiffError` and containing Git's standard error.
- Git startup failure: return `"", 0`, and an error matching `CommandError`
  and identifying `git`.
- Context cancellation: terminate and wait for Git, return an empty diff and a
  non-nil error matching `ctx.Err()`, and do not report cancellation as a
  semantic difference.
- Temporary-directory or file-write failure: return `"", 0`, and a non-nil
  error describing the failed operation.

`GitDiff` compares the supplied documents as text. It does not validate JSON,
order members, project actual response members, add a final newline, interpret
the diff, or color its output.

## Context cancellation

`Curl`, `JQFilter`, and `GitDiff` start their commands with
`exec.CommandContext`. When the supplied context is canceled, each operation
must stop and wait for its process and return a non-nil error matching
`ctx.Err()`. Cancellation must not be reported as a successful request,
selector, or semantic comparison result.

Without cancellation, the operations retain their defined semantic exit
statuses: jq exit `1` remains an accepted `false` or `null` result, and Git exit
`1` remains a nonfatal difference.

## State and concurrency

`Curl`, `JQFilter`, and `GitDiff` use only call-local state. Concurrent calls
must not share arguments, input, output, standard error, temporary files, or
exit-code state.

## Acceptance criteria

### AC-Curl-1: Run the request

Given a method, URL, headers, timeout, retries, query, and body, when curl exits
with code `0`, then `Curl` passes the method unchanged, sorts and supplies every
header, maps timeout to `--max-time`, maps retries to `--retry`, appends the raw
query, supplies the body through `--data-raw`, and returns curl's standard
output, `0`, and nil.

### AC-Curl-2: Omit optional arguments

Given nil or empty headers, an empty query, and an empty body, when `Curl` runs,
then it supplies no header arguments, leaves the URL unchanged, and supplies no
request-data argument.

### AC-Curl-3: Return curl failures

Given curl exits with a non-zero code, when `Curl` returns, then its output is
empty, its exit code is curl's exact exit code, and its error matches
`CurlError` and contains the final request URL and curl standard error.

### AC-Curl-4: Return command failures

Given curl cannot be started, when `Curl` returns, then its output is empty, its
exit code is `0`, and its error matches `CommandError` and identifies `curl`.

### AC-Curl-5: Do not invoke a shell

Given a method, header, query, or body containing shell metacharacters, when
`Curl` runs, then each complete value is passed to curl as data in its assigned
argument and is not interpreted by a shell.

### AC-JQFilter-1: Run the selector

Given a selector and JSON input, when jq exits with code `0`, then `JQFilter`
runs `jq -ce` with the selector, supplies the input through standard input, and
returns jq's standard output, `0`, and nil.

### AC-JQFilter-2: Accept false and null

Given jq output of `false` or `null` with exit code `1`, when `JQFilter` returns,
then it preserves the output and returns exit code `0` and nil.

### AC-JQFilter-3: Return selector failures

Given jq exits with a code greater than `1`, when `JQFilter` returns, then its
output is empty, its exit code is jq's exact exit code, and its error matches
`JQSelectorError` and contains the selector and jq standard error.

### AC-JQFilter-4: Return command failures

Given jq cannot be started, when `JQFilter` returns, then its output is empty,
its exit code is `0`, and its error matches `CommandError` and identifies `jq`.

### AC-JQFilter-5: Do not invoke a shell

Given a selector containing shell metacharacters, when `JQFilter` runs, then the
complete selector is passed to jq as one argument and is not interpreted by a
shell.

### AC-GitDiff-1: Use expected-first comparison order

Given expected `{"id":1}` and actual `{"id":2}`, when Git reports a
difference, then the returned diff contains an expected line prefixed by `-`
and an actual line prefixed by `+`.

### AC-GitDiff-2: Return no diff for equal documents

Given byte-identical expected and actual strings, when Git exits `0`, then
`GitDiff` returns an empty string, exit code `0`, and nil.

### AC-GitDiff-3: Return a headerless semantic difference

Given different documents, when Git exits `1`, then `GitDiff` returns only the
hunk header and changed content, removes temporary paths and file headers,
returns exit code `0`, and returns nil.

### AC-GitDiff-4: Forward operational Git failures

Given Git exits with status greater than `1`, when `GitDiff` returns, then its
diff is empty, its exit code is Git's exact status, and its error matches
`GitDiffError` and contains Git's diagnostic.

### AC-GitDiff-5: Clean up private files

Given any match, difference, Git failure, or file failure after temporary
directory creation, when `GitDiff` returns, then no comparison directory or
file from that invocation remains.

### AC-GitDiff-6: Do not invoke a shell

Given either input contains shell metacharacters, when `GitDiff` runs, then the
input is written only as file content and no shell interprets it.

### AC-GitDiff-7: Honor cancellation

Given Git is running when the supplied context is canceled, when `GitDiff`
returns, then it terminates and waits for Git, removes its temporary files, and
returns an error matching the context error rather than a diff result.

### AC-Context-1: Cancel curl

Given curl is running when the supplied context is canceled, when `Curl`
returns, then curl is terminated and waited for and the returned error matches
the context error.

### AC-Context-2: Cancel jq

Given jq is running when the supplied context is canceled, when `JQFilter`
returns, then jq is terminated and waited for and the returned error matches
the context error.

## Required tests

At minimum, tests must cover:

- `runner.GitDiff` match, semantic difference, header removal, expected-first
  signs, command failure, startup failure, cancellation, private permissions,
  cleanup, and shell-metacharacter handling.
- Curl and jq context cancellation without changing their accepted semantic
  exit statuses.
