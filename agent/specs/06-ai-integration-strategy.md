# Spec 06: AI Integration Strategy

The core differentiator of this project management tool is its server-side integration with Large Language Models (LLMs) to automate task lifecycle planning, definition of done (DoD) breakdown, and project-health analysis.

---

## 1. LLM Integration Pipeline
The backend Go service acts as a proxy controller between the user interface and the LLM endpoint (e.g. OpenAI API or Anthropic).

```
 +-------------+    1. User Prompt     +-------------+    2. Inject Context     +-------------+
 |  Vue UI     | --------------------> |  Go Backend | -----------------------> | LLM Provider|
 | (Dashboard) | <-------------------- | (Echo App)  | <----------------------- | (JSON Res)  |
 +-------------+  4. Render Require-   +-------------+     3. Parse Structured  +-------------+
                    ments                                    JSON Payload
```

---

## 2. Core LLM Operations

### Operation A: High-Level Goal Decomposition (Planning Copilot)
- **Objective:** Take a user's natural language project objective (e.g., *"Set up Docker deployment for an Express API and Nginx proxy"*) and transform it into a structured set of tasks grouped by standard development phases, where each task features a nested "Definition of Done" requirements list.
- **Prompt Engineering Strategy (System Blueprint):**
  - Instruct the model to act as a World-Class Software Architect and Agile Product Manager.
  - Define the target development phases explicitly: `backlog`, `planning`, `in_progress`, `review`, and `completed`.
  - Enforce a structured output format (such as JSON Schema or standard XML tags) containing:
    1. Task Title.
    2. Task Description.
    3. Suggested Phase.
    4. Requirements: An array of 3-5 concrete, binary-verifiable sub-requirements (e.g., *"Create Dockerfile using a multi-stage node:alpine build"* instead of *"Write Dockerfile"*).

### Operation B: Completeness Review & Verification (Optional Enhancement)
- **Objective:** Provide automated feedback on task completeness.
- **Workflow:** The user describes what they have implemented (or paste a git diff/log summary) for a task. The LLM reviews the text description against the original requirements.
- **Output:** The LLM returns an advisory assessment detailing which requirements look fully satisfied, which ones look partially complete, and suggestions for what is missing.

---

## 3. High-Level Prompt Blueprint (Task Decomposition)
Below is the conceptual prompt scaffolding compiled by the Go backend before dispatching to the LLM:

```yaml
SystemPrompt: |
  You are an expert technical product manager. Your task is to decompose high-level development goals into a structured, phase-based backlog.
  
  You must partition tasks into these specific phases:
  - backlog: Raw ideas, future extensions.
  - planning: Architecture design, DB schema draft, API contract agreement.
  - in_progress: Implementation, logic writing, initial unit testing.
  - review: Code review, QA, manual validation, integration tests.

  Every task must contain a detailed "Definition of Done" requirements list. Each single DoD item is defined as a 'requirement'. Each requirement must be concrete, binary (either true or false), and testable. Avoid vague descriptors like "make sure it works." Use explicit assertions like "write test suite achieving 80% coverage."

  Respond exclusively in a clean structured format representing:
  {
    "tasks": [
      {
        "title": "Task title",
        "description": "Task description",
        "phase": "suggested_phase_string",
        "requirements": ["verifiable requirement 1", "verifiable requirement 2"]
      }
    ]
  }

UserPrompt: |
  Break down the following feature for a web application:
  "${USER_GOAL_INPUT}"
```

---

## 4. Latency, Resiliency, & Fallbacks

### Managing Latency
LLM API calls can take between 2 to 10 seconds. To maintain a highly responsive UI, the system must:
1. Trigger immediate user-interface loading indicators (`q-spinner`) with motivating loading tips (e.g. *"AI is mapping your database tables..."*).
2. Utilize low-latency models for task breakdown (e.g., `gpt-4o-mini` or Claude `haiku`) as they provide an excellent balance of speed and structured JSON execution capability.

### Error Recovery & Graceful Degradation
- **JSON Parsing Failures:** LLMs occasionally append conversational preambles even when instructed to output strict JSON. The Go backend must use regex-based extractions (e.g. searching for JSON brackets `{...}`) or native provider tool-calling parameters to ensure parsing stability.
- **Fallback Workflows:** If the API times out or returns an error (e.g. quota limit reached), the Go backend must return a graceful fallback status. The frontend UI will display a helpful error card: *"We are having trouble reaching our AI assistant. However, you can manually create your project and add tasks here"*—preventing the application from breaking.