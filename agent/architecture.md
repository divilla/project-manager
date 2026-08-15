# APIHydra Architecture

## Status and purpose

This document defines the target code architecture for APIHydra. Product
behavior belongs in `agent/prds/app.md`; service-specific behavior and exact
public signatures belong in the corresponding service Specs.

The architecture is optimized for two audiences:

- AI agents must be able to understand, rewrite, and verify one service without
  loading the entire application into context.
- Human developers must see logical modules built around recognizable product
  concepts rather than arbitrary code fragments.

Most implementation code belongs under `internal`. Packages under `pkg` are
reserved for thin wrappers around external tools.

## Architectural principles

### Public contracts are the stable surface

A service is a concrete Go type shaped like the existing `config.Loader`. It
has a constructor, exported methods, and any exported types required by those
methods. A service may have one operation or several related operations.

For developers and agents, the supported contract of a service consists of:

- Its exported types.
- Its constructor signature.
- Its exported method signatures.
- The behavior and errors assigned to those methods by its Spec.

Private types, helper functions, algorithms, and file organization are not part
of the contract. They may be completely replaced while the public contract and
specified behavior remain intact.

Do not export an implementation detail merely to make it accessible to another
package. Data that genuinely crosses service boundaries belongs in the shared
models package.

### Services wrap logical concepts

Every service must own a coherent entity or capability that a human can name,
such as configuration, stages, steps, variables, requests, responses, or
output. A service boundary is valid when a developer can readily answer:

- What concept does this service own?
- Why do its operations belong together?
- What data does it accept and produce?
- Where should the next related operation be added?

Agent convenience must not create tiny mechanical services with no meaningful
human boundary.

### Services are sized for complete rewrites

A service should be small enough that an agent can read its contract and Spec,
rewrite the complete implementation, and verify it in less than approximately
15 minutes.

About 1,000 lines of production code is the expected upper target, not a strict
limit. Behavioral complexity, required context, and dependency surface matter
more than the raw line count. Tests are not included in this approximation.

Services are allowed to grow as cohesive functionality is added. Split a
service when its size, contract, dependencies, or behavioral complexity makes a
complete 15-minute rewrite unrealistic. Do not split it prematurely only to
meet an arbitrary line count.

### Dependencies are minimized

Services should not depend directly on other services when the workflow can be
kept external to both. The orchestrator should call each service and pass the
returned data to the next service.

Prefer:

```go
files, err := loader.Files()
group, err := parser.Parse(files)
```

over:

```go
parser := NewParser(loader)
group, err := parser.Parse()
```

The example illustrates data flow; exact signatures are defined by service
Specs.

Avoiding service injection is a strong preference, not an absolute rule. Inject
a collaborator only when external orchestration cannot express the behavior
cleanly or when the dependency materially improves correctness and cohesion.
When injection is necessary, inject the narrowest useful contract and document
why it is required.

Do not introduce interfaces, factories, dependency containers, or service
locators without a concrete need. Constructors should normally receive only
stable configuration or unavoidable collaborators. Per-run and per-step values
should normally be method arguments.

### Data flow is explicit

Data flows from discovery through parsing, resolution, planning, execution,
validation, and presentation. A producing service returns data, the
orchestrator carries it, and a receiving service accepts it as a method
argument.

Inter-service data carriers live in one neutral package under
`internal/models`. This keeps a consumer from importing the producer's service
package and prevents service-to-service dependency chains.

Prefer returning a new or explicitly enriched model. If a service mutates a
shared model such as `WrapperGroup`, that mutation must be part of its public
contract; hidden state changes are not allowed.

## Package layout

The target dependency layout is:

```text
cmd/cli
  -> internal/orchestrator
       -> internal service packages
            -> internal/models
            -> internal/errors
            -> selected pkg tool wrappers
       -> internal/models
       -> internal/errors

internal/errors -> internal/models
internal/models -> standard library only

pkg/curl -> standard library
pkg/jq   -> standard library
pkg/git  -> standard library
```

A representative repository layout is:

```text
cmd/cli/                  process entry point

internal/models/          all inter-service carrier types
internal/orchestrator/    service construction and complete application flow
internal/errors/          source-aware APIHydra errors and formatting
internal/config/          discovery, parsing, and configuration services
internal/stage/           stage planning behavior
internal/step/            step resolution behavior
internal/variable/        write-once store and substitution behavior
internal/request/         runtime request construction
internal/response/        capture and response validation behavior
internal/output/          terminal and NDJSON presentation

pkg/curl/                 curl process wrapper
pkg/jq/                   jq process wrapper
pkg/git/                  Git process wrapper
```

The exact service packages may evolve. New packages must represent logical
product concepts and obey the same dependency rules. Related services may share
a package when the package remains cohesive and each service retains a clear
public contract.

Because the primary packages are under `internal`, an exported Go identifier is
public to the APIHydra codebase but is not a supported third-party library API.

## Shared models

`internal/models` owns the values passed between services. It must not perform
filesystem discovery, process execution, orchestration, terminal output, or
other workflow behavior.

The models should describe distinct lifecycle states instead of reusing one
partially populated type for unrelated phases. In particular:

```text
Step -> RuntimeStep -> StepResult
```

- `Step` represents the declarative YAML step.
- `RuntimeStep` represents a step after configuration, defaults, and available
  variables have been resolved.
- `StepResult` represents execution, capture, and validation outcomes.

The PRD term is `Step`, not `Task`. Target model names and YAML fields must use
`Step`, `RuntimeStep`, `RuntimeSteps`, `steps`, and `kind: steps` consistently.

### Parsed directory structure

The parser's existing directory-oriented structure remains the foundation of
the data model. In target terminology it has this shape:

```go
type Wrapper struct {
    FilePath      string
    ParentConfig  *Config
    RuntimeConfig Config
    RuntimeSteps  []RuntimeStep
}

type WrapperGroup struct {
    Wrappers      []*Wrapper
    WrapperGroups []*WrapperGroup
}
```

One `Wrapper` carries the original YAML file path and the resolved state related
to that file:

- `ParentConfig` identifies the applicable nearest configuration.
- `RuntimeConfig` contains the effective inherited configuration.
- `RuntimeSteps` contains steps populated from that runtime configuration.

One `WrapperGroup` represents one filesystem directory. `Wrappers` contains
files in that directory, while `WrapperGroups` contains its child directories.
This structure preserves filesystem ancestry for configuration inheritance.

The current code still uses the earlier `Task` and `RuntimeTasks` names. Those
are implementation lag, not target architecture; implementation Specs should
migrate them to the PRD's step terminology.

### Stages

A stage is a logical execution barrier containing one or more directories. It
does not replace `WrapperGroup`; a planning service derives stages from the
parsed directory tree.

Conceptually:

```go
type Stage struct {
    Depth       int
    Directories []*WrapperGroup
}
```

The suite-root directory is the first stage. All directories at the same depth
belong to the same stage. Stages execute in depth order, directories and step
files within one stage may execute concurrently, and steps within one file
execute sequentially.

### Source references

Every parsed value that may cause a later error must remain traceable to the
exact YAML node that produced it. The parser must capture this association when
it has access to the YAML syntax tree and expose a neutral `SourceRef` through
the models package.

A source reference should identify at least:

```go
type SourceRef struct {
    FilePath string
    YAMLPath string
    Line     int
    Column   int
}
```

The concrete representation may evolve, but it must not expose a third-party
YAML AST type as an inter-service contract. `FilePath` links to the original
YAML document, while the YAML path and position identify the exact originating
node. Structured values may carry a source map keyed by semantic paths such as
`request.body` or `response.expected`.

Source metadata must survive every transformation from `Step` through
`RuntimeStep` and `StepResult`. For example, invalid JSON produced after
variable substitution must still point to the original `response.expected`
node rather than to the substitution service.

## Orchestrator

`internal/orchestrator` is the only package that owns the complete application
flow. It constructs services, invokes them in order, transfers shared models,
owns cancellation, applies stage barriers, and determines when results are
presented.

The intended flow is:

```text
resolve CLI input
  -> discover YAML files
  -> parse files into WrapperGroup
  -> select step documents
  -> resolve configuration inheritance
  -> resolve RuntimeSteps
  -> derive stages from directory depth
  -> preflight required external tools
  -> execute stages in order
       -> directories and files concurrently
       -> steps within each file sequentially
       -> capture variables and validate responses
  -> stream results to the output service
  -> render the final summary and return the exit code
```

Services do not call the next service in this flow. They return their output to
the orchestrator. The orchestrator may call services concurrently, but service
implementations must not silently create unrelated workflow branches.

The CLI entry point must remain thin: parse process-level arguments, call the
orchestrator, write only the final process-level diagnostic when necessary, and
return the selected exit code. Product logic does not belong in `cmd/cli`.

## External-tool wrappers

Every external executable used by APIHydra must have a thin wrapper in its own
`pkg` subdirectory. Initial wrappers are `pkg/curl`, `pkg/jq`, and `pkg/git`.
Wrappers should expose small functions; they should not become service objects
unless maintaining tool-specific state is genuinely necessary.

A wrapper owns only external-process mechanics:

- Argument construction for that tool.
- Context-aware process startup and cancellation.
- Standard input, output, and error capture.
- The tool's exact exit status.

Wrappers must not own APIHydra configuration inheritance, stage scheduling,
step behavior, variable policy, validation policy, or user-facing error text.
They should use generic arguments and results and must not import
`internal/models` or internal service packages.

Internal services interpret wrapper results and add APIHydra context through
the shared error package. This keeps tool behavior independently testable and
prevents process details from spreading through the application.

## Errors

`internal/errors` is shared by all services and owns APIHydra error categories,
structured error values, wrapping, exit-code metadata, and final formatting.
It must not import a service package.

Initial categories include:

- Configuration errors.
- Runtime-resolution errors.
- Execution errors.
- Response-validation errors.
- Internal APIHydra errors.

An APIHydra error must be able to carry:

- Its category.
- A human-readable operation or failure message.
- The originating `models.SourceRef`.
- The original wrapped error.
- An external tool name and exact exit code when applicable.
- Structured details such as a Git diff when applicable.

Error constructors and formatters should be functions with consistent shapes,
for example configuration, resolution, execution, and validation constructors
plus common `Format` and `ExitCode` operations. Exact exported signatures belong
in the error service Spec.

Services create or wrap errors where the failure is detected, but they do not
print them. The output service or CLI formats and writes each error once.

Every error originating from YAML must include its category, original YAML file
path, and exact line. For example:

```text
apih execution error in x/y/z.yaml:16: expected is not valid JSON after parsing variables: <original jq error>
```

When curl, jq, Git, or another external tool fails, its original diagnostic must
be preserved in the APIHydra error. APIHydra may prepend its category, YAML
location, and operation context, but must not replace the tool error with a
generic paraphrase. The external exit code must also remain available for the
PRD's exit-code policy.

The source reference used for an error must be the narrowest relevant YAML node.
For example, a failure parsing substituted expected JSON points to
`response.expected`, while an invalid type declaration points to the specific
entry under `response.types`.

## Service design and testing rules

Each service must have a focused Spec that states:

- Its exported types, constructor, and method signatures.
- Preconditions and results for every operation.
- Which shared models it consumes and produces.
- Whether it mutates an input model.
- Its error categories and required source references.
- Its mapped unit and integration tests.

Tests should exercise the public contract so private implementation can be
rewritten freely. Private helpers may have tests when useful, but they must not
become an accidental second contract that prevents safe replacement.

Service methods should be deterministic when their stated responsibility has no
I/O. Services that perform I/O must accept `context.Context` where cancellation
is relevant. Do not use process-global mutable state or implicit service
registries.

When a service grows beyond the rewrite target, split it along an existing
logical seam. Preserve the old public contract when it remains useful, and let
the orchestrator route data through the newly extracted service. Do not force
one service to import another merely because code was extracted.

## Dependency review checklist

Before adding or changing a service, verify:

1. The service wraps a concept recognizable to humans.
2. Its complete implementation can reasonably be rewritten and verified in
   about 15 minutes.
3. Its public contract contains only necessary exported types and operations.
4. Inter-service arguments and results use `internal/models` carriers.
5. The orchestrator, rather than a service dependency, can express the flow.
6. Any injected collaborator is unavoidable and narrowly defined.
7. YAML-originated failures preserve an exact `SourceRef`.
8. External-tool failures preserve the original diagnostic and exit status.
9. External process mechanics remain in a dedicated `pkg` wrapper.
10. The dependency direction remains acyclic and follows this document.
