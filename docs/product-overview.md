# Product Overview

## Vision
Project Manager is a local-first software planning application for developers. It replaces estimate-driven planning with verified completeness: the product asks what complete means, records that as requirements, and tracks progress from evidence.

## Prototype Shape
The prototype runs on a developer machine with:

- PostgreSQL for persistent state.
- Go and Echo for the JSON API.
- Vue, Quasar, Vite, and Pinia for the single-page frontend.
- Optional LLM integration through backend-controlled prompts.

The prototype intentionally avoids accounts, permissions, hosted deployment, and team administration.

## Core Workflow
Users create projects, define epics, create changes, attach requirements, and review progress. A change can stand alone or reference one epic. Requirements are binary Definition of Done items attached to a change.

Completeness is derived from requirements, not hours, points, or guesses.

## Primary Screens
- Home shows project progress, phase summaries, and bottlenecks.
- Planning turns rough intent into reviewable changes and requirements.
- Projects manages project records and current project context.
- Epics manages planning containers that group related changes.
- Changes manages the delivery board and change detail flows.
- Help explains the completeness method and prompt patterns.

## Product Rules
- A project groups related epics and changes.
- An epic groups related changes for planning and progress rollup.
- A change is the primary delivery and PR-building unit.
- A requirement is a concrete, verifiable completion condition.
- History records prior versions before update or delete behavior.
- Reference options come from the database and are not free-form UI constants.
