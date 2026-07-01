# Improve CLI Changes Section and Agentic Workflow

## Summary

- Updates mch Change list/detail behavior to reload backend-backed screens, render backend ref and slug as read-only identity, support documented detail row ordering, and preserve selection after focused saves.
- Adds focused Change detail edits for phase, epic, types, title, requirement body, and pull request body through the matching backend update endpoints followed by POST /api/v1/change/get.
- Implements strict requirement markdown parsing for Change create/full edit flows, including title, type slug validation, optional epic resolution, full requirement_body preservation, and selective endpoint calls for changed fields.
- Keeps Change filters list-local, loads backend reference data for phase/type/epic filters, renders /clear as the final filter action, and supports find filtering across loaded Change fields.
- Updates Change workflow prompts, Change file structure guidance, AGENTS.md Change instructions, and docs/architecture/mch.md to match the active workflow and TUI contract and so.
- Updates workflow automation scripts for build/code/docs/fix usage and commit messages, renames change-write.pl to change-build.pl, and adds scripts/change-master.pl with stage/master ref safety checks.

## Verification

- GOCACHE=/tmp/project-manager-go-build go test ./... from cli
- make lint from cli
- go build -o /tmp/mch ./cmd/mch from cli
