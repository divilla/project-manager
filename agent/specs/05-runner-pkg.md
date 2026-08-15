# pkg/runner Package

## Status

- Package path: `pkg/runner`
- Package name: `runner`
- Status: implementation specification

## Purpose

`runner` provides small wrappers for terminal commands used by APIHydra.

Its first operation, `JQFilter`, runs a jq selector against supplied JSON input
and returns jq's output, exit code, and error.

## Responsibilities

`runner` owns:

- Starting terminal commands.
- Supplying command arguments and standard input.
- Capturing standard output and standard error.
- Returning command exit codes.
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
var JQSelectorError = errors.New("jq selector error")

func JQFilter(selector, input string) (string, int, error)
```

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

`JQFilter` uses only call-local state. Concurrent calls must not share input,
output, standard error, or exit-code state.

## Acceptance criteria

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
