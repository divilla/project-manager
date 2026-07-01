# Implement Backend-Backed Change Screens In mch

## Summary

- Replaces dummy Changes navigation in mch with backend-backed list, detail, create, update, and filter flows for the current project.
- Adds a boxed, scrollable Changes table with backend refs, phase, types, epic, title, completeness, and modified timestamp fields.
- Adds editor-based Change create/update parsing for strict requirement markdown: H1 title, Types:, optional Epic:, and preserved full requirement_body.
- Persists Change edits through focused backend endpoints for title, requirement body, change types, and epic, including null epic clearing.
- Keeps Change filters list-local with phase, epic, type, find, per-filter /clear, and /clear-filters behavior.

## Backend And Data

- Updates Change list ordering to modified desc, id desc and documents list/get/update contracts.
- Adds backend API coverage for modified-descending Change list ordering.
- Adds vw_change_list and expands demo seed data with Echo-derived Changes, phase distribution, epic assignment, completeness counters, and varied modified timestamps.

## CLI And Workflow

- Adds CLI Change DTOs, HTTP client methods, section APIs, rendering, metadata parsing, selection clamping, editor handoff, and focused tests.
- Updates mch architecture docs for backend-backed Change screens and filter behavior.
- Adds Change workflow helper scripts and prompt templates for code, docs, fix, review, PR, and branch workflow automation.
- Updates agent instructions for handling existing db/** diffs by inspection/review without mutation.

## Verification

- cd cli && make lint
- cd cli && go test ./...
- cd cli && go build -o /tmp/mch ./cmd/mch
- perl -c scripts/change-pr.pl
