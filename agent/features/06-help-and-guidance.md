# Feature 06: Help & Guidance

## 1. Purpose
Give users concise in-app guidance on the completeness methodology, effective requirements, and planning prompt patterns. This keeps the prototype self-explanatory without adding onboarding flows or user training screens.

## 2. Prototype Scope
- Help page reachable from the top navigation.
- Markdown-rendered documentation content.
- Guidance for writing requirements.
- Guidance for using the Planning Copilot.
- Explanation of phase meanings and dashboard metrics.

## 3. Content Sections
The Help page should include:

- Completeness over estimation.
- What counts as a requirement.
- Examples of strong and weak requirements.
- Task phase definitions.
- Task type definitions.
- Planning Copilot prompt examples.
- Dashboard metric explanation.
- Prototype limitations.
- Existing database/reference-data rules.
- Task and requirement history behavior.

## 4. UX Notes
The Help page should be readable and practical:

- Left-side anchor navigation on desktop.
- Single-column content on smaller screens.
- Markdown typography for headings, lists, and code snippets.
- No marketing copy.

## 5. Suggested Prompt Templates
Include templates such as:

```text
Break down this feature into implementation tasks and binary requirements:
[feature description]
```

```text
Review this task and suggest missing Definition of Done requirements:
[task title and description]
```

```text
Convert this rough idea into tasks grouped by the valid phases from the database:
[idea]
```

## 6. Acceptance Criteria
- Help page is available at `/help`.
- Page explains the completeness model clearly.
- Page includes concrete requirement examples.
- Page includes Planning Copilot prompt examples.
- Help content can be updated without changing application logic.
- Help explains that `task_phase` and `task_type` options come from the existing database and are not edited by the app.
- Help explains that task/requirement updates and deletes preserve previous versions in history tables.
