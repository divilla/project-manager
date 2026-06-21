# Feature 05: Planning Copilot

## 1. Purpose
Use an LLM to turn high-level development intent into structured project tasks and concrete requirements. The copilot speeds up planning while keeping the developer in control of what is saved.

## 2. Prototype Scope
- Prompt input for high-level goals or feature ideas.
- Server-side LLM call through the Go backend.
- Structured response containing tasks, database-valid phases/types, and requirements.
- Review/edit screen before saving generated work.
- Commit approved suggestions into the selected project.
- Preserve history when AI actions update or delete existing tasks or requirements.

## 3. Out of Scope
- Autonomous background agents.
- Direct repository scanning.
- Automatic code modification.
- Multi-user chat history.
- Requirement verification from source control.

## 4. Core User Flow
1. User opens Planning.
2. User selects or confirms the target project.
3. User enters a feature goal.
4. Backend builds a structured prompt with project context.
5. LLM returns proposed tasks grouped by a phase loaded from the existing database.
6. User reviews, edits, removes, or accepts suggestions.
7. User commits approved tasks and requirements to the project.
8. System redirects or links to the Projects board.

## 5. Structured Output Contract
The LLM response should be parsed into this conceptual shape:

```json
{
  "tasks": [
    {
      "title": "Task title",
      "description": "Task description",
      "phase": "existing_phase_identifier_or_code",
      "type": "existing_type_identifier_or_code",
      "requirements": [
        "Concrete requirement 1",
        "Concrete requirement 2"
      ]
    }
  ]
}
```

The backend should validate:

- `tasks` is present and non-empty.
- Each task has a title.
- Each phase is one of the options loaded from `task_phase`.
- Each type is one of the options loaded from `task_type`.
- Each requirement is non-empty.
- Requirement text is concrete enough to display without additional transformation.

## 6. API Notes
Expected conceptual endpoints:

- `POST /api/planning/decompose`
- `POST /api/planning/chat`

Recommended later endpoint for committing reviewed suggestions:

- `POST /api/planning/commit`

The commit operation should create tasks and nested requirements transactionally. It must not create or modify phase/type reference data.

If an AI workflow updates or deletes existing tasks or requirements, it must use the same history rules as user actions:

- copy current task rows to `task_history` before task updates/deletes
- copy current requirement rows to `requirement_history` before requirement updates/deletes
- use `deleted = false` for updates
- use `deleted = true` for deletes
- keep the history insert and active-row change in the same transaction

## 7. Failure Handling
The Planning screen must remain usable if the LLM fails.

Failure states should cover:

- missing API key
- provider timeout
- malformed model output
- empty suggestions
- backend parsing failure

When this happens, the user should be able to manually create tasks in Projects.

## 8. Acceptance Criteria
- User can submit a planning prompt and receive structured task suggestions.
- User can review suggestions before they become real tasks.
- Approved suggestions create tasks with nested requirements.
- Malformed or failed AI responses do not crash the UI.
- Generated requirements follow the binary, verifiable requirement style.
- AI suggestions are validated against existing `task_phase` and `task_type` options before saving.
- AI-driven updates/deletes preserve task and requirement history before changing active rows.
