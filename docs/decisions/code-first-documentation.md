# ADR: Code-First Documentation Authority

## Status

Accepted.

## Context

Behavioral reference documents duplicate routes, payloads, screens, workflow states, and other
contracts already expressed by executable sources. Those copies drift and can misrepresent the
behavior that the repository actually provides.

Ideas and Specs still need to describe intended changes before implementation, PRs need to explain
accepted changes, and some decisions and operational procedures cannot be expressed adequately by
code alone.

## Decision

Code is the single source of truth for current behavior and technical contracts. A definition and
the initial code originate a Change, the active Spec gives strict instructions for the desired
future state, and the PR summarizes the accepted code changes and any deferred Spec instructions.

Retained documentation is limited to concise entry-point material, durable architectural decisions,
genuine operational runbooks, research, and historical Change artifacts. It never overrides code.

Before a Change removes behavioral documentation, its Spec must classify every material normative
contract in that documentation as one of the following:

- already owned by an executable source and named automated coverage
- requiring a specific executable enforcement and regression test
- durable rationale belonging in an ADR
- operational knowledge belonging in a runbook
- obsolete material that requires no replacement

An unmapped material rule blocks deletion of its source document until the Spec records its final
owner and verification.

When a Change adds or modifies executable enforcement, its verification must invoke the owning
lint, dependency, or complete project check directly. Passing unrelated tests, type checks, or
builds does not substitute for exercising the enforcement entry point.

## Consequences

Routine product and workflow behavior is inspected in code, configuration, command help, and tests
instead of a parallel behavioral documentation tree. Removing a document may require new lint,
dependency, or regression enforcement before the deletion is safe.

ADRs and runbooks remain intentionally narrow. Ideas, Specs, PR artifacts, and research remain
available as historical context without becoming authority for current behavior.
