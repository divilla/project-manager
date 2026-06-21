# Plan 05: Prototype Evaluation & Next Steps

This document details the evaluation guidelines for the **AI-Powered Project Management Tool** prototype, establishing a structured mechanism for testing, logging feedback, and outlining the development steps for the full-scale production system (V2).

---

## 1. Local Testing & Validation Routine
Before sharing the prototype with a broader developer audience or moving to V2, the local system must undergo a 30-day "dogfooding" trial period.

### Core Trial Routine:
1. **Initial Setup:** The developer spins up the local Quasar and Echo servers and creates their primary software development project workspace.
2. **Scoping Sprints:** Every major feature addition or refactoring job must be processed through the AI Planning screen. The developer reviews the LLM's suggested requirements lists and adjusts items as needed.
3. **Daily Tracking:** The developer commits to tracking *all* work progress via the requirement progress engine. No task should be marked as "Completed" on the board unless all its constituent requirement sub-tasks have been verified.
4. **Dashboard Auditing:** The developer monitors the Home Dashboard to identify bottleneck phases using the existing `task_phase` values and assesses if the "Phase Completeness" percentage matches physical product progress.
5. **History Auditing:** The developer spot-checks `task_history` and `requirement_history` after edits and deletes to verify previous versions are captured and deletes are marked with `deleted = true`.

---

## 2. Evaluation Feedback Log (Suggested Template)
During the dogfooding trial, the developer should record observations to guide V2 requirements. A local log (such as `FEEDBACK_LOG.md`) should track:

| Category | Observation / Pain Point | Core Idea for Improvement |
| :--- | :--- | :--- |
| **Requirement Granularity** | *E.g., "Requirements are sometimes too small (e.g. 'write import statement') or too broad ('implement full database')."* | Refine LLM system instructions to enforce a 3-5 item limit per task, with each item representing roughly 1-4 hours of work. |
| **Calculation Accuracy** | *E.g., "If a task has no requirements, it jumps straight from 0% to 100% on phase transition. This causes progress bar jumps."* | Adjust the progress engine behavior in application code without changing phase reference data. |
| **History Completeness** | *E.g., "Requirement toggles are recorded, but task completeness recalculations are not."* | Ensure every task/requirement update path writes the current version to the matching history table inside the same transaction. |
| **AI Generation Speed** | *E.g., "Waiting 8 seconds for OpenAI response on poor connection stalls the planning flow."* | Implement client-side optimistic UI skeletons or migrate prompt execution to local Ollama (Llama3/Mistral) for sub-second offline generation. |

---

## 3. Transition Roadmap to Full-Scale V2
Once the completeness paradigm is validated, the project will expand from a single-user developer utility to a multi-tenant enterprise suite. 

The primary feature updates required for the V2 transition are grouped below:

### Area A: Secure Multi-User Infrastructure
- **User Authentication:** Integrate an industry-standard OAuth2 or JWT provider. Set up registration, login, password recovery, and secure cookie-based session tracking.
- **Task Assignments:** Add user relation tables. Tasks will feature an `assignee_id` field enabling personalized user filter views (e.g., *"Show my tasks"*).
- **Audit Logs:** Extend the existing `task_history` and `requirement_history` model with actor/source metadata if multi-user identity is added later.

### Area B: Developer Ecosystem Integrations
- **Git/GitHub Hooks:** Introduce a webhook endpoint `/api/integrations/github` in Echo. When a developer pushes code containing commit keywords (e.g., `closes #24 [subtask-3]`), the backend automatically checks off the corresponding sub-requirement in the database.
- **CI/CD Triggers:** Connect to local/cloud runner systems (e.g., GitHub Actions, Jenkins). Completion of a unit test runner suite can automatically verify and check off the corresponding "Unit Tests" requirement on active tasks.

### Area C: Advanced Enterprise Features
- **Project Portfolios:** Allow grouping projects into portfolios or organizational units.
- **Multiple Phase Workflows:** Future workflow customization requires explicit human-owned database design. The prototype must continue to use the existing `task_phase` reference data as-is.
- **Data Exporting:** Support exporting project roadmaps and completeness audits as CSV, JSON, or beautifully formatted PDF status summaries for business executives.
