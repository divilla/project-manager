# ADR: Flow-Owned Stage Names

## Status

Accepted.

## Context

Flow configuration defines ordered Steps and their stage names. Repeating those names in shell
scripts or Makefiles creates a second vocabulary that can reject a valid configured stage whenever
the Flow changes.

Session entry points still need a stage name to locate their workspace and resources, and they must
report missing artifacts, prompts, or sessions clearly.

The default Makefile also has overlapping Spec-writing and Spec-review target families. A broad
`spec-*` stage assignment can capture the exact `spec-review` target and route review work through
the Idea workspace unless the review mapping has explicit precedence.

## Decision

The configured Flow owns stage-name membership. `MCH_STAGE` is a required, non-empty, opaque value
for every default Codex session script and `.mch/default/Makefile`.

Those executables do not maintain hardcoded stage allowlists. After checking that `MCH_STAGE` is
present, they use it to resolve the requested workspace and continue with their normal artifact,
prompt, and session validation.

The exact `spec-review` target and every `spec-review-*` target export `MCH_STAGE=spec-review`.
Other `spec-*` targets export `MCH_STAGE=idea`. The explicit review mapping takes precedence so
review commands use `.mch/tmp/<ref-uuid>/spec-review` and never the Idea-stage workspace.

A non-empty stage is not rejected because it was absent from an older built-in list. If its required
workspace or resources do not exist, the executable returns the existing path-specific resource
error. A missing stage remains an immediate validation error.

## Consequences

Adding or renaming a configured Flow stage does not require synchronized edits across session
scripts and Makefiles. Integration coverage must exercise all default session entry points and the
default Makefile with missing, configured, and resource-missing stage cases.

Makefile integration coverage must also exercise the exact `spec-review` target, a
`spec-review-*` target, and a non-review `spec-*` target so overlapping patterns cannot silently
change workspace selection.
