# Plan 04: Phase 3 AI Features

The objective of Phase 3 is to integrate the **AI Planning Copilot**. By the end of this phase, the developer can enter high-level goals on the Planning page, have an LLM automatically decompose them into actionable, phase-grouped tasks and testable requirements, review the result, and write the full plan into the workspace with a single click.

---

## 1. Step-by-Step Developer Implementation Tasks

### Step 3.1: LLM Connector Setup (Backend)
1. In your Go environment variables configuration (e.g. `.env`), register the LLM credentials (such as `OPENAI_API_KEY` or `ANTHROPIC_API_KEY`) and preferred model selection (e.g. `gpt-4o-mini`).
2. Create an `ai` package in Go. Write an initialization function that creates an authorized client targeting the selected provider.
3. Write a helper function `InvokeLLM(systemPrompt, userPrompt string) (string, error)` that wraps standard HTTP request creation, sets appropriate timeouts (e.g., 15 seconds), and handles the API payload dispatch.

### Step 3.2: AI Planning Endpoint (`/api/planning/decompose`)
1. Create a request handler matching `POST /api/planning/decompose`.
2. This handler must:
   - Accept the user's plain-text feature description from the body.
   - Inject the prompt structure defined in `specs/06-ai-integration-strategy.md` containing strict formatting guidelines.
   - Dispatch the assembled prompt package to the LLM via `InvokeLLM`.
   - Parse the returning payload, extracting the structured JSON structure.
   - Return the validated list of proposed tasks and requirements arrays back to the client.

### Step 3.3: Planning Copilot Interface (Frontend)
1. Build out `PlanningPage.vue` as a rich, conversational workspace.
2. Structure the screen into two vertical halves:
   - **Left Half (Input):** A large `<q-input type="textarea">` with a clear placeholder prompt (e.g. *"Explain the feature or module you are looking to build..."*). Below the input, place a "Generate Project Blueprint" `<q-btn>` featuring a loading spinner.
   - **Right Half (Results Panel):** An empty state card explaining how the planner operates. Once a response is received, render the LLM suggestions dynamically using a structured collapsible list grouped by the standard phases.
3. For each suggested task, render:
   - The task title.
   - An editable description.
   - A nested requirements list of the proposed sub-requirements with checkboxes to allow the developer to prune, rename, or customize the suggestions before committing.

### Step 3.4: One-Click Commit Action
1. At the bottom of the suggestions list on the Planning screen, build a prominent "Commit Blueprint to Project Board" button.
2. Clicking this button sends a batch `POST` request to `/api/projects/:id/batch-tasks` containing the selected, edited task array.
3. On the Go backend, process this batch operation within a secure SQL transaction:
   - Insert each task row mapping to the specified phase.
   - Read the generated task IDs.
   - Insert all nested requirement rows pointing back to their newly created parent tasks.
4. Upon receiving a success response, redirect the user's viewport automatically to the Projects board (`/projects`) to begin development.

---

## 2. Success Criteria & Verification Checklist

To complete Phase 3, verify the following checks pass:

- [ ] **AI Endpoint Testing:** Triggering `POST /api/planning/decompose` via an API testing tool (with a body like `{"prompt": "Build an email notification system"}`) successfully returns a parseable JSON array of suggested tasks.
- [ ] **Interactive Loader:** The Quasar frontend displays a highly visible loading state while waiting for the LLM to process and return response payloads.
- [ ] **Aesthetic Layout:** The generated tasks on the Planning screen are displayed logically and separated visually by phase, matching the styling constraints defined in `specs/05-user-interface-flows.md`.
- [ ] **Customization:** The developer can check/uncheck, rename, or edit task descriptions before committing the plan to the active project workspace.
- [ ] **Transactional Writing:** Clicking the "Commit" action successfully performs a bulk SQL operation, creating all tasks and requirements simultaneously. Checking your PostgreSQL database verifies that task-foreign-keys map correctly.
- [ ] **Graceful Failures:** Simulating an LLM timeout or cutting the internet connection displays a clean, user-friendly toast notification in Quasar without throwing console crashes.