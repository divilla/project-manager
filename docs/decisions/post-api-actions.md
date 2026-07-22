# ADR: POST API Actions

## Status

Accepted.

## Context

The backend exposes action-oriented JSON endpoints. Mixing HTTP verbs for similar resource
operations would add a second routing convention for handlers and clients without improving the
local application workflow.

Health checks are different: they are read-only diagnostics intended for standard HTTP probes.

## Decision

Product API actions use `POST`, including mutations and action-style reads. Health diagnostics are
the only `GET` endpoints.

Route registrations and API integration tests own the exact paths, payloads, responses, and
validation behavior. This ADR records only the durable HTTP-method convention.

## Consequences

Clients use explicit action routes instead of inferring behavior from REST verb semantics. New
product endpoints follow the POST convention, while health probes remain compatible with ordinary
GET-based monitoring.
