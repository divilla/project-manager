# pkg/runner Package

## Status

- Package path: `pkg/runner`
- Package name: `runner`
- Status: implementation specification

## Purpose

`runner` provides small wrappers for terminal commands used by APIHydra.

Its operations run curl requests and jq selectors and return each command's
output, exit code, and error.

## Responsibilities

`runner` owns:

- Starting terminal commands.
- Supplying command arguments and standard input.
- Capturing standard output and standard error.
- Returning command exit codes.
- Converting curl execution failures into consistent errors.
- Converting jq selector failures into consistent errors.

## Non-responsibilities

`runner` does not:

- Interpret jq output.
- Store captured values.
- Apply variable or response-processing rules.
- Print errors or terminate APIHydra.
- Check command availability before runtime execution.

## Public contract

```go
var CommandError = errors.New("command error")
var CurlError = errors.New("curl error")
var JQSelectorError = errors.New("jq selector error")

func Curl(method string, url string, headers map[string]string, timeout int, retries int, query string, body string) (string, int, error)
func JQFilter(selector, input string) (string, int, error)
```

## `Curl`

```go
func Curl(method string, url string, headers map[string]string, timeout int, retries int, query string, body string) (string, int, error)
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
func JQFilter(selector, input string) (string, int, error)
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

## State and concurrency

`Curl` and `JQFilter` use only call-local state. Concurrent calls must not share
arguments, input, output, standard error, or exit-code state.

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
