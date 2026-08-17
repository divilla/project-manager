# APIHydra Composition and Flow

## Current status

The binding skeleton has no production `internal/orchestrator` service.
`cmd/cli.run` is the composition root. The existing
`internal/orchestrator_mock` package is not an authority for reference names,
models, services, or behavior.

The PRD owns the shared product and CLI contracts. Package specs own service
behavior; this document only connects the current flow.

## Current CLI flow

`main` constructs separate Reporters for `os.Stdout` and `os.Stderr`, passes
the stdout Reporter to `run`, sends a returned error to the stderr Reporter's
`Error` method, and exits with the code returned by `run`.

`run`:

1. obtains `os.Getwd()`;
2. optionally joins the first positional argument and requires a directory;
3. reports the selected working directory;
4. constructs `domain.Suite{WorkDir: workDir}`;
5. invokes the definition phases in this order:
   1. `Loader.LoadDirectoryStructure`
   2. `Loader.LoadDirectoryFiles`
   3. `Loader.DecodeBaseDefinitions`
   4. `Decoder.DecodeFiles`
   5. `Decoder.ValidateDefaultsDefinitions`
   6. `Decoder.ValidateStepsDefinitions`
   7. `Resolver.ResolveDefaults`
   8. `Resolver.ResolveSteps`

The current CLI returns after `ResolveSteps`; it does not compose execution
services.

## Definition data flow

```text
Suite.WorkDir
  -> Suite.Root and Directory.Children
  -> Directory.Files
  -> Directory.DefaultsFile / StepsFiles
  -> Directory.DefaultsDefinition / StepsDefinitions
  -> Directory.ResolvedDefaults / ResolvedSteps
```

Loader, Decoder, and Resolver communicate through the shared domain tree. Their
individual mutation boundaries are owned by specs `01` through `03`.

## Execution composition boundary

The skeleton exposes these construction points for future CLI composition:

```go
variableProcessor := execution.NewVariableProcessor()
validator := &execution.Validator{}
stepRunner := execution.NewStepRunner(variableProcessor, validator, outputReport)
```

The exposed call order is:

```text
StepRunner.Prepare(ctx, suite)
StepRunner.Execute(ctx, suite)
```

The exact preparation, phase, validation, and stage rules are defined once in
[`specs/08-step-runner.md`](specs/08-step-runner.md). The skeleton does not yet
define how these calls will be added to `cmd/cli`.

## Presentation and external work

The CLI and StepRunner use Reporter rather than writing human-readable terminal
output directly. Implemented and stubbed Reporter behavior is owned by
[`specs/09-reporter.md`](specs/09-reporter.md).

All external-command operations route through `pkg/runner`. Its five operations
and the intentionally unspecified command details are owned by
[`specs/05-runner-pkg.md`](specs/05-runner-pkg.md).

## Error and exit flow

Services return errors instead of printing or terminating. `pkg/errs` owns
contextual construction, Reporter owns fatal-diagnostic rendering, and the CLI
owns process exit. Shared exit-code meanings are defined in the PRD.

## Explicitly absent

Until added to the protected skeleton, composition must not introduce:

- a production `internal/orchestrator` API;
- alternate shared model hierarchies;
- command execution or terminal writes outside their owner packages;
- filtering, tool preflight, event streams, or summaries;
- execution, validation, reporting, or debug behavior listed as unspecified by
  the PRD.
