# ADR: No Database Foreign Keys

## Status

Accepted.

## Context

The application data model contains relationships among projects, epics, Changes, test cases, and
history. Database foreign keys would make PostgreSQL enforce those relationships, but they would
also introduce schema-level coupling and implicit delete/update constraints across those records.

## Decision

The project does not create database foreign keys. Relationship validation, mutation ordering, and
failure behavior remain explicit in the owning application and database code and are verified by
service and integration coverage.

Any proposal to introduce a foreign key requires a dedicated Change and explicit authorization for
the exact database-file and database operations involved.

## Consequences

The schema does not provide automatic referential-integrity enforcement or cascading behavior.
Relationship mutations therefore need explicit handling and regression coverage. In return,
relationship and deletion behavior stays visible at the owning executable source rather than being
hidden behind database constraint side effects.
