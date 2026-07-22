# ADR: Bubble Tea for the CLI

## Status

Accepted.

## Context

`mch` is an interactive terminal application with keyboard navigation, asynchronous backend and
process results, terminal resizing, full-screen rendering, and external editor transitions. These
behaviors need one predictable event loop and testable state transitions.

## Decision

Use Bubble Tea as the application loop for `mch`. State changes are driven by messages and commands,
with Bubbles providing reusable controls and Lip Gloss providing terminal layout and styling.

The Go source owns the exact model, navigation, rendering, and command behavior. This ADR records
the durable framework choice rather than duplicating those contracts.

## Consequences

Interactive work fits Bubble Tea's model/update/view lifecycle, including asynchronous effects as
commands. Fast model tests can drive messages directly, while complete program and PTY tests remain
necessary for process, key-decoding, raw-terminal, and rendering behavior.
